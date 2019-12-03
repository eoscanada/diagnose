package main

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	bt "cloud.google.com/go/bigtable"
	"github.com/eoscanada/dhammer"
	"github.com/eoscanada/kvdb"
	"go.uber.org/zap"
)

func (d *Diagnose) EOSKVDBTrxsValidation(w http.ResponseWriter, req *http.Request) {
	zlog.Info("diagnose - EOS  - KVDB Trx Validation",
		zap.String("kvdb_connection_info", d.KvdbConnectionInfo))

	reqCtx, cancel := context.WithCancel(req.Context())

	conn, err := d.upgrader.Upgrade(w, req, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	go readWebsocket(conn, cancel)

	startTime := time.Now()
	trxsTable := d.EOSdb.Transactions

	processRowRange := func(ctx context.Context, ranges []interface{}) ([]interface{}, error) {
		zlog.Info("processing ranges", zap.Int("range_count", len(ranges)), zap.Reflect("ranges", ranges))
		var results []interface{}
		for _, r := range ranges {
			rowRange, _ := r.(bt.RowRange)
			trxsTable.BaseTable.ReadRows(ctx, rowRange, func(row bt.Row) bool {
				key := row.Key()
				trxID := key[0:64]

				results = append(results, &Transaction{
					Prefix:   trxID[0:8],
					Id:       trxID,
					BlockNum: kvdb.BlockNum(key[65:73]),
				})
				return true
			}, bt.RowFilter(bt.ConditionFilter(bt.ColumnFilter("written"), nil, bt.StripValueFilter())))
		}
		zlog.Info("finished process ranges", zap.Int("trx_count", len(results)))
		return results, nil
	}

	concurrency := 16
	zlog.Info("concurrency count", zap.Int("concurrency_count", concurrency))

	rowRanges := createTrxRowSets(concurrency)

	hammer := dhammer.NewHammer(1, len(rowRanges), processRowRange)
	hammer.Start(reqCtx)
	maybeSendWebsocket(conn, WebsocketTypeProgress, Progress{Elapsed: time.Now().Sub(startTime)})

	for _, rowRange := range rowRanges {
		zlog.Info("pushing in hammer", zap.Reflect("row_range", rowRange.String()))
		maybeSendWebsocket(conn, WebsocketTypeMessage, &Message{
			Msg: fmt.Sprintf("Processing group range: start %s", rowRange.String()),
		})
		hammer.In <- rowRange
	}
	hammer.Close()

	for {
		select {
		case <-hammer.Done():
			zlog.Info("hammer completion")
			return
		case trxInt, ok := <-hammer.Out:
			if !ok {
				return
			}
			trx := trxInt.(*Transaction)
			maybeSendWebsocket(conn, WebsocketTypeTransaction, trx)
		}
	}

}

func createTrxRowSets(concurrentReadCount int) []bt.RowRange {
	letters := "123456789abcdef"
	if concurrentReadCount > len(letters)+1 {
		panic(fmt.Errorf("only accepting concurrent <= %d, got %d", len(letters), concurrentReadCount))
	}

	step := int(math.Ceil(float64(len(letters)) / float64(concurrentReadCount)))
	startPrefix := ""
	var endPrefix string

	var rowRanges []bt.RowRange

	for i := 0; i < len(letters); i += step {
		endPrefix = string(letters[i]) + strings.Repeat("0", 63) + ":"
		rowRanges = append(rowRanges, bt.NewRange(startPrefix, endPrefix))

		startPrefix = endPrefix
	}

	// FIXME: Find a way to get up to last possible keys of `a:` set without copying the `prefixSuccessor` method from eosdb
	//        Hard-coded for now.
	rowRanges = append(rowRanges, bt.NewRange(startPrefix, strings.Repeat("f", 64)+";"))

	return rowRanges
}
