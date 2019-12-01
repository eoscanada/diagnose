package eos

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/eoscanada/kvdb/eosdb"
	"github.com/gorilla/websocket"

	"github.com/eoscanada/diagnose/utils"
	"github.com/eoscanada/dstore"

	"github.com/abourget/llerrgroup"

	bt "cloud.google.com/go/bigtable"
	"github.com/eoscanada/diagnose/renderer"
	"github.com/eoscanada/kvdb"
	"github.com/gorilla/mux"
)

var processingTrxProblems bool
var processingBlockHoles bool
var processingDbHoles bool
var processingSearchHoles bool

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

func (e *Diagnose) SetupRoutes(s *mux.Router) {
	s.Path("/block_holes").Methods("GET").HandlerFunc(e.BlockHoles)
	s.Path("/search_holes").Methods("GET").HandlerFunc(e.SearchHoles)

	//////////////////////////////////////////////////////////////////////////////////////////
	s.Path("/trx_problems").Methods("GET").HandlerFunc(e.trxProblems)
	s.Path("/db_holes").Methods("GET").HandlerFunc(e.DBHoles)
	s.Path("/verify_stats").Methods("GET").HandlerFunc(e.verifyStats)
	s.Path("/verify_stats_top_accounts").Methods("GET").HandlerFunc(e.verifyStatsTopAccounts)
}

const (
	BlockRangeStatusValid = "valid"
	BlockRangeStatusHole  = "hole"
)

type BlockRange struct {
	StarBlock uint32 `json:"startBlock"`
	EndBlock  uint32 `json:"endBlock"`
	Message   string `json:"message"`
	Status    string `json:"status"`
}

func NewValidBlockRange(startBlock, endBlock uint32, message string) *BlockRange {
	return &BlockRange{
		StarBlock: startBlock,
		EndBlock:  endBlock,
		Message:   message,
		Status:    BlockRangeStatusValid,
	}
}

func NewMissingBlockRange(startBlock, endBlock uint32, message string) *BlockRange {
	return &BlockRange{
		StarBlock: startBlock,
		EndBlock:  endBlock,
		Message:   message,
		Status:    BlockRangeStatusHole,
	}
}

func (e *Diagnose) SetUpgrader(upgrader *websocket.Upgrader) {
	e.upgrader = upgrader
}

func sendMessage(conn *websocket.Conn, obj interface{}) error {

	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return err
	}
	return nil
}

func printMessage(obj interface{}) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", data)
	return nil

}

func (e *Diagnose) BlockHoles(w http.ResponseWriter, req *http.Request) {
	if processingBlockHoles {
		// Print out to progress
		return
	}

	processingBlockHoles = true
	defer func() { processingBlockHoles = false }()

	conn, err := e.upgrader.Upgrade(w, req, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	number := regexp.MustCompile(`(\d{10})`)

	const fileBlockSize = 100
	var expected uint32
	var count int
	var baseNum32 uint32
	currentStartBlk := uint32(0)

	e.BlocksStore.Walk("", "", func(filename string) error {
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

	number := regexp.MustCompile(`.*/(\d+)\.bleve\.tar\.(zst|gz)$`)

	var expected uint32
	var count int
	var baseNum32 uint32

	shardPrefix := fmt.Sprintf("shards-%d/", e.SearchShardSize)
	currentStartBlk := uint32(0)

	e.SearchStore.Walk(shardPrefix, "", func(filename string) error {
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

func (e *Diagnose) DBHoles(w http.ResponseWriter, r *http.Request) {
	renderer.PutPreambule(w, "Checking block holes in EOSDB")
	if processingDbHoles {
		renderer.PutLine(w, "<h1>Already running, try later</h1>")
		return
	}
	renderer.PutLine(w, "<p>Bigtable: %s</p>\n", e.KvdbConnectionInfo)

	processingDbHoles = true
	defer func() { processingDbHoles = false }()

	count := int64(0)
	holeFound := false
	started := false
	previousNum := int64(0)
	previousValidNum := int64(0)

	startTime := time.Now()
	batchStartTime := time.Now()

	blocksTable := e.EOSdb.Blocks

	// You can test on a lower range with `bt.NewRange("ff76abbf", "ff76abcf")`
	err := blocksTable.BaseTable.ReadRows(r.Context(), bt.InfiniteRange(""), func(row bt.Row) bool {
		count++

		num := int64(math.MaxUint32 - kvdb.BlockNum(row.Key()))

		isValid := utils.HasAllColumns(row, blocksTable.ColBlock, blocksTable.ColTransactionRefs, blocksTable.ColTransactionTraceRefs, blocksTable.ColMetaWritten, blocksTable.ColMetaIrreversible)

		if !started {
			previousNum = num + 1
			previousValidNum = num + 1
			batchStartTime = time.Now()

			renderer.PutLine(w, "<p><strong>Start block %d</strong></p>\n", num)
		}

		difference := previousNum - num
		differenceInvalid := previousValidNum - num

		if difference > 1 && started {
			holeFound = true
			renderer.PutLine(w, "<p><strong>Found block hole: [%d, %d]</strong></p>\n", num+1, previousNum-1)
		}

		if differenceInvalid > 1 && started && isValid {
			holeFound = true
			renderer.PutLine(w, "<p><strong>Found missing column(s) hole: [%d, %d]</strong></p>\n", num+1, previousValidNum-1)
		}

		previousNum = num
		if isValid {
			previousValidNum = num
		}

		if count%200000 == 0 {
			now := time.Now()
			renderer.PutLine(w, "<p>200K rows read @ #%d (batch %s, total %s) ...</p>\n", num, now.Sub(batchStartTime), now.Sub(startTime))
			batchStartTime = time.Now()
		}

		started = true

		return true
	}, bt.RowFilter(bt.StripValueFilter()))

	differenceInvalid := previousValidNum - previousNum
	if differenceInvalid > 1 && started {
		holeFound = true
		renderer.PutLine(w, "<p><strong>Found missing column(s) hole: [%d, %d]</strong></p>\n", previousNum, previousValidNum-1)
	}

	if err != nil {
		renderer.PutLine(w, "<p><strong>Error: %s</strong></p>\n", err.Error())
		return
	}

	if !holeFound {
		renderer.PutLine(w, "<p><strong>NO HOLE FOUND!</strong></p>\n")
	}

	renderer.PutLine(w, "<p><strong>Completed at block num %d (%d blocks seen) in %s</strong></p>\n", previousNum, count, time.Now().Sub(startTime))
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

func (e *Diagnose) GetBlockStoreUrl() string {
	return e.BlocksStoreUrl
}
func (e *Diagnose) GetSearchIndexesStoreUrl() string {
	return e.SearchIndexesStoreUrl
}
func (e *Diagnose) GetKvdbConnectionInfo() string {
	return e.KvdbConnectionInfo
}
