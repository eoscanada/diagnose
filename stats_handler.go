package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/eoscanada/bstream"
	pbbstream "github.com/eoscanada/bstream/pb/dfuse/bstream/v1"
	pbdeos "github.com/eoscanada/bstream/pb/dfuse/codecs/deos"
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
	flushWriter(w)

	stats := &StatsBlockHandler{
		writer:        w,
		startBlockNum: startBlockNum,
		stopBlockNum:  stopBlockNum,
		logInterval:   logIntervalInBlock(startBlockNum, stopBlockNum),
	}

	source := bstream.NewFileSource(
		pbbstream.Protocol_EOS,
		d.blocksStore,
		startBlockNum,
		int(d.parallelDownloadCount),
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
	delta := int64(stopBlock) - int64(startBlock)
	if delta < 10000 {
		return 1000
	}

	if delta < 100000 {
		return 10000
	}

	if delta < 1000000 {
		return 100000
	}

	return 250000
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

	startBlockNum               uint64
	stopBlockNum                uint64
	highestPreprocessedBlockNum uint64
	highestProcessBlockNum      uint64

	timeStart   time.Time
	logInterval int
}

func (s *StatsBlockHandler) preprocessBlock(block *bstream.Block) (interface{}, error) {
	blockNum := block.Num()
	if blockNum < s.startBlockNum || blockNum > s.stopBlockNum {
		return nil, nil
	}

	if blockNum > s.highestPreprocessedBlockNum {
		s.highestPreprocessedBlockNum = blockNum
	}

	s.BlockCount++

	blk := block.ToNative().(*pbdeos.Block)
	for _, trx := range blk.GetTransactionTraces() {
		s.TransactionCount++

		seenAccountCreationAction := false
		seenTokenTransactionAction := false
		for _, action := range trx.ActionTraces {
			s.ActionCount++

			receiver := action.Receiver
			account := action.Account()
			actionName := action.Name()
			isNotif := receiver != account

			if receiver == "eosio" && account == "eosio" && actionName == "newaccount" {
				s.AccountCreationActionCount++
				seenAccountCreationAction = true
			}

			if actionName == "transfer" || actionName == "close" || actionName == "issue" || actionName == "retire" {
				seenTokenTransactionAction = true

				if isNotif {
					s.TokenNotifCount++
				} else {
					s.TokenActionCount++
				}
			}
		}

		if seenAccountCreationAction {
			s.AccountCreationTransactionCount++
		}

		if seenTokenTransactionAction {
			s.TokenTransactionCount++
		}
	}

	return nil, nil
}

func (s *StatsBlockHandler) handleBlock(block *bstream.Block, obj interface{}) error {
	blockNum := uint64(block.Num())
	if blockNum > s.stopBlockNum {
		return io.EOF
	}

	if blockNum > s.highestProcessBlockNum {
		s.highestProcessBlockNum = blockNum
	}

	if s.timeStart.IsZero() {
		s.timeStart = time.Now()
		putLine(s.writer, "<small>Processing blocks...</small><br>")
	}

	if s.logInterval > 0 && blockNum%uint64(s.logInterval) == 0 {
		putLine(s.writer, "<small>At block %d</small><br>", blockNum)
		s.logStats()
	}

	if blockNum == s.stopBlockNum {
		s.ElapsedTime = time.Since(s.timeStart)
		s.logStats()
	}

	return nil
}

func (s *StatsBlockHandler) logStats() {
	fields := []zap.Field{
		zap.Uint64("action_count", s.ActionCount),
		zap.Uint64("transaction_count", s.TransactionCount),
		zap.Uint64("block_count", s.BlockCount),
		zap.Uint64("token_action_count", s.TokenActionCount),
		zap.Uint64("token_notif_count", s.TokenNotifCount),
		zap.Uint64("token_transaction_count", s.TokenTransactionCount),
		zap.Uint64("account_creation_action_count", s.AccountCreationActionCount),
		zap.Uint64("account_creation_transaction_count", s.AccountCreationTransactionCount),
		zap.Uint64("highest_preprocessed_block_num", s.highestPreprocessedBlockNum),
		zap.Uint64("highest_process_block_num", s.highestProcessBlockNum),
	}

	elapsedTime := s.ElapsedTime
	if uint64(elapsedTime) == 0 {
		elapsedTime = time.Since(s.timeStart)
	}

	fields = append(fields, zap.Duration("elapsed_time", elapsedTime))

	zlog.Info("ongoing stats", fields...)
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
