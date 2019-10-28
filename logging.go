package main

import (
	"github.com/eoscanada/bstream"
	"github.com/eoscanada/derr"
	"github.com/eoscanada/diagnose/diagnose"
	"github.com/eoscanada/diagnose/eos"
	"github.com/eoscanada/diagnose/eth"
	"github.com/eoscanada/diagnose/renderer"
	"github.com/eoscanada/dtracing"
	"github.com/eoscanada/logging"
)

var zlog = logging.MustCreateLoggerWithServiceName("diagnose")

func init() {
	bstream.SetLogger(zlog)
	derr.SetLogger(zlog)
	dtracing.SetLogger(zlog)
	diagnose.SetLogger(zlog)
	eos.SetLogger(zlog)
	eth.SetLogger(zlog)
	renderer.SetLogger(zlog)
}
