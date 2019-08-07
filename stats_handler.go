package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/eoscanada/bstream/hlog"

	"github.com/eoscanada/bstream"
	"github.com/eoscanada/derr"
	"github.com/eoscanada/dhttp"
	"github.com/eoscanada/validator"
	"go.uber.org/zap"
)

var statsTemplate *template.Template
var statsResultTemplate *template.Template
var doingStats bool

func init() {
	var err error

	statsTemplate, err = template.New("stats").Parse(statsTemplateContent)
	derr.ErrorCheck("unable to create template stats", err)

	statsResultTemplate, err = template.New("stats_results").Parse(statsResultsTemplateContent)
	derr.ErrorCheck("unable to create template stats_results", err)
}

func (d *Diagnose) verifyStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if doingStats {
		dhttp.WriteHTML(ctx, w, alreadyRunningTemplate, nil)
		return
	}

	doingStats = true
	defer func() { doingStats = false }()

	request := &statsRequest{}
	err := dhttp.ExtractRequest(ctx, r, request, dhttp.NewRequestValidator(statsRequestValidationRules))

	fmt.Printf("Request %#v\n", request)

	if err != nil {
		dhttp.WriteError(ctx, w, derr.Wrap(err, "invalid request"))
		return
	}

	startBlockNum, _ := strconv.ParseUint(request.StartBlockNum, 10, 64)
	stopBlockNum, _ := strconv.ParseUint(request.StopBlockNum, 10, 64)

	dhttp.WriteHTML(ctx, w, statsTemplate, map[string]interface{}{
		"StartBlock": startBlockNum,
		"StopBlock":  stopBlockNum,
	})

	stats := &StatsBlockHandler{
		writer:       w,
		stopBlockNum: stopBlockNum,
		logInterval:  logIntervalInBlock(startBlockNum, stopBlockNum),
	}

	source := bstream.NewFileSource(
		bstream.BlockKindEOS,
		d.blocksStore,
		startBlockNum,
		32,
		bstream.PreprocessFunc(stats.preprocessBlock),
		bstream.HandlerFunc(stats.handleBlock),
	)
	source.SetLogger(zlog.With(zap.String("name", "stats")))

	zlog.Info("starting block source")
	source.Run()

	err = source.Err()
	if err != nil && err != io.EOF {
		dhttp.WriteError(ctx, w, derr.Wrap(err, "filesource failed"))
		return
	}

	dhttp.WriteHTML(ctx, w, statsResultTemplate, stats)
}

type statsRequest struct {
	StartBlockNum string `schema:"start_block"`
	StopBlockNum  string `schema:"stop_block"`
}

var statsRequestValidationRules = validator.Rules{
	"start_block": []string{"required", "eos.blockNum"},
	"stop_block":  []string{"required", "eos.blockNum"},
}

// For later when we would like to support out of the box Start Date/End Date and perform the get block query
// func beginningOfPreviousMonth(t time.Time) time.Time {
// 	beginningOfCurrentMonth := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())

// 	return beginningOfCurrentMonth.AddDate(0, -1, 0)
// }

// func endOfPreviousMonth(t time.Time) time.Time {
// 	return beginningOfPreviousMonth(t).AddDate(0, 1, 0).Add(-time.Second)
// }

func logIntervalInBlock(startBlock, stopBlock uint64) int {
	delta := int64(startBlock) - int64(stopBlock)
	if delta < 10000 {
		return 1000
	}

	if delta < 100000 {
		return 10000
	}

	if delta < 1000000 {
		return 100000
	}

	if delta < 10000000 {
		return 1000000
	}

	return 10000000
}

type StatsBlockHandler struct {
	writer http.ResponseWriter

	ActionCount      uint64
	TransactionCount uint64
	BlockCount       uint64

	TokenActionCount      uint64
	TokenNotifCount       uint64
	TokenTransactionCount uint64

	AccountCreationActionCount      uint64
	AccountCreationTransactionCount uint64

	ElapsedTime time.Duration

	stopBlockNum uint64
	timeStart    time.Time
	logInterval  int
}

