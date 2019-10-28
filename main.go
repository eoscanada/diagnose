package main

import (
	"flag"

	"github.com/eoscanada/derr"
	"github.com/eoscanada/eos-go"
)

var flagHttpListenAddr = flag.String("listen-http-addr", ":8080", "TCP Listener addr for http")
var flagProtocol = flag.String("protocol", "ETH", "Protocol to load, EOS or ETH")
var flagNamespace = flag.String("namespace", "eth-ropsten", "k8s namespace inspected by this diagnose instance")

var flagAPIURL = flag.String("api-url", "https://ropsten.eth.dfuse.io", "The API node to reach for information about the chain")

var flagBlocksStore = flag.String("blocks-store", "gs://dfuseio-global-blocks-us/eth-ropsten/v2", "Blocks logs storage location")

//var flagDB = flag.String("db-connection", "dfuseio-global:dfuse-saas:eth-ropsten", "Big table connection string as 'project:instance:table-prefix'")
var flagSearchIndexesStore = flag.String("search-indexes-store", "gs://dfuseio-global-indices-us/eth-ropsten/v2", "GS location of search indexes storage for EOS")

//var flagSearchShardSize = flag.String("search-shard-size", "500", "Number of blocks to store in a given Bleve index")

var flagEOSDB = flag.String("eosdb-connection", "dfuseio-global:dfuse-saas:aca3-v4", "eosdb connection string as 'project:instance:table-prefix'")
var flagETHDB = flag.String("ethdb-connection", "dfuseio-global:dfuse-saas:ropsten-v2", "ethdb connection string as 'project:instance:table-prefix'")

var flagParallelDownloadCount = flag.Uint64("parallel-download-count", 6, "How many blocks file to download in parallel")
var flagSkipK8S = flag.Bool("skip-k8s", false, "Useful in development to avoid setuping access to a K8S cluster")

func main() {
	flag.Parse()

	d := Diagnose{
		addr:                  *flagHttpListenAddr,
		namespace:             *flagNamespace,
		protocol:              *flagProtocol,
		api:                   eos.New(*flagAPIURL),
		parallelDownloadCount: *flagParallelDownloadCount,
		blockStore:            *flagBlocksStore,
		searchIndexesStore:    *flagSearchIndexesStore,
		kvdbConnection:        "",
	}

	d.setupRoutes()

	zlog.Info("setting up stores")
	err := d.setupStores(*flagBlocksStore, *flagSearchIndexesStore)
	derr.Check("failed setting up store", err)

	performK8sSetup := !*flagSkipK8S
	if performK8sSetup {
		zlog.Info("setting up k8s clientset")
		err = d.setupK8s()
		derr.Check("failed setting up k8s", err)
	}

	zlog.Info("setting up eosdb")
	err = d.setupEOSDB(*flagEOSDB)
	derr.Check("failed setting up bigtable for EOS", err)

	zlog.Info("setting up ethdb")
	err = d.setupETHDB(*flagETHDB)
	derr.Check("failed setting up bigtable for ETH", err)

	zlog.Info("serving http")
	err = d.Serve()
	derr.Check("failed serving http", err)
}
