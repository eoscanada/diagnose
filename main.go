package main

import (
	"flag"

	"github.com/eoscanada/derr"
	"github.com/eoscanada/eos-go"
)

var flagNamespace = flag.String("namespace", "default", "k8s namespace inspected by this diagnose instance")
var flagAPIURL = flag.String("api-url", "https://mainnet.eos.dfuse.io", "The EOSIO API node to reach for information about the chain")
var flagEOSDB = flag.String("eosdb-connection", "dfuseio-global:dfuse-saas:aca3-v4", "eosdb connection string as 'project:instance:table-prefix'")
var flagBlocksStore = flag.String("blocks-store", "gs://dfuseio-global-blocks-us/eos-mainnet/aca3/v2", "Blocks logs storage location")
var flagSearchIndexesStore = flag.String("search-indexes-store", "gs://dfuseio-global-indices-us/eos-mainnet/aca3-v12", "GS location of search indexes storage")
var flagParallelDownloadCount = flag.Uint64("parallel-download-count", 6, "How many blocks file to download in parallel")
var flagSkipK8S = flag.Bool("skip-k8s", false, "Useful in development to avoid setuping access to a K8S cluster")

func main() {
	flag.Parse()

	d := Diagnose{
		addr:                  ":8080",
		namespace:             *flagNamespace,
		api:                   eos.New(*flagAPIURL),
		parallelDownloadCount: *flagParallelDownloadCount,
	}

	d.setupRoutes()

	zlog.Info("setting up stores")
	err := d.setupStores(*flagBlocksStore, *flagSearchIndexesStore)
	derr.ErrorCheck("failed setting up store", err)

	performK8sSetup := !*flagSkipK8S
	if performK8sSetup {
		zlog.Info("setting up k8s clientset")
		err = d.setupK8s()
		derr.ErrorCheck("failed setting up k8s", err)
	}

	zlog.Info("setting up eosdb")
	err = d.setupEOSDB(*flagEOSDB)
	derr.ErrorCheck("failed setting up bigtable", err)

	zlog.Info("serving http")
	err = d.Serve()
	derr.ErrorCheck("failed serving http", err)
}
