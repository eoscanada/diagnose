package main

import (
	"github.com/eoscanada/derr"
	"github.com/eoscanada/logging"
)

var zlog = logging.MustCreateLogger()

func init() {
	derr.SetLogger(zlog)
}
