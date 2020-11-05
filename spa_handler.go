package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"regexp"

	"github.com/koding/websocketproxy"
	"go.uber.org/zap"
)

// SPAHandler SPAHandler config
type SPAHandler struct {
	htmlRoot         string
	staticPathRegexp *regexp.Regexp
	devMode          bool
	reverseProxy     *httputil.ReverseProxy
	websocketProxy   *websocketproxy.WebsocketProxy
}

func NewSPAHandler(htmlRoot string, devMode bool) *SPAHandler {
	zlog.Info("setting single page appliactio handler",
		zap.String("html_root", htmlRoot),
		zap.Bool("dev_mode", devMode))

	host := "localhost:3000"
	u, err := url.Parse(fmt.Sprintf("http://%s/", host))
	if err != nil {
		zlog.Fatal("parsing url", zap.Error(err), zap.String("host", host))
	}
	reverseProxy := httputil.NewSingleHostReverseProxy(u)

	wsURL, _ := url.Parse(fmt.Sprintf("ws://%s/", host))
	wsProxy := websocketproxy.NewProxy(wsURL)

	return &SPAHandler{
		htmlRoot:         filepath.Clean(htmlRoot),
		staticPathRegexp: regexp.MustCompile("^/static/"),
		devMode:          devMode,
		reverseProxy:     reverseProxy,
		websocketProxy:   wsProxy,
	}
}

func (p *SPAHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if p.devMode {
		p.ServeHTTPForDevelopment(w, r)
	} else {
		p.HandleHTTPSRedirect(w, r)
		p.ServeHTTPForProduction(w, r)
	}
}

// ServeHTTPForDevelopment Proxies the locally running development app server
func (p *SPAHandler) ServeHTTPForDevelopment(w http.ResponseWriter, r *http.Request) {
	r.Header.Del("Accept-Encoding")
	if r.Header.Get("Connection") == "Upgrade" {
		zlog.Debug("proxying websocket connection (upgrade)", zap.String("path", r.URL.Path))
		p.websocketProxy.ServeHTTP(w, r)
	} else {
		zlog.Debug("proxying", zap.String("path", r.URL.Path))
		p.reverseProxy.ServeHTTP(w, r)
	}
}

// HandleHTTPSRedirect Redirects http to https, and adds the proper headers top HTTPS requests
func (p *SPAHandler) HandleHTTPSRedirect(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-Forwarded-Proto") == "http" {
		// Redirect http to https
		target := "https://" + r.Host + r.URL.Path
		if len(r.URL.RawQuery) > 0 {
			target += "?" + r.URL.RawQuery
		}
		http.Redirect(w, r, target, http.StatusMovedPermanently)
	} else {
		w.Header().Add("Strict-Transport-Security", "max-age=600; includeSubDomains; preload")
	}
}

// ServeHTTPForProduction Serves the production app
func (p *SPAHandler) ServeHTTPForProduction(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if path == "/" {
		path = "/index.html"
	}

	var fileFound bool
	filePath := filepath.Join(p.htmlRoot, path)
	_, err := os.Stat(filePath)
	if err == nil {
		fileFound = true
	}
	zlog.Info("http prod serve", zap.String("file_path", path), zap.Bool("file_found", fileFound))
	if fileFound && path != "/index.html" {
		// If file exists, serve that file
		http.FileServer(http.Dir(p.htmlRoot)).ServeHTTP(w, r)
	} else if p.staticPathRegexp.MatchString(path) {
		// 404 Error
		http.Error(w, "resource not found", http.StatusNotFound)

	} else {
		// For any other request, bust cache and serve index.html
		bustCache(w)
		http.ServeFile(w, r, filepath.Join(p.htmlRoot, "/index.html"))
	}
}

func bustCache(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
}
