package main

import (
	"context"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/eoscanada/dstore"
	"go.uber.org/zap"
)

func (d *Diagnose) BlockHoles(w http.ResponseWriter, req *http.Request) {
	const fileBlockSize = 100
	zlog.Info("diagnose - block holes",
		zap.String("block_store_url", d.BlocksStoreUrl),
		zap.Uint32("block_logs_size", fileBlockSize),
	)
	conn, err := d.upgrader.Upgrade(w, req, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(req.Context())

	number := regexp.MustCompile(`(\d{10})`)

	var expected uint32
	var count int
	var baseNum32 uint32
	currentStartBlk := uint32(0)
	startTime := time.Now()

	go readWebsocket(conn, cancel)

	maybeSendWebsocket(conn, WebsocketTypeProgress, Progress{Elapsed: time.Now().Sub(startTime)})
	d.BlocksStore.Walk("", "", func(filename string) error {
		select {
		case <-ctx.Done():
			zlog.Debug("context canceled")
			return dstore.StopIteration
		default:
		}

		if count%5000 == 0 {
			maybeSendWebsocket(conn, WebsocketTypeProgress, Progress{Elapsed: time.Now().Sub(startTime)})
		}

		match := number.FindStringSubmatch(filename)
		if match == nil {
			return nil
		}

		count++
		baseNum, _ := strconv.ParseUint(match[1], 10, 32)
		baseNum32 = uint32(baseNum)

		if baseNum32 != expected {
			maybeSendWebsocket(conn, WebsocketTypeBlockRange, NewValidBlockRange(currentStartBlk, (expected-fileBlockSize), "valid range"))
			maybeSendWebsocket(conn, WebsocketTypeBlockRange, NewMissingBlockRange(expected, (baseNum32-fileBlockSize), "hole found"))
			currentStartBlk = baseNum32
		}
		expected = baseNum32 + fileBlockSize

		if count%10000 == 0 {
			maybeSendWebsocket(conn, WebsocketTypeBlockRange, NewValidBlockRange(currentStartBlk, baseNum32, "valid range"))
			currentStartBlk = baseNum32 + fileBlockSize
		}

		return nil
	})
	maybeSendWebsocket(conn, WebsocketTypeBlockRange, NewValidBlockRange(currentStartBlk, baseNum32, "valid range"))
	zlog.Info("diagnose - block holes - completed")
}
