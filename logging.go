package main

import (
	"github.com/eoscanada/derr"
	"github.com/eoscanada/logging"
	"go.uber.org/zap"
)

var zlog = zap.NewNop()

func setupLogger() {
	zlog = logging.MustCreateLoggerWithServiceName("diagnose")
	derr.SetLogger(zlog)
	//dmesh.SetLogger(zlog)
}
