package main

import (
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	bt "cloud.google.com/go/bigtable"
	"github.com/abourget/llerrgroup"
	"github.com/eoscanada/bstream/store"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eosdb"
	"github.com/eoscanada/eosdb/bigtable"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Diagnose struct {
	addr   string
	routes *mux.Router

	api *eos.API

	namespace string
	bigtable  *bigtable.Bigtable
	eosdb     eosdb.DBReader

	blocksStore store.ArchiveStore
	searchStore *store.SimpleGStore

	cluster *kubernetes.Clientset

	parallelDownloadCount uint64
}

func (d *Diagnose) setupRoutes() {
	configureValidators()

	r := mux.NewRouter()
	r.Path("/").Methods("GET").HandlerFunc(d.index)
	r.Path("/dfuse.css").Methods("GET").HandlerFunc(d.css)

	r.Path("/v1/diagnose/verify_stats").Methods("GET").HandlerFunc(d.verifyStats)
	r.Path("/v1/diagnose/verify_stats_top_accounts").Methods("GET").HandlerFunc(d.verifyStatsTopAccounts)
	r.Path("/v1/diagnose/verify_eosdb_block_holes").Methods("GET").HandlerFunc(d.verifyEOSDBBlockHoles)
	r.Path("/v1/diagnose/verify_eosdb_trx_problems").Methods("GET").HandlerFunc(d.verifyEOSDBTrxProblems)
	r.Path("/v1/diagnose/verify_blocks_holes").Methods("GET").HandlerFunc(d.verifyBlocksHoles)
	r.Path("/v1/diagnose/verify_search_holes").Methods("GET").HandlerFunc(d.verifySearchHoles)
	r.Path("/v1/diagnose/services_health_checks").Methods("GET").HandlerFunc(d.getServicesHealthChecks)
	r.Path("/v1/diagnose/").Methods("POST").Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	d.routes = r
}

var doingEOSDBBlockHoles bool

func (d *Diagnose) verifyEOSDBBlockHoles(w http.ResponseWriter, r *http.Request) {
	putPreambule(w, "Checking block holes in EOSDB")
	if doingEOSDBBlockHoles {
		putLine(w, "<h1>Already running, try later</h1>")
		return
	}

	doingEOSDBBlockHoles = true
	defer func() { doingEOSDBBlockHoles = false }()

	count := int64(0)
	holeFound := false
	started := false
	previousNum := int64(0)
	previousValidNum := int64(0)

	startTime := time.Now()
	batchStartTime := time.Now()

	blocksTable := d.bigtable.Blocks

	// You can test on a lower range with `bt.NewRange("ff76abbf", "ff76abcf")`
	err := blocksTable.BaseTable.ReadRows(r.Context(), bt.InfiniteRange(""), func(row bt.Row) bool {
		count++

		num := int64(math.MaxUint32 - bigtable.BlockNum(row.Key()))

		isValid := hasAllColumns(row, blocksTable.ColBlockJSON, blocksTable.ColMetaHeader, blocksTable.ColMetaWritten, blocksTable.ColMetaIrreversible, blocksTable.ColTrxExecutedIDs)

		if !started {
			previousNum = num + 1
			previousValidNum = num + 1
			batchStartTime = time.Now()

			putLine(w, "<p><strong>Start block %d</strong></p>\n", num)
		}

		difference := previousNum - num
		differenceInvalid := previousValidNum - num

		if difference > 1 && started {
			holeFound = true
			putLine(w, "<p><strong>Found block hole: [%d, %d]</strong></p>\n", num+1, previousNum-1)
		}

		if differenceInvalid > 1 && started && isValid {
			holeFound = true
			putLine(w, "<p><strong>Found missing column(s) hole: [%d, %d]</strong></p>\n", num+1, previousValidNum-1)
		}

		previousNum = num
		if isValid {
			previousValidNum = num
		}

		if count%200000 == 0 {
			now := time.Now()
			putLine(w, "<p>200K rows read @ #%d (batch %s, total %s) ...</p>\n", num, now.Sub(batchStartTime), now.Sub(startTime))
			batchStartTime = time.Now()
		}

		started = true

		return true
	}, bt.RowFilter(bt.StripValueFilter()))

	differenceInvalid := previousValidNum - previousNum
	if differenceInvalid > 1 && started {
		holeFound = true
		putLine(w, "<p><strong>Found missing column(s) hole: [%d, %d]</strong></p>\n", previousNum, previousValidNum-1)
	}

	if err != nil {
		putLine(w, "<p><strong>Error: %s</strong></p>\n", err.Error())
		return
	}

	if !holeFound {
		putLine(w, "<p><strong>NO HOLE FOUND!</strong></p>\n")
	}

	putLine(w, "<p><strong>Completed at block num %d (%d blocks seen) in %s</strong></p>\n", previousNum, count, time.Now().Sub(startTime))
}

var doingEOSDBTrxProblems = false

func (d *Diagnose) verifyEOSDBTrxProblems(w http.ResponseWriter, r *http.Request) {
	putPreambule(w, "Checking transaction problems in EOSDB")

	if doingEOSDBTrxProblems {
		putLine(w, "<h1>Already running, try later</h1>")
		return
	}

	doingEOSDBTrxProblems = true
	defer func() { doingEOSDBTrxProblems = false }()

	count := int64(0)
	problemFound := false
	startTime := time.Now()

	trxsTable := d.bigtable.Transactions

	processRowRange := func(rowRange bt.RowSet) error {
		return trxsTable.BaseTable.ReadRows(r.Context(), rowRange, func(row bt.Row) bool {
			key := row.Key()
			trxID := key[0:64]
			prefixTrxID := trxID[0:8]
			blockNum := bigtable.BlockNum(key[65:73])

			count++
			problemFound = true
			putSyncLine(w, `<p><strong>Found problem with <a href="%s">%s</a> @ #%d (missing meta:written column)</strong></p>`+"\n", inferEosqTrxLink(trxID), prefixTrxID, blockNum)

			return true
		}, bt.RowFilter(bt.ConditionFilter(bt.ColumnFilter("meta:written"), nil, bt.StripValueFilter())))
	}

	concurrentReadCount := runtime.NumCPU() - 1
	if concurrentReadCount > 16 {
		concurrentReadCount = 16
	}

	rowRanges := createTrxRowSets(concurrentReadCount)
	group := llerrgroup.New(concurrentReadCount)

	putLine(w, "<h2>Starting groups (concurrency %d)</h2>", concurrentReadCount)
	putLine(w, "<small>Note: there's no progress report within a group</small>")

	for _, rowRange := range rowRanges {
		rowRange := rowRange

		if group.Stop() {
			putSyncLine(w, "<h4>Group completed %s</h4>", rowRange)
			break
		}

		group.Go(func() error {
			putSyncLine(w, "<h4>Group range %s starting...</h4>", rowRange)
			return processRowRange(rowRange)
		})
	}

	zlog.Debug("waiting for all parallel stream rows operation to finish")
	if err := group.Wait(); err != nil {
		putLine(w, "<p><strong>Error: %s</strong></p>\n", err.Error())
		return
	}

	if !problemFound {
		putLine(w, "<p><strong>No problem found!</strong></p>\n")
	}

	putLine(w, "<p><strong>Completed (%d problematic trxs seen) in %s</strong></p>\n", count, time.Now().Sub(startTime))

}

var doingBlocksHoles bool

func (d *Diagnose) verifyBlocksHoles(w http.ResponseWriter, r *http.Request) {
	putPreambule(w, "Checking holes in block logs")
	if doingBlocksHoles {
		putLine(w, "<h1>Already running, try later</h1>")
		return
	}

	doingBlocksHoles = true
	defer func() { doingBlocksHoles = false }()

	number := regexp.MustCompile(`(\d{10})`)

	var holeFound bool
	var expected uint32
	var count int
	err := d.blocksStore.Walk("", func(filename string) error {
		match := number.FindStringSubmatch(filename)
		if match == nil {
			return nil
		}

		count++
		baseNum, _ := strconv.ParseUint(match[1], 10, 32)
		baseNum32 := uint32(baseNum)
		if baseNum32 != expected {
			holeFound = true
			putLine(w, "<p><strong>HOLE FOUND: %d - %d</strong></p>\n", expected, baseNum)
		}
		expected = baseNum32 + 100

		if count%10000 == 0 {
			putLine(w, "<p>%s...</p>\n", filename)
		}

		return nil
	})
	if err != nil {
		putLine(w, "<pre>Failed walking file list: %s</pre>\n", err)
		return
	}

	if !holeFound {
		putLine(w, "<p><strong>NO HOLE FOUND!</strong></p>\n")
	}

	putLine(w, "<p>Validated up to block log: %d</p>\n", expected-100)
}

func (d *Diagnose) verifySearchHoles(w http.ResponseWriter, r *http.Request) {
	putPreambule(w, "Checking holes in Search indexes")

	number := regexp.MustCompile(`.*/(\d+)\.bleve\.tar\.gz`)

	fileList, err := d.searchStore.ListFiles("shards-5000/", math.MaxUint32)
	if err != nil {
		putLine(w, "<pre>Failed walking file list: %s</pre>", err)
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
			putLine(w, "<p><strong>HOLE FOUND: %d - %d</strong></p>\n", expected, baseNum)
		}
		expected = baseNum32 + 5000
	}

	if !holeFound {
		putLine(w, "<p><strong>NO HOLE FOUND!</strong></p>")
	}

	putLine(w, "<p>Validated up to index: %d</p>", expected-5000)
}

