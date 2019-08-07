package main

import (
	"github.com/eoscanada/bstream"
	"github.com/eoscanada/derr"
	"github.com/eoscanada/dtracing"
	"github.com/eoscanada/logging"
)

var zlog = logging.MustCreateLogger()

func init() {
	bstream.SetLogger(zlog)
	derr.SetLogger(zlog)
	dtracing.SetLogger(zlog)
}
