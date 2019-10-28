package main

import (
	"flag"
	"fmt"

	"github.com/eoscanada/derr"
	"github.com/eoscanada/eos-go"
)

var flagHttpListenAddr = flag.String("listen-http-addr", ":8080", "TCP Listener addr for http")

var flagProtocol = flag.String("protocol", "ETH", "Protocol to load, EOS or ETH")
var flagNamespace = flag.String("namespace", "eth-ropsten", "k8s namespace inspected by this diagnose instance")

var flagBlocksStore = flag.String("blocks-store", "gs://dfuseio-global-blocks-us/eth-ropsten/v2", "Blocks logs storage location")

var flagSearchIndexesStore = flag.String("search-indexes-store", "gs://dfuseio-global-indices-us/eth-ropsten/v2", "GS location of search indexes storage for EOS")
var flagSearchShardSize = flag.Uint("search-shard-size", 200, "Number of blocks to store in a given Bleve index")

var flagDB = flag.String("db-connection", "dfuseio-global:dfuse-saas:ropsten-v2", "Big table connection string as 'project:instance:table-prefix'")

var flagAPIURL = flag.String("api-url", "https://ropsten.eth.dfuse.io", "The API node to reach for information about the chain")

var flagParallelDownloadCount = flag.Uint64("parallel-download-count", 6, "How many blocks file to download in parallel")
var flagSkipK8S = flag.Bool("skip-k8s", false, "Useful in development to avoid setuping access to a K8S cluster")

func main() {
	flag.Parse()
	setupLogger()

	d := Diagnose{
		Addr:                  *flagHttpListenAddr,
		Namespace:             *flagNamespace,
		Protocol:              *flagProtocol,
		Api:                   eos.New(*flagAPIURL),
		ParallelDownloadCount: *flagParallelDownloadCount,
		BlockStoreUrl:         *flagBlocksStore,
		SearchIndexesStoreUrl: *flagSearchIndexesStore,
		SearchShardSize:       uint32(*flagSearchShardSize),
		KvdbConnection:        *flagDB,
	}

	zlog.Info("setting up stores")
	err := d.SetupStores(*flagBlocksStore, *flagSearchIndexesStore)
	derr.Check("failed setting up store", err)

	performK8sSetup := !*flagSkipK8S
	if performK8sSetup {
		zlog.Info("setting up k8s clientset")
		err = d.SetupK8s()
		derr.Check("failed setting up k8s", err)
	}

	zlog.Info("setting up kvdb")
	err = d.SetupDB(*flagDB, *flagProtocol)
	derr.Check(fmt.Sprintf("failed setting up bigtable for %s", *flagProtocol), err)

	d.SetupRoutes()

	zlog.Info("serving http")
	err = d.Serve()
	derr.Check("failed serving http", err)
}