func (d *Diagnose) getServicesHealthChecks(w http.ResponseWriter, r *http.Request) {
	putPreambule(w, "Services health checks")

	services, err := d.cluster.CoreV1().Services(d.namespace).List(meta_v1.ListOptions{})
	if err != nil {
		putLine(w, "<pre>Failed listing services: %s</pre>", err)
		return
	}

	for _, svc := range services.Items {
		endpoints, err := d.cluster.CoreV1().Endpoints(d.namespace).Get(svc.Name, meta_v1.GetOptions{})
		if err != nil {
			putLine(w, "<pre>failed getting service %q: %s</pre>", svc.Name, err)
			continue
		}

		putLine(w, "<h4>Service: %q</h4>", svc.Name)

		for _, subset := range endpoints.Subsets {
			for _, addr := range subset.Addresses {
				for _, port := range subset.Ports {
					theURL := fmt.Sprintf("http://%s:%d/healthz?secret=dfuse&full=true", addr.IP, port.Port)
					putLine(w, "<pre>Querying %s\n", theURL)
					// Query the health endpoint
					resp, err := http.Get(theURL)
					if err != nil {
						putLine(w, "Failed: %s\n", err)
					} else {
						cnt, _ := ioutil.ReadAll(resp.Body)
						putLine(w, "Status: %d\n\n", resp.StatusCode)
						putLine(w, string(cnt))
						putLine(w, "\n")
					}
					putLine(w, "</pre>")
				}
			}
		}
	}
}

