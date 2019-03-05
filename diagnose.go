package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/eoscanada/bstream/store"
	"github.com/eoscanada/eosdb"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Diagnose struct {
	addr   string
	routes *mux.Router

	namespace string
	eosdb     eosdb.DBReader

	blocksStore store.ArchiveStore
	searchStore *store.SimpleGStore

	cluster *kubernetes.Clientset
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

	putLine(w, "<html><head><title>Checking holes in EOSDB</title></head><h1>Checking holes in EOSDB</h1>")

	// TODO: navigate all of Bigtable blocks, and make sure there
	// are no holes, otherwise print in the logs that there are
	// some yucky schtuff.
}

func (d *Diagnose) getServicesHealthChecks(w http.ResponseWriter, r *http.Request) {
	putLine(w, "<html><head><title>Services health checks</title></head><h1>All services health checks</h1>")

	services, err := d.cluster.CoreV1().Services(d.namespace).List(meta_v1.ListOptions{})
	if err != nil {
		putLine(w, "<pre>Failed listing services: %s</pre>", err)
		return
	}

	for _, svc := range services.Items {
		endpoints, err := d.cluster.CoreV1().Endpoints(d.namespace).Get(svc.Name, meta_v1.GetOptions{})
		if err != nil {
			putLine(w, "<pre>failed getting service %q: %s</pre>", svc.Name, err)
			continue
		}

		putLine(w, "<h4>Service: %q</h4>", svc.Name)

		for _, subset := range endpoints.Subsets {
			for _, addr := range subset.Addresses {
				for _, port := range subset.Ports {
					theURL := fmt.Sprintf("http://%s:%d/healthz", addr.IP, port.Port)
					putLine(w, "<pre>Querying %s\n", theURL)
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

func (d *Diagnose) index(w http.ResponseWriter, r *http.Request) {
	// TODO: fetch in-cluster schtuff..

	data := &tplData{}

	d.renderTemplate(w, data)
}

func (d *Diagnose) Serve() error {
	zlog.Info("Serving on address", zap.String("addr", d.addr))
	return http.ListenAndServe(d.addr, d.routes)
}

func putLine(w http.ResponseWriter, format string, v ...interface{}) {
	flush := w.(http.Flusher)
	fmt.Fprintf(w, format, v...)
	flush.Flush()
}
