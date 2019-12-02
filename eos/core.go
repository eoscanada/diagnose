package eos

import (
	"fmt"
	"math"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/eoscanada/kvdb/eosdb"
	"github.com/gorilla/websocket"

	"github.com/eoscanada/dstore"

	"github.com/abourget/llerrgroup"

	bt "cloud.google.com/go/bigtable"
	"github.com/eoscanada/diagnose/renderer"
	"github.com/eoscanada/kvdb"
)

var processingTrxProblems bool

type Diagnose struct {
	Namespace string

	BlocksStoreUrl        string
	SearchIndexesStoreUrl string
	SearchShardSize       uint32
	KvdbConnectionInfo    string
	upgrader              *websocket.Upgrader

	BlocksStore dstore.Store
	SearchStore dstore.Store

	EOSdb                 *eosdb.EOSDatabase
	ParallelDownloadCount uint64
}

func (e *Diagnose) trxProblems(w http.ResponseWriter, r *http.Request) {
	renderer.PutPreambule(w, "Checking transaction problems in EOSDB")

	if processingTrxProblems {
		renderer.PutLine(w, "<h1>Already running, try later</h1>")
		return
	}

	processingTrxProblems = true
	defer func() { processingTrxProblems = false }()

	count := int64(0)
	problemFound := false
	startTime := time.Now()

	trxsTable := e.EOSdb.Transactions

	processRowRange := func(rowRange bt.RowSet) error {
		return trxsTable.BaseTable.ReadRows(r.Context(), rowRange, func(row bt.Row) bool {
			key := row.Key()
			trxID := key[0:64]
			prefixTrxID := trxID[0:8]
			blockNum := kvdb.BlockNum(key[65:73])

			count++
			problemFound = true
			renderer.PutSyncLine(w, `<p><strong>Found problem with <a href="%s">%s</a> @ #%d (missing meta:written column)</strong></p>`+"\n", inferEosqTrxLink(trxID), prefixTrxID, blockNum)

			return true
		}, bt.RowFilter(bt.ConditionFilter(bt.ColumnFilter("meta:written"), nil, bt.StripValueFilter())))
	}

	concurrentReadCount := runtime.NumCPU() - 1
	if concurrentReadCount > 16 {
		concurrentReadCount = 16
	}

	rowRanges := createTrxRowSets(concurrentReadCount)
	group := llerrgroup.New(concurrentReadCount)

	renderer.PutLine(w, "<h2>Starting groups (concurrency %d)</h2>", concurrentReadCount)
	renderer.PutLine(w, "<small>Note: there's no progress report within a group</small>")

	for _, rowRange := range rowRanges {
		rowRange := rowRange

		if group.Stop() {
			renderer.PutSyncLine(w, "<h4>Group completed %s</h4>", rowRange)
			break
		}

		group.Go(func() error {
			renderer.PutSyncLine(w, "<h4>Group range %s starting...</h4>", rowRange)
			return processRowRange(rowRange)
		})
	}

	zlog.Debug("waiting for all parallel stream rows operation to finish")
	if err := group.Wait(); err != nil {
		renderer.PutLine(w, "<p><strong>Error: %s</strong></p>\n", err.Error())
		return
	}

	if !problemFound {
		renderer.PutLine(w, "<p><strong>No problem found!</strong></p>\n")
	}

	renderer.PutLine(w, "<p><strong>Completed (%d problematic trxs seen) in %s</strong></p>\n", count, time.Now().Sub(startTime))

}

func createTrxRowSets(concurrentReadCount int) []bt.RowSet {
	letters := "123456789abcdef"
	if concurrentReadCount > len(letters)+1 {
		panic(fmt.Errorf("only accepting concurrent <= %d, got %d", len(letters), concurrentReadCount))
	}

	step := int(math.Ceil(float64(len(letters)) / float64(concurrentReadCount)))
	startPrefix := ""
	var endPrefix string

	var rowRanges []bt.RowSet

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

func inferEosqTrxLink(trxID string) string {
	return inferEosqLink("/tx/" + trxID)
}

// FIXME: Make network configurable so we can link to right place ...
func inferEosqLink(path string) string {
	return fmt.Sprintf("https://eosq.app/%s", path)
}
