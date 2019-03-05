package main

import (
	"flag"
	"strings"

	"go.uber.org/zap"
)

var flagHealthCheckServices = flag.String("health-check-services", "relayer-v9", "Comma-separated list of services for which endpoints we want to check health-checks")
var flagNamespace = flag.String("namespace", "default", "k8s namespace inspected by this diagnose instance")
var flagEOSDB = flag.String("eosdb-connection", "", "eosdb connection string as 'project:instance:table-prefix'")

func main() {
	flag.Parse()

	d := Diagnose{
		addr:           ":8080",
		healthServices: strings.Split(*flagHealthCheckServices, ","),
		namespace:      *flagNamespace,
	}

	d.setupRoutes()

	if err := d.setupK8s(); err != nil {
		zlog.Fatal("failed setting up k8s", zap.Error(err))
	}

	if err := d.setupEOSDB(*flagEOSDB); err != nil {
		zlog.Fatal("failed bigtable setup", zap.Error(err))
	}

	if err := d.Serve(); err != nil {
		zlog.Error("failed listening", zap.Error(err))
	}
}
