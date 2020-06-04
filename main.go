package main

import (
	"flag"
	"fmt"

	"github.com/eoscanada/derr"
	"github.com/eoscanada/dmesh"
	"github.com/eoscanada/kvdb"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var flagHTTPListenAddr = flag.String("listen-http-addr", ":8080", "TCP Listener addr for http")
var flagProtocol = flag.String("protocol", "ETH", "Protocol to load, EOS or ETH")
var flagNamespace = flag.String("namespace", "eth-ropsten", "k8s namespace inspected by this diagnose instance")
var flagBlocksStoreURL = flag.String("blocks-store", "gs://dfuseio-global-blocks-us/eth-ropsten/v2", "Blocks logs storage location")
var flagSearchIndexesStoreURL = flag.String("search-indexes-store", "gs://dfuseio-global-indices-us/eth-ropsten/v2", "GS location of search indexes storage for EOS")
var flagSearchShardSize = flag.Uint("search-shard-size", 200, "Number of blocks to store in a given Bleve index")
var flagBigTable = flag.String("db-connection", "dfuseio-global:dfuse-saas:ropsten-v2", "Big table connection string as 'project:instance:table-prefix'")
var flagAPIURL = flag.String("api-url", "https://ropsten.eth.dfuse.io", "The API node to reach for information about the chain")
var flagSkipK8S = flag.Bool("skip-k8s", false, "Useful in development to avoid setuping access to a K8S cluster")
var flagDev = flag.Bool("dev", false, "Useful in development to link to localhost:3000 instead of needing full react build")
var flagMeshStoreAddr = flag.String("mesh-store-addr", ":2379", "address of the backing etcd cluster for mesh service discovery")
var flagMeshServiceVersion = flag.String("mesh-service-version", "v1", "service version within dmesh")
var flagServeFilePath = flag.String("serve-file-path", "./frontend/public", "path to files to serve under `/`")

func main() {
	flag.Parse()
	setupLogger()

	zlog.Info("checking up kvdb info")
	_, err := kvdb.NewConnectionInfo(*flagBigTable)
	derr.Check(fmt.Sprintf("unable to parse kvdb connection info %s", *flagBigTable), err)

	//initalise dmesh client
	dmeshStore, err := dmesh.NewStore(*flagMeshStoreAddr)
	derr.Check("unable to setup dmesh store (etcd)", err)
	defer dmeshStore.Close()

	performK8sSetup := !*flagSkipK8S
	var cluster *kubernetes.Clientset
	if performK8sSetup {
		zlog.Info("setting up k8s clientset")
		config, err := rest.InClusterConfig()
		derr.Check("unable to retrieve kubernetes cluster config", err)

		cluster, err = kubernetes.NewForConfig(config)
		derr.Check("unable to create kubernetes client set", err)
	}

	diagnose := Diagnose{
		addr:                  *flagHTTPListenAddr,
		Protocol:              *flagProtocol,
		Namespace:             *flagNamespace,
		BlocksStoreURL:        *flagBlocksStoreURL,
		SearchIndexesStoreURL: *flagSearchIndexesStoreURL,
		SearchShardSize:       uint32(*flagSearchShardSize),
		SearchShardSizes:      []uint32{50, 200, 500, 1000, 5000, 10000, 50000},
		KvdbConnectionInfo:    *flagBigTable,
		DmeshServiceVersion:   *flagMeshServiceVersion,
		cluster:               cluster,
		dmeshStore:            dmeshStore,
		serveFilePath:         *flagServeFilePath,
	}

	diagnose.SetupRoutes(*flagDev)

	zlog.Info("serving http")
	err = diagnose.Serve()
	derr.Check("failed serving http", err)
}
