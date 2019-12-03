package main

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/eoscanada/dstore"
	"go.uber.org/zap"
)

func (d *Diagnose) SearchHoles(w http.ResponseWriter, req *http.Request) {
	zlog.Info("diagnose - search indexes",
		zap.String("search_indexes_store_url", d.SearchIndexesStoreUrl),
		zap.Uint32("shard_size", d.SearchShardSize),
	)
	conn, err := d.upgrader.Upgrade(w, req, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(req.Context())

	number := regexp.MustCompile(`.*/(\d+)\.bleve\.tar\.(zst|gz)$`)

	var expected uint32
	var count int
	var baseNum32 uint32

	shardPrefix := fmt.Sprintf("shards-%d/", d.SearchShardSize)
	startTime := time.Now()

	currentStartBlk := uint32(0)

	go readWebsocket(conn, cancel)

	maybeSendWebsocket(conn, WebsocketTypeProgress, Progress{Elapsed: time.Now().Sub(startTime)})
	_ = d.SearchStore.Walk(shardPrefix, "", func(filename string) error {

		if count%5000 == 0 {
			maybeSendWebsocket(conn, WebsocketTypeProgress, Progress{Elapsed: time.Now().Sub(startTime)})
		}

		select {
		case <-ctx.Done():
			zlog.Debug("context canceled")
			return dstore.StopIteration
		default:
		}

		match := number.FindStringSubmatch(filename)
		if match == nil {
			return nil
		}

		count++
		baseNum, _ := strconv.ParseUint(match[1], 10, 32)
		baseNum32 = uint32(baseNum)
		if baseNum32 != expected {
			maybeSendWebsocket(conn, WebsocketTypeBlockRange, NewValidBlockRange(currentStartBlk, (expected-d.SearchShardSize), "valid range"))
			maybeSendWebsocket(conn, WebsocketTypeBlockRange, NewMissingBlockRange(expected, (baseNum32-d.SearchShardSize), "hole found"))
			currentStartBlk = baseNum32
		}
		expected = baseNum32 + d.SearchShardSize

		if count%1000 == 0 {
			maybeSendWebsocket(conn, WebsocketTypeBlockRange, NewValidBlockRange(currentStartBlk, baseNum32, "valid range"))
			currentStartBlk = baseNum32 + d.SearchShardSize
		}

		return nil
	})
	maybeSendWebsocket(conn, WebsocketTypeBlockRange, NewValidBlockRange(currentStartBlk, baseNum32, "valid range"))
	zlog.Info("diagnose - search indexes - completed")
}
