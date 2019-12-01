package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"

	"github.com/eoscanada/diagnose/renderer"
	"github.com/eoscanada/dmesh"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"go.etcd.io/etcd/clientv3"
	"go.uber.org/zap"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type KVDBdb interface {
}
type RootServer struct {
	addr string

	Protocol              string `json:"protocol,omitempty"`
	Namespace             string `json:"namespace,omitempty"`
	BlocksStoreUrl        string `json:"blockStoreUrl,omitempty"`
	SearchIndexesStoreUrl string `json:"indexesStoreUrl,omitempty"`
	SearchShardSize       uint32 `json:"shardSize,omitempty"`
	KvdbConnectionInfo    string `json:"kvdbConnectionInfo,omitempty"`
	DmeshServiceVersion   string `json:"dmeshServiceVersion,omitempty"`

	diagnose   Diagnose
	router     *mux.Router
	upgrader   *websocket.Upgrader
	cluster    *kubernetes.Clientset
	dmeshStore *clientv3.Client
}

func NewRootServer(addr, protocol, namespace, blocksStoreUrl, searchIndexesStoreUrl, kvdbConnectionInfo string, searchShardSize uint32, dmeshServiceVersion string, diagnose Diagnose, cluster *kubernetes.Clientset, dmeshStore *clientv3.Client) *RootServer {
	return &RootServer{
		addr:                  addr,
		Protocol:              protocol,
		BlocksStoreUrl:        blocksStoreUrl,
		SearchIndexesStoreUrl: searchIndexesStoreUrl,
		SearchShardSize:       searchShardSize,
		KvdbConnectionInfo:    kvdbConnectionInfo,
		DmeshServiceVersion:   dmeshServiceVersion,
		Namespace:             namespace,
		diagnose:              diagnose,
		cluster:               cluster,
		dmeshStore:            dmeshStore,
	}
}

func (r *RootServer) SetupRoutes() {
	configureValidators()

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	r.upgrader = &upgrader

	router := mux.NewRouter()
	router.Path("/").Methods("GET").HandlerFunc(r.index)
	router.Path("/dfuse.css").Methods("GET").HandlerFunc(r.css)

	apiRouter := router.PathPrefix("/api").Subrouter()

	// basic diagnose path
	apiRouter.Path("/v1/services_health_checks").Methods("GET").HandlerFunc(r.getServicesHealthChecks)
	apiRouter.Path("/v1/diagnose/").Methods("POST").Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	apiRouter.Path("/v1/config").Methods("Get").HandlerFunc(r.config)
	apiRouter.Path("/v1/search_peers").Methods("Get").HandlerFunc(r.searchPeers)

	// protocol paths
	r.diagnose.SetUpgrader(r.upgrader)
	r.diagnose.SetupRoutes(apiRouter)

	r.router = router

}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

func (r *RootServer) searchPeers(w http.ResponseWriter, req *http.Request) {

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

func (r *RootServer) config(w http.ResponseWriter, req *http.Request) {
	json.NewEncoder(w).Encode(r)
}

func (r *RootServer) getServicesHealthChecks(w http.ResponseWriter, req *http.Request) {
	renderer.PutPreambule(w, "Services health checks")

	services, err := r.cluster.CoreV1().Services(r.Namespace).List(meta_v1.ListOptions{})
	if err != nil {
		renderer.PutLine(w, "<pre>Failed listing services: %s</pre>", err)
		return
	}

	for _, svc := range services.Items {
		endpoints, err := r.cluster.CoreV1().Endpoints(r.Namespace).Get(svc.Name, meta_v1.GetOptions{})
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

func (r *RootServer) index(w http.ResponseWriter, req *http.Request) {
	data := &tplData{
		Router:             r.router,
		Protocol:           r.Protocol,
		Namespace:          r.Namespace,
		BlockStore:         r.diagnose.GetBlockStoreUrl(),
		SearchIndexesStore: r.diagnose.GetSearchIndexesStoreUrl(),
		KVDBConnection:     r.diagnose.GetKvdbConnectionInfo(),
	}

	r.renderTemplate(w, data)
}

func (d *RootServer) css(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "dfuse.css")
}

func (r *RootServer) Serve() error {

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
