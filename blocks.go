package main

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
)

var processingBlockHoles bool

func (d *Diagnose) BlockHoles(w http.ResponseWriter, req *http.Request) {
	if processingBlockHoles {
		return
	}

	processingBlockHoles = true
	defer func() { processingBlockHoles = false }()

	conn, err := d.upgrader.Upgrade(w, req, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(req.Context())

	number := regexp.MustCompile(`(\d{10})`)
	const fileBlockSize = 100
	var expected uint32
	var count int
	var baseNum32 uint32
	currentStartBlk := uint32(0)

	go websocketCloser(conn, cancel)

	d.BlocksStore.Walk("", "", func(filename string) error {
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
}
