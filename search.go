package main

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/eoscanada/dstore"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func (d *Diagnose) SearchHoles(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)

	shardSize, err := strconv.ParseUint(params["shard_size"], 10, 32)
	if err != nil {
		shardSize = uint64(d.SearchShardSize)
	}

	indexesURL := params["indexes_url"]
	if indexesURL == "" {
		indexesURL = d.SearchIndexesStoreURL
	}

	zlog.Info("diagnose - search indexes",
		zap.String("indexes_store_url", indexesURL),
		zap.Uint32("default_shard_size", uint32(shardSize)),
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
	seenFirstBlock := false

	shardPrefix := fmt.Sprintf("shards-%d/", shardSize)
	startTime := time.Now()

	currentStartBlk := uint32(0)

	go readWebsocket(conn, cancel)

	zlog.Info("creating indexes store")
	searchStore, err := dstore.NewSimpleStore(indexesURL)
	if err != nil {
		maybeSendWebsocket(conn, WebsocketTypeMessage, Message{Msg: err.Error()})
		return
	}

	maybeSendWebsocket(conn, WebsocketTypeProgress, Progress{Elapsed: time.Now().Sub(startTime)})
	_ = searchStore.Walk(shardPrefix, "", func(filename string) error {

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

		if !seenFirstBlock {
			currentStartBlk = baseNum32
			expected = baseNum32
			seenFirstBlock = true
		}

		if baseNum32 != expected {
			maybeSendWebsocket(conn, WebsocketTypeBlockRange, NewValidBlockRange(currentStartBlk, (expected-uint32(shardSize)), "valid range"))
			maybeSendWebsocket(conn, WebsocketTypeBlockRange, NewMissingBlockRange(expected, (baseNum32-uint32(shardSize)), "hole found"))
			currentStartBlk = baseNum32
		}
		expected = baseNum32 + uint32(shardSize)

		if count%1000 == 0 {
			maybeSendWebsocket(conn, WebsocketTypeBlockRange, NewValidBlockRange(currentStartBlk, baseNum32, "valid range"))
			currentStartBlk = baseNum32 + uint32(shardSize)
		}

		return nil
	})
	maybeSendWebsocket(conn, WebsocketTypeBlockRange, NewValidBlockRange(currentStartBlk, baseNum32, "valid range"))
	zlog.Info("diagnose - search indexes - completed")
}
