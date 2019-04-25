package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"time"

	bt "cloud.google.com/go/bigtable"

	"github.com/eoscanada/bstream/store"
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

	namespace string
	bigtable  *bigtable.Bigtable
	eosdb     eosdb.DBReader

	blocksStore store.ArchiveStore
	searchStore *store.SimpleGStore

	cluster *kubernetes.Clientset
}

func (d *Diagnose) setupRoutes() {
	r := mux.NewRouter()
	r.Path("/").Methods("GET").HandlerFunc(d.index)
	r.Path("/v1/diagnose/verify_eosdb_holes").Methods("GET").HandlerFunc(d.verifyEOSDBHoles)
	r.Path("/v1/diagnose/verify_blocks_holes").Methods("GET").HandlerFunc(d.verifyBlocksHoles)
	r.Path("/v1/diagnose/verify_search_holes").Methods("GET").HandlerFunc(d.verifySearchHoles)
	r.Path("/v1/diagnose/services_health_checks").Methods("GET").HandlerFunc(d.getServicesHealthChecks)
	r.Path("/v1/diagnose/").Methods("POST").Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	}))

	d.routes = r
}

var doingEOSDBHoles bool

func (d *Diagnose) verifyEOSDBHoles(w http.ResponseWriter, r *http.Request) {
	if doingEOSDBHoles {
		putLine(w, "<h1>Already running, try later</h1>")
		return
	}

	doingEOSDBHoles = true
	defer func() { doingEOSDBHoles = false }()

	putLine(w, "<html><head><title>Checking holes in EOSDB</title></head><h1>Checking holes in EOSDB</h1>")

	count := int64(0)
	holeFound := false
	started := false
	previousNum := int64(0)
	previousValidNum := int64(0)
	startTime := time.Now()

	blocksTable := d.bigtable.Blocks

	// You can test on a lower range with `bt.NewRange("ff76abbf", "ff76abcf")`
	err := blocksTable.BaseTable.ReadRows(context.Background(), bt.InfiniteRange(""), func(row bt.Row) bool {
		count++

		num := int64(math.MaxUint32 - bigtable.BlockNum(row.Key()))

		isValid := hasAllColumns(row, blocksTable.ColBlockJSON, blocksTable.ColMetaHeader, blocksTable.ColMetaWritten, blocksTable.ColMetaIrreversible, blocksTable.ColTrxExecutedIDs)

		if !started {
			previousNum = num + 1
			previousValidNum = num + 1
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
			putLine(w, "<p>#%d ... (%s)</p>\n", num, time.Now().Sub(startTime))
		}

		started = true

		return true
	})

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

var doingBlocksHoles bool

func (d *Diagnose) verifyBlocksHoles(w http.ResponseWriter, r *http.Request) {
	if doingBlocksHoles {
		putLine(w, "<h1>Already running, try later</h1>")
		return
	}

	doingBlocksHoles = true
	defer func() { doingBlocksHoles = false }()

	putLine(w, "<html><head><title>Checking holes in Block logs</title></head><h1>Checking holes in Block logs</h1>")

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
	putLine(w, "<html><head><title>Checking holes in Search indexes</title></head><h1>Checking holes in Search indexes</h1>")

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
	putLine(w, "<html><head><title>Services health checks</title></head><h1>All services health checks</h1>")

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
	// TODO: fetch in-cluster schtuff..

	data := &tplData{}

	d.renderTemplate(w, data)
}

func (d *Diagnose) Serve() error {
	zlog.Info("Serving on address", zap.String("addr", d.addr))
	return http.ListenAndServe(d.addr, d.routes)
}

func putLine(w http.ResponseWriter, format string, v ...interface{}) {
	flush := w.(http.Flusher)
	fmt.Fprintf(w, format, v...)
	flush.Flush()
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
