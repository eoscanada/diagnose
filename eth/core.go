package eth

import (
	"fmt"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/eoscanada/diagnose/utils"

	"github.com/eoscanada/dstore"

	"github.com/eoscanada/diagnose/renderer"

	bt "cloud.google.com/go/bigtable"

	"github.com/eoscanada/kvdb/ethdb"

	"github.com/gorilla/mux"
)

var processingBlockHoles bool
var processingDBHoles bool
var processingSearchHoles bool

type ETHDiagnose struct {
	Namespace string

	SearchShardSize string

	BlocksStore dstore.Store
	SearchStore dstore.Store

	ETHdb *ethdb.ETHDatabase
}

func (e *ETHDiagnose) SetupRoutes(s *mux.Router) {

	s.Path("/block_holes").Methods("GET").HandlerFunc(e.blockHoles)
	s.Path("/db_holes").Methods("GET").HandlerFunc(e.dbHoles)
	s.Path("/search_holes").Methods("GET").HandlerFunc(e.searchHoles)
}

func (e *ETHDiagnose) blockHoles(w http.ResponseWriter, r *http.Request) {
	renderer.PutPreambule(w, "Checking holes in block logs")
	if processingBlockHoles {
		renderer.PutLine(w, "<h1>Already running, try later</h1>")
		return
	}

	processingBlockHoles = true
	defer func() { processingBlockHoles = false }()

	number := regexp.MustCompile(`(\d{10})`)

	var holeFound bool
	var expected uint32
	var count int
	err := e.BlocksStore.Walk("", "", func(filename string) error {
		match := number.FindStringSubmatch(filename)
		if match == nil {
			return nil
		}

		count++
		baseNum, _ := strconv.ParseUint(match[1], 10, 32)
		baseNum32 := uint32(baseNum)
		if baseNum32 != expected {
			holeFound = true
			renderer.PutLine(w, "<p><strong>HOLE FOUND: %d - %d</strong></p>\n", expected, baseNum)
		}
		expected = baseNum32 + 100

		if count%10000 == 0 {
			renderer.PutLine(w, "<p>%s...</p>\n", filename)
		}

		return nil
	})
	if err != nil {
		renderer.PutLine(w, "<pre>Failed walking file list: %s</pre>\n", err)
		return
	}

	if !holeFound {
		renderer.PutLine(w, "<p><strong>NO HOLE FOUND!</strong></p>\n")
	}

	renderer.PutLine(w, "<p>Validated up to block log: %d</p>\n", expected-100)
}

func (e *ETHDiagnose) dbHoles(w http.ResponseWriter, r *http.Request) {
	renderer.PutPreambule(w, "Checking block holes in ETHDB")
	if processingDBHoles {
		renderer.PutLine(w, "<h1>Already running, try later</h1>")
		return
	}

	processingDBHoles = true
	defer func() { processingDBHoles = false }()

	count := int64(0)
	holeFound := false
	started := false
	previousNum := uint64(0)
	previousValidNum := uint64(0)

	startTime := time.Now()
	batchStartTime := time.Now()

	blocksTable := e.ETHdb.Blocks

	// You can test on a lower range with `bt.NewRange("ff76abbf", "ff76abcf")`
	err := blocksTable.BaseTable.ReadRows(r.Context(), bt.InfiniteRange("blkn:"), func(row bt.Row) bool {
		count++

		num, _, e := ethdb.Keys.ReadBlockNum(row.Key())
		if e != nil {
			renderer.PutLine(w, "<p><strong>Error: %s</strong></p>\n", e.Error())
			return false
		}

		isValid := utils.HasAllColumns(row, blocksTable.ColHeaderProto, blocksTable.ColMetaIrreversible, blocksTable.ColMetaMapping, blocksTable.ColMetaWritten, blocksTable.ColTrxRefsProto, blocksTable.ColUnclesProto)

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

func (e *ETHDiagnose) searchHoles(w http.ResponseWriter, r *http.Request) {
	renderer.PutPreambule(w, "Checking holes in Search indexes")
	if processingSearchHoles {
		renderer.PutLine(w, "<h1>Already running, try later</h1>")
		return
	}

	processingSearchHoles = true
	defer func() { processingSearchHoles = false }()

	number := regexp.MustCompile(`.*/(\d+)\.bleve\.tar\.(zst|gz)$`)

	shardPrefix := fmt.Sprintf("shards-%s/", e.SearchShardSize)
	fileList, err := e.SearchStore.ListFiles(shardPrefix, "", math.MaxUint32)
	if err != nil {
		renderer.PutLine(w, "<pre>Failed walking file list: %s</pre>", err)
		return
	}

	var holeFound bool
	var expected uint32
	for _, filename := range fileList {
		match := number.FindStringSubmatch(filename)
		if match == nil {
			continue
		}

		baseNum, _ := strconv.ParseUint(match[1], 10, 32)
		baseNum32 := uint32(baseNum)
		if baseNum32 != expected {
			holeFound = true
			renderer.PutLine(w, "<p><strong>HOLE FOUND: %d - %d</strong></p>\n", expected, baseNum)
		}
		expected = baseNum32 + 200
	}

	if !holeFound {
		renderer.PutLine(w, "<p><strong>NO HOLE FOUND!</strong></p>")
	}

	renderer.PutLine(w, "<p>Validated up to index: %d</p>", expected-200)
}
