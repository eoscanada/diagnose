package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/eoscanada/eosdb"
	"github.com/gorilla/mux"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Diagnose struct {
	addr   string
	routes *mux.Router

	healthServices []string
	namespace      string
	eosdb          eosdb.DBReader

	cluster *kubernetes.Clientset
}

func (d *Diagnose) setupK8s() error {
	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	d.cluster = clientset

	return nil
}

func (d *Diagnose) setupRoutes() {
	r := mux.NewRouter()
	r.Path("/").Methods("GET").HandlerFunc(d.index)
	r.Path("/v1/diagnose/verify_eosdb_holes").Methods("GET").HandlerFunc(d.verifyEOSDBHoles)
	r.Path("/v1/diagnose/services_health_checks").Methods("GET").HandlerFunc(d.getServicesHealthChecks)
	r.Path("/v1/diagnose/").Methods("POST").Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	}))

	d.routes = r
}

func (d *Diagnose) verifyEOSDBHoles(w http.ResponseWriter, r *http.Request) {
	// TODO: receive parameters to limit search..

	flush := w.(http.Flusher)

	fmt.Fprintf(w, "<html><head><title>Checking holes in EOSDB</title></head><h1>Checking holes in EOSDB</h1>")
	flush.Flush()

	// TODO: navigate all of Bigtable blocks, and make sure there
	// are no holes, otherwise print in the logs that there are
	// some yucky schtuff.
}

func (d *Diagnose) getServicesHealthChecks(w http.ResponseWriter, r *http.Request) {
	// TODO: call the health check on all listed services' endpoints
	// use `d.cluster` to get the Services, get the corresponding `Endpoints`.
	// Depending on the name, poke a certain endpoint and pack the result, and return
	// to the user.

	putLine(w, "<html><head><title>Services health checks</title></head><h1>All services health checks</h1>")

	for _, serviceName := range d.healthServices {
		endpoints, err := d.cluster.CoreV1().Endpoints(d.namespace).Get(serviceName, meta_v1.GetOptions{})
		if err != nil {
			putLine(w, "<pre>failed getting service %q: %s</pre>", serviceName, err)
			continue
		}

		for _, subset := range endpoints.Subsets {
			for _, addr := range subset.Addresses {
				for _, port := range subset.Ports {
					theURL := fmt.Sprintf("http://%s:%d/healthz", addr.IP, port.Port)
					putLine(w, "<pre>Service %q. Querying %s\n", serviceName, theURL)
					// Query the health endpoint
					resp, err := http.Get(theURL)
					if err != nil {
						putLine(w, "Failed: %s\n", err)
					} else {
						cnt, _ := ioutil.ReadAll(resp.Body)
						putLine(w, "Status: %d\n\n", resp.StatusCode)
						putLine(w, string(cnt))
						putLine(w, "\n")
					}
					putLine(w, "</pre>")
				}
			}
		}
	}
}

func putLine(w http.ResponseWriter, format string, v ...interface{}) {
	flush := w.(http.Flusher)
	fmt.Fprintf(w, format, v...)
	flush.Flush()
}

func (d *Diagnose) index(w http.ResponseWriter, r *http.Request) {
	// TODO: fetch in-cluster schtuff..

	data := &tplData{}

	d.renderTemplate(w, data)
}

func (d *Diagnose) Serve() error {
	return http.ListenAndServe(d.addr, d.routes)
}
