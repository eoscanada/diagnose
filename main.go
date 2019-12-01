package main

import (
	"flag"
	"fmt"

	"github.com/eoscanada/derr"
	"github.com/eoscanada/diagnose/eos"
	"github.com/eoscanada/diagnose/eth"
	"github.com/eoscanada/dmesh"
	"github.com/eoscanada/dstore"
	"github.com/eoscanada/kvdb"
	"github.com/eoscanada/kvdb/eosdb"
	"github.com/eoscanada/kvdb/ethdb"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var flagHttpListenAddr = flag.String("listen-http-addr", ":8080", "TCP Listener addr for http")
var flagProtocol = flag.String("protocol", "ETH", "Protocol to load, EOS or ETH")
var flagNamespace = flag.String("namespace", "eth-ropsten", "k8s namespace inspected by this diagnose instance")
var flagBlocksStore = flag.String("blocks-store", "gs://dfuseio-global-blocks-us/eth-ropsten/v2", "Blocks logs storage location")
var flagSearchIndexesStore = flag.String("search-indexes-store", "gs://dfuseio-global-indices-us/eth-ropsten/v2", "GS location of search indexes storage for EOS")
var flagSearchShardSize = flag.Uint("search-shard-size", 200, "Number of blocks to store in a given Bleve index")
var flagBigTable = flag.String("db-connection", "dfuseio-global:dfuse-saas:ropsten-v2", "Big table connection string as 'project:instance:table-prefix'")
var flagAPIURL = flag.String("api-url", "https://ropsten.eth.dfuse.io", "The API node to reach for information about the chain")
var flagParallelDownloadCount = flag.Uint64("parallel-download-count", 6, "How many blocks file to download in parallel")
var flagSkipK8S = flag.Bool("skip-k8s", false, "Useful in development to avoid setuping access to a K8S cluster")
var flagMeshStoreAddr = flag.String("mesh-store-addr", ":2379", "address of the backing etcd cluster for mesh service discovery")
var flagMeshServiceVersion = flag.String("mesh-service-version", "v1", "service version within dmesh")

func main() {
	flag.Parse()
	setupLogger()

	zlog.Info("setting up search indexes stores")
	searchStore, err := dstore.NewSimpleStore(*flagSearchIndexesStore)
	derr.Check("unable to setup search index store", err)

	zlog.Info("setting up search block stores")
	blocksStore, err := dstore.NewDBinStore(*flagBlocksStore)
	derr.Check("unable to setup block store store", err)

	zlog.Info("setting up kvdb")
	kvdbInfo, err := kvdb.NewConnectionInfo(*flagBigTable)
	derr.Check(fmt.Sprintf("unable to parse kvdb connection info %s", *flagBigTable), err)

	//initalise dmesh client
	dmeshStore, err := dmesh.NewStore(*flagMeshStoreAddr)
	derr.Check("unable to setup dmesh store (etcd)", err)
	defer dmeshStore.Close()

	var diagnose Diagnose

	switch *flagProtocol {
	case "EOS":
		db, err := eosdb.New(kvdbInfo.TablePrefix, kvdbInfo.Project, kvdbInfo.Instance, false)
		derr.Check("failed setting up bigtable for EOS", err)
		diagnose = &eos.Diagnose{
			Namespace:             *flagNamespace,
			BlocksStoreUrl:        *flagBlocksStore,
			SearchIndexesStoreUrl: *flagSearchIndexesStore,
			SearchShardSize:       uint32(*flagSearchShardSize),
			KvdbConnectionInfo:    *flagBigTable,
			BlocksStore:           blocksStore,
			SearchStore:           searchStore,
			EOSdb:                 db,
			ParallelDownloadCount: *flagParallelDownloadCount,
		}

	case "ETH":
		db, err := ethdb.New(kvdbInfo.TablePrefix, kvdbInfo.Project, kvdbInfo.Instance, false)
		derr.Check("failed setting up bigtable for ETH", err)

		diagnose = &eth.Diagnose{
			Namespace:             *flagNamespace,
			BlocksStoreUrl:        *flagBlocksStore,
			SearchIndexesStoreUrl: *flagSearchIndexesStore,
			SearchShardSize:       uint32(*flagSearchShardSize),
			KvdbConnectionInfo:    *flagBigTable,
			BlocksStore:           blocksStore,
			SearchStore:           searchStore,
			ETHdb:                 db,
		}

	}

	performK8sSetup := !*flagSkipK8S
	var cluster *kubernetes.Clientset
	if performK8sSetup {
		zlog.Info("setting up k8s clientset")
		config, err := rest.InClusterConfig()
		derr.Check("unable to retrieve kubernetes cluster config", err)

		cluster, err = kubernetes.NewForConfig(config)
		derr.Check("unable to create kubernetes client set", err)
	}

	r := NewRootServer(*flagHttpListenAddr, *flagProtocol, *flagNamespace, *flagBlocksStore, *flagSearchIndexesStore, *flagBigTable, uint32(*flagSearchShardSize), *flagMeshServiceVersion, diagnose, cluster, dmeshStore)

	r.SetupRoutes()

	zlog.Info("serving http")
	err = r.Serve()
	derr.Check("failed serving http", err)
}
