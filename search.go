package main

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
)

var processingSearchHoles bool

func (e *Diagnose) SearchHoles(w http.ResponseWriter, req *http.Request) {
	if processingSearchHoles {
		// Print out to progress
		return
	}
	processingSearchHoles = true
	defer func() { processingSearchHoles = false }()

	conn, err := e.upgrader.Upgrade(w, req, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(req.Context())

	number := regexp.MustCompile(`.*/(\d+)\.bleve\.tar\.(zst|gz)$`)

	var expected uint32
	var count int
	var baseNum32 uint32

	shardPrefix := fmt.Sprintf("shards-%d/", e.SearchShardSize)
	currentStartBlk := uint32(0)

	go websocketCloser(conn, cancel)

	e.SearchStore.Walk(shardPrefix, "", func(filename string) error {

		select {
		case <-ctx.Done():
			fmt.Println("CONTEXT CANCELED")
			return context.Canceled
		default:

		}

		match := number.FindStringSubmatch(filename)
		if match == nil {
			return nil
		}

		count++
		baseNum, _ := strconv.ParseUint(match[1], 10, 32)
		baseNum32 = uint32(baseNum)
		fmt.Printf("checking %d, expected %d\n", baseNum32, expected)
		if baseNum32 != expected {
			sendMessage(conn, NewValidBlockRange(currentStartBlk, (expected-e.SearchShardSize), "valid range"))
			sendMessage(conn, NewMissingBlockRange(expected, (baseNum32-e.SearchShardSize), "hole found"))
			currentStartBlk = baseNum32
		}
		expected = baseNum32 + e.SearchShardSize

		if count%1000 == 0 {
			sendMessage(conn, NewValidBlockRange(currentStartBlk, baseNum32, "valid range"))
			currentStartBlk = baseNum32 + e.SearchShardSize
		}

		return nil
	})
	sendMessage(conn, NewValidBlockRange(currentStartBlk, baseNum32, "valid range"))
}
