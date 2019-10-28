package main

import (
	"github.com/eoscanada/diagnose/eos"
	"github.com/eoscanada/diagnose/eth"
	"github.com/eoscanada/diagnose/renderer"
	"github.com/eoscanada/logging"
	"go.uber.org/zap"
)

var zlog = zap.NewNop()

func setupLogger() {
	zlog = logging.MustCreateLoggerWithServiceName("diagose")
	eos.SetLogger(zlog)
	eth.SetLogger(zlog)
	renderer.SetLogger(zlog)
}