func (d *Diagnose) index(w http.ResponseWriter, r *http.Request) {
	data := &tplData{}

	d.renderTemplate(w, data)
}

func (d *Diagnose) css(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "dfuse.css")
}

func (d *Diagnose) Serve() error {
	zlog.Info("Serving on address", zap.String("addr", d.addr))
	return http.ListenAndServe(d.addr, d.routes)
}

func putPreambule(w http.ResponseWriter, title string) {
	putLine(w, `<html><head><title>%s</title><link rel="stylesheet" type="text/css" href="/dfuse.css"></head><body><div style="width:90%%; margin: 2rem auto;"><h1>%s</h1>`, title, title)
}

func putErrorLine(w http.ResponseWriter, prefix string, err error) {
	putLine(w, "<p><strong>%s: %s</strong></p>\n", prefix, err.Error())
}

var lock sync.Mutex

func putSyncLine(w http.ResponseWriter, format string, v ...interface{}) {
	line := fmt.Sprintf(format, v...)

	flush := w.(http.Flusher)

	lock.Lock()
	defer lock.Unlock()

	fmt.Fprint(w, line)
	flush.Flush()

	zlog.Info("html output line", zap.String("line", line))
}

func putLine(w http.ResponseWriter, format string, v ...interface{}) {
	line := fmt.Sprintf(format, v...)

	flush := w.(http.Flusher)
	fmt.Fprint(w, line)
	flush.Flush()

	zlog.Info("html output line", zap.String("line", line))
}

func flushWriter(w http.ResponseWriter) {
	flusher := w.(http.Flusher)
	flusher.Flush()
}

func hasAllColumns(row bt.Row, columns ...string) bool {
	for _, column := range columns {
		if !hasBtColumn(row, column) {
			return false
		}
	}

	return true
}

func hasBtColumn(row bt.Row, familyColumn string) bool {
	for _, cols := range row {
		for _, el := range cols {
			if el.Column == familyColumn {
				return true
			}
		}
	}

	return false
}

func inferEosqTrxLink(trxID string) string {
	return inferEosqLink("/tx/" + trxID)
}

// FIXME: Make network configurable so we can link to right place ...
func inferEosqLink(path string) string {
	return fmt.Sprintf("https://eosq.app/%s", path)
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

	// FIXME: Find a way to get up to last possible keys of `a:` set without copying the `prefixSuccessor` method from bigtable
	//        Hard-coded for now.
	rowRanges = append(rowRanges, bt.NewRange(startPrefix, strings.Repeat("f", 64)+";"))

	return rowRanges
}
