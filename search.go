package main

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"

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

	ctx := req.Context()

	number := regexp.MustCompile(`.*/(\d+)\.bleve\.tar\.(zst|gz)$`)

	var expected uint32
	var count int
	var baseNum32 uint32

	shardPrefix := fmt.Sprintf("shards-%d/", d.SearchShardSize)
	currentStartBlk := uint32(0)

	go websocketRead(conn)

	_ = d.SearchStore.Walk(shardPrefix, "", func(filename string) error {

		select {
		case <-ctx.Done():
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
			_ = sendMessage(conn, NewValidBlockRange(currentStartBlk, (expected-d.SearchShardSize), "valid range"))
			_ = sendMessage(conn, NewMissingBlockRange(expected, (baseNum32-d.SearchShardSize), "hole found"))
			currentStartBlk = baseNum32
		}
		expected = baseNum32 + d.SearchShardSize

		if count%1000 == 0 {
			_ = sendMessage(conn, NewValidBlockRange(currentStartBlk, baseNum32, "valid range"))
			currentStartBlk = baseNum32 + d.SearchShardSize
		}

		return nil
	})
	_ = sendMessage(conn, NewValidBlockRange(currentStartBlk, baseNum32, "valid range"))
	zlog.Info("diagnose - search indexes - completed")
}