type BlockStats struct {
	ActionCount      uint64
	TransactionCount uint64

	TokenActionCount      uint64
	TokenNotifCount       uint64
	TokenTransactionCount uint64

	AccountCreationActionCount      uint64
	AccountCreationTransactionCount uint64
}

func (s *StatsBlockHandler) preprocessBlock(block *bstream.Block) (interface{}, error) {
	blockNum := block.Num()
	if blockNum > s.stopBlockNum {
		return nil, nil
	}

	blk := block.ToNative().(*hlog.Block)
	stats := &BlockStats{}

	for _, trx := range blk.AllExecutedTransactionTraces() {
		stats.TransactionCount++

		seenAccountCreationAction := false
		seenTokenTransactionAction := false
		for _, action := range trx.AllActions() {
			stats.ActionCount++

			receiver := action.Receiver()
			account := action.Account()
			actionName := action.ActionName()
			isNotif := receiver != account

			if receiver == "eosio" && account == "eosio" && actionName == "newaccount" {
				stats.AccountCreationActionCount++
				seenAccountCreationAction = true
			}

			if actionName == "transfer" || actionName == "close" || actionName == "issue" || actionName == "retire" {
				stats.TokenActionCount++
				seenTokenTransactionAction = true

				if isNotif {
					stats.TokenNotifCount++
				}
			}
		}

		if seenAccountCreationAction {
			stats.AccountCreationTransactionCount++
		}

		if seenTokenTransactionAction {
			stats.TokenTransactionCount++
		}
	}

	return stats, nil
}

func (s *StatsBlockHandler) handleBlock(block *bstream.Block, obj interface{}) error {
	blockNum := block.Num()
	if blockNum > s.stopBlockNum {
		return io.EOF
	}

	if s.timeStart.IsZero() {
		s.timeStart = time.Now()
		putLine(s.writer, "<small>Processing blocks...</small><br>")
	}

	if s.logInterval > 0 && blockNum%uint64(s.logInterval) == 0 {
		putLine(s.writer, "<small>At block %d</small><br>", blockNum)
	}

	stats := obj.(*BlockStats)
	if stats == nil {
		return nil
	}

	s.BlockCount++
	s.ActionCount += stats.ActionCount
	s.TransactionCount += stats.TransactionCount

	s.TokenActionCount += stats.TokenActionCount
	s.TokenNotifCount += stats.TokenNotifCount
	s.TokenTransactionCount += stats.TokenTransactionCount

	s.AccountCreationActionCount += stats.AccountCreationActionCount
	s.AccountCreationTransactionCount += stats.AccountCreationTransactionCount

	if blockNum == s.stopBlockNum {
		s.ElapsedTime = time.Since(s.timeStart)
	}

	return nil
}

var statsTemplateContent = `
<html>
<head>
    <title>dfuse diagnose</title>
    <link rel="stylesheet" type="text/css" href="/dfuse.css">
</head>
<body>
    <div style="width:90%; margin: 2rem auto;">
		<h2>Verify Stats</h2>

		<ul>
			<li>Start Block: {{ .StartBlock }}</li>
			<li>Stop Block: {{ .StopBlock }}</li>
		</ul>
`

var statsResultsTemplateContent = `
	<h3>Stats Results</h3>
	<ul>
		<li>Block Count: {{ .BlockCount }}</li>
		<li>Transaction Count: {{ .TransactionCount }}</li>
		<li>Action Count: {{ .ActionCount }}</li>

		<br>

		<li>Account Creation Transaction Count: {{ .AccountCreationTransactionCount }}</li>
		<li>Token Transaction Count: {{ .TokenTransactionCount }}</li>

		<br>

		<li>Account Creation Action Count: {{ .AccountCreationActionCount }}</li>
		<li>Token Action Count: {{ .TokenActionCount }}</li>
		<li>Token Notif Count: {{ .TokenNotifCount }}</li>

		<br>

		<li>Elapsed Time: {{ .ElapsedTime }}</li>
	</ul>
`
