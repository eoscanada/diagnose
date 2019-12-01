package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

	eosd "github.com/eoscanada/diagnose/eos"
	ethd "github.com/eoscanada/diagnose/eth"

	"github.com/eoscanada/diagnose/renderer"
	"github.com/eoscanada/dstore"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/kvdb/eosdb"
	"github.com/eoscanada/kvdb/ethdb"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type RootServer struct {
	Addr   string
	routes *mux.Router

	Api *eos.API

	Namespace             string
	Protocol              string
	BlockStoreUrl         string
	SearchIndexesStoreUrl string
	SearchShardSize       uint32
	KvdbConnection        string

	BlocksStore dstore.Store
	SearchStore dstore.Store

	EOSdb *eosdb.EOSDatabase
	ETHdb *ethdb.ETHDatabase

	Cluster *kubernetes.Clientset

	ParallelDownloadCount uint64
}

func (d *RootServer) SetupRoutes() {
	configureValidators()

	router := mux.NewRouter()
	router.Path("/").Methods("GET").HandlerFunc(d.index)
	router.Path("/dfuse.css").Methods("GET").HandlerFunc(d.css)

	switch d.Protocol {
	case "EOS":
		eosDiagnose := &eosd.EOSDiagnose{
			Namespace:             d.Namespace,
			BlocksStoreUrl:        d.BlockStoreUrl,
			SearchIndexesStoreUrl: d.SearchIndexesStoreUrl,
			SearchShardSize:       d.SearchShardSize,
			KvdbConnection:        d.KvdbConnection,
			BlocksStore:           d.BlocksStore,
			SearchStore:           d.SearchStore,
			EOSdb:                 d.EOSdb,
			ParallelDownloadCount: d.ParallelDownloadCount,
		}
		s := router.PathPrefix("/v1/EOS").Subrouter()
		eosDiagnose.SetupRoutes(s)

	case "ETH":
		ethDiagnose := &ethd.ETHDiagnose{
			Namespace:             d.Namespace,
			BlocksStoreUrl:        d.BlockStoreUrl,
			SearchIndexesStoreUrl: d.SearchIndexesStoreUrl,
			SearchShardSize:       d.SearchShardSize,
			KvdbConnection:        d.KvdbConnection,
			BlocksStore:           d.BlocksStore,
			SearchStore:           d.SearchStore,
			ETHdb:                 d.ETHdb,
		}
		s := router.PathPrefix("/v1/ETH").Subrouter()
		ethDiagnose.SetupRoutes(s)

	}

	router.Path("/v1/diagnose/services_health_checks").Methods("GET").HandlerFunc(d.getServicesHealthChecks)
	router.Path("/v1/diagnose/").Methods("POST").Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	d.routes = router
}

func (d *RootServer) getServicesHealthChecks(w http.ResponseWriter, r *http.Request) {
	renderer.PutPreambule(w, "Services health checks")

	services, err := d.Cluster.CoreV1().Services(d.Namespace).List(meta_v1.ListOptions{})
	if err != nil {
		renderer.PutLine(w, "<pre>Failed listing services: %s</pre>", err)
		return
	}

	for _, svc := range services.Items {
		endpoints, err := d.Cluster.CoreV1().Endpoints(d.Namespace).Get(svc.Name, meta_v1.GetOptions{})
		if err != nil {
			renderer.PutLine(w, "<pre>failed getting service %q: %s</pre>", svc.Name, err)
			continue
		}

		renderer.PutLine(w, "<h4>Service: %q</h4>", svc.Name)

		for _, subset := range endpoints.Subsets {
			for _, addr := range subset.Addresses {
				for _, port := range subset.Ports {
					theURL := fmt.Sprintf("http://%s:%d/healthz?secret=dfuse&full=true", addr.IP, port.Port)
					renderer.PutLine(w, "<pre>Querying %s\n", theURL)
					// Query the health endpoint
					resp, err := http.Get(theURL)
					if err != nil {
						renderer.PutLine(w, "Failed: %s\n", err)
					} else {
						cnt, _ := ioutil.ReadAll(resp.Body)
						renderer.PutLine(w, "Status: %d\n\n", resp.StatusCode)
						renderer.PutLine(w, string(cnt))
						renderer.PutLine(w, "\n")
					}
					renderer.PutLine(w, "</pre>")
				}
			}
		}
	}
}

func (d *RootServer) index(w http.ResponseWriter, r *http.Request) {
	data := &tplData{
		Router:             d.routes,
		Protocol:           d.Protocol,
		Namespace:          d.Namespace,
		BlockStore:         d.BlockStoreUrl,
		SearchIndexesStore: d.SearchIndexesStoreUrl,
		KVDBConnection:     d.KvdbConnection,
	}

	d.renderTemplate(w, data)
}

func (d *RootServer) css(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "dfuse.css")
}

func (d *RootServer) Serve() error {
	zlog.Info("Serving on address", zap.String("Addr", d.Addr))
	return http.ListenAndServe(d.Addr, d.routes)
}
