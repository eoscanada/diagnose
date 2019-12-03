package main

import (
	"net/http"
	"regexp"
	"strconv"

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

	ctx := req.Context()

	number := regexp.MustCompile(`(\d{10})`)

	var expected uint32
	var count int
	var baseNum32 uint32
	currentStartBlk := uint32(0)

	go websocketRead(conn)

	d.BlocksStore.Walk("", "", func(filename string) error {
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
			sendMessage(conn, NewValidBlockRange(currentStartBlk, (expected-fileBlockSize), "valid range"))
			sendMessage(conn, NewMissingBlockRange(expected, (baseNum32-fileBlockSize), "hole found"))
			currentStartBlk = baseNum32
		}
		expected = baseNum32 + fileBlockSize

		if count%10000 == 0 {
			sendMessage(conn, NewValidBlockRange(currentStartBlk, baseNum32, "valid range"))
			currentStartBlk = baseNum32 + fileBlockSize
		}

		return nil
	})
	sendMessage(conn, NewValidBlockRange(currentStartBlk, baseNum32, "valid range"))
	zlog.Info("diagnose - block holes - completed")
}
