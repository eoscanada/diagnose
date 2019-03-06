package main

import (
	"flag"
	"os"

	"go.uber.org/zap"
)

var flagNamespace = flag.String("namespace", "default", "k8s namespace inspected by this diagnose instance")
var flagEOSDB = flag.String("eosdb-connection", "", "eosdb connection string as 'project:instance:table-prefix'")
var flagBlocksStore = flag.String("blocks-store", "gs://eoscanada-public-nodeos-archive/nodeos-mainnet-v9", "Blocks logs storage location")
var flagSearchIndexesStore = flag.String("search-indexes-store", "gs://eoscanada-public-indices-archive/search-aca3-v7", "GS location of search indexes storage")

func main() {
	flag.Parse()

	d := Diagnose{
		addr:      ":8080",
		namespace: *flagNamespace,
	}

	d.setupRoutes()

	zlog.Info("setting up stores")
	if err := d.setupStores(*flagBlocksStore, *flagSearchIndexesStore); err != nil {
		zlog.Error("failed setting up store", zap.Error(err))
		os.Exit(1)
	}

	zlog.Info("setting up k8s clientset")
	if err := d.setupK8s(); err != nil {
		zlog.Error("failed setting up k8s", zap.Error(err))
		os.Exit(1)
	}

	zlog.Info("setting up eosdb")
	if err := d.setupEOSDB(*flagEOSDB); err != nil {
		zlog.Error("failed bigtable setup", zap.Error(err))
		os.Exit(1)
	}

	zlog.Info("serving http")
	if err := d.Serve(); err != nil {
		zlog.Error("failed listening", zap.Error(err))
	}
}
