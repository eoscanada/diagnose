package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/eoscanada/dmesh"
	"github.com/eoscanada/dstore"
	"github.com/eoscanada/kvdb/eosdb"
	"github.com/eoscanada/kvdb/ethdb"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"go.etcd.io/etcd/clientv3"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
)

type Diagnose struct {
	addr string

	Protocol              string `json:"protocol,omitempty"`
	Namespace             string `json:"namespace,omitempty"`
	BlocksStoreUrl        string `json:"blockStoreUrl,omitempty"`
	SearchIndexesStoreUrl string `json:"indexesStoreUrl,omitempty"`
	SearchShardSize       uint32 `json:"shardSize,omitempty"`
	KvdbConnectionInfo    string `json:"kvdbConnectionInfo,omitempty"`
	DmeshServiceVersion   string `json:"dmeshServiceVersion,omitempty"`

	BlocksStore dstore.Store
	SearchStore dstore.Store

	EOSdb *eosdb.EOSDatabase
	ETHdb *ethdb.ETHDatabase

	router     *mux.Router
	upgrader   *websocket.Upgrader
	cluster    *kubernetes.Clientset
	dmeshStore *clientv3.Client
}

func (d *Diagnose) SetupRoutes() {
	configureValidators()

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	d.upgrader = &upgrader

	router := mux.NewRouter()
	router.Path("/").Methods("GET").HandlerFunc(d.index)

	apiRouter := router.PathPrefix("/api").Subrouter()

	// basic diagnose path
	//apiRouter.Path("/v1/services_health_checks").Methods("GET").HandlerFunc(d.getServicesHealthChecks)
	apiRouter.Path("/diagnose/").Methods("POST").Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	apiRouter.Path("/config").Methods("Get").HandlerFunc(d.config)
	apiRouter.Path("/search_peers").Methods("Get").HandlerFunc(d.searchPeers)
	apiRouter.Path("/test").Methods("Get").HandlerFunc(d.test)
	apiRouter.Path("/block_holes").Methods("GET").HandlerFunc(d.BlockHoles)
	apiRouter.Path("/search_holes").Methods("GET").HandlerFunc(d.SearchHoles)
	apiRouter.Path("/db_holes").Methods("GET").HandlerFunc(d.DBHoles)

	//s.Path("/trx_problems").Methods("GET").HandlerFunc(e.trxProblems)
	//s.Path("/db_holes").Methods("GET").HandlerFunc(e.DBHoles)
	//s.Path("/verify_stats").Methods("GET").HandlerFunc(e.verifyStats)
	//s.Path("/verify_stats_top_accounts").Methods("GET").HandlerFunc(e.verifyStatsTopAccounts)

	// protocol paths
	//d.diagnose.SetUpgrader(d.upgrader)
	//d.diagnose.SetupRoutes(apiRouter)

	d.router = router

}

func (r *Diagnose) test(w http.ResponseWriter, req *http.Request) {
	fmt.Printf("--------------------------- test\n")
	conn, err := r.upgrader.Upgrade(w, req, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(req.Context())

	go websocketCloser(conn, cancel)

	for i := 0; i < 10; i++ {
		select {
		case <-ctx.Done():
			fmt.Printf("main loop context canceled\n")
			return
		default:
			msg := fmt.Sprintf("hello world: %d", i)
			if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
				fmt.Printf("ERRROR\n")
				return
			}
			fmt.Printf("Sent: %s\n", msg)
			time.Sleep(2 * time.Second)
		}
	}
}

func (r *Diagnose) searchPeers(w http.ResponseWriter, req *http.Request) {

	conn, err := r.upgrader.Upgrade(w, req, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	ctx := req.Context()
	servicePrefix := fmt.Sprintf("%s/search", r.DmeshServiceVersion)
	eventChan := dmesh.Observe(ctx, r.dmeshStore, r.Namespace, servicePrefix)
	for {
		select {
		case <-ctx.Done():
			break
		case peer := <-eventChan:
			fmt.Printf("received peer info\n")

			data, err := json.Marshal(peer)
			if err != nil {
				return
			}
			fmt.Printf("peer marshall: %s\n", data)
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
		}
	}
}

func (r *Diagnose) config(w http.ResponseWriter, req *http.Request) {
	json.NewEncoder(w).Encode(r)
}

//func (r *Diagnose) getServicesHealthChecks(w http.ResponseWriter, req *http.Request) {
//	renderer.PutPreambule(w, "Services health checks")
//
//	services, err := r.cluster.CoreV1().Services(r.Namespace).List(meta_v1.ListOptions{})
//	if err != nil {
//		renderer.PutLine(w, "<pre>Failed listing services: %s</pre>", err)
//		return
//	}
//
//	for _, svc := range services.Items {
//		endpoints, err := r.cluster.CoreV1().Endpoints(r.Namespace).Get(svc.Name, meta_v1.GetOptions{})
//		if err != nil {
//			renderer.PutLine(w, "<pre>failed getting service %q: %s</pre>", svc.Name, err)
//			continue
//		}
//
//		renderer.PutLine(w, "<h4>Service: %q</h4>", svc.Name)
//
//		for _, subset := range endpoints.Subsets {
//			for _, addr := range subset.Addresses {
//				for _, port := range subset.Ports {
//					theURL := fmt.Sprintf("http://%s:%d/healthz?secret=dfuse&full=true", addr.IP, port.Port)
//					renderer.PutLine(w, "<pre>Querying %s\n", theURL)
//					// Query the health endpoint
//					resp, err := http.Get(theURL)
//					if err != nil {
//						renderer.PutLine(w, "Failed: %s\n", err)
//					} else {
//						cnt, _ := ioutil.ReadAll(resp.Body)
//						renderer.PutLine(w, "Status: %d\n\n", resp.StatusCode)
//						renderer.PutLine(w, string(cnt))
//						renderer.PutLine(w, "\n")
//					}
//					renderer.PutLine(w, "</pre>")
//				}
//			}
//		}
//	}
//}

func (r *Diagnose) index(w http.ResponseWriter, req *http.Request) {
	//data := &tplData{
	//	Router:             r.router,
	//	Protocol:           r.Protocol,
	//	Namespace:          r.Namespace,
	//	BlockStore:         r.diagnose.GetBlockStoreUrl(),
	//	SearchIndexesStore: r.diagnose.GetSearchIndexesStoreUrl(),
	//	KVDBConnection:     r.diagnose.GetKvdbConnectionInfo(),
	//}
	//
	//r.renderTemplate(w, data)
}

func (r *Diagnose) Serve() error {

	// http
	httpListener, err := net.Listen("tcp", r.addr)
	if err != nil {
		return fmt.Errorf("http listen failed: %s", r.addr)
	}

	corsMiddleware := NewCORSMiddleware()
	httpServer := http.Server{
		Handler: corsMiddleware(r.router),
	}

	zlog.Info("serving HTTP", zap.String("http_addr", r.addr))
	err = httpServer.Serve(httpListener)
	if err != nil {
		zlog.Panic("unable to start HTTP server", zap.String("http_addr", r.addr), zap.Error(err))
	}

	zlog.Info("completed HTTP server")
	return nil
}

func NewCORSMiddleware() mux.MiddlewareFunc {
	allowedHeaders := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization", "X-Eos-Push-Guarantee"})
	allowedOrigins := handlers.AllowedOrigins([]string{"*"})
	allowedMethods := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "OPTIONS"})
	maxAge := handlers.MaxAge(86400) // 24 hours - hard capped by Firefox / Chrome is max 10 minutes

	return handlers.CORS(allowedHeaders, allowedOrigins, allowedMethods, maxAge)
}
