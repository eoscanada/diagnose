package eos

import (
	"fmt"
	"html/template"
	"io"
	"math"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/eoscanada/diagnose/renderer"

	"github.com/eoscanada/bstream"
	pbbstream "github.com/eoscanada/bstream/pb/dfuse/bstream/v1"
	pbdeos "github.com/eoscanada/bstream/pb/dfuse/codecs/deos"
	"github.com/eoscanada/derr"
	"github.com/eoscanada/dhttp"
	"github.com/eoscanada/validator"
	"go.uber.org/zap"
)

var statsTopAccountsTemplate *template.Template
var statsTopAccountsResultTemplate *template.Template
var doingStatsTopAccounts bool

func init() {
	var err error

	statsTopAccountsTemplate, err = template.New("stats").Parse(statsTopAccountsTemplateContent)
	derr.ErrorCheck("unable to create template stats", err)

	statsTopAccountsResultTemplate, err = template.New("stats_results").Parse(statsTopAccountsResultsTemplateContent)
	derr.ErrorCheck("unable to create template stats_results", err)
}

func (e *Diagnose) verifyStatsTopAccounts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if doingStatsTopAccounts {
		dhttp.WriteHTML(ctx, w, alreadyRunningTemplate, nil)
		return
	}

	doingStatsTopAccounts = true
	defer func() { doingStatsTopAccounts = false }()

	request := &statsTopAccountsRequest{}
	err := dhttp.ExtractRequest(ctx, r, request, dhttp.NewRequestValidator(statsTopAccountsRequestValidationRules))

	zlog.Info("request", zap.Any("request", request))

	if err != nil {
		dhttp.WriteError(ctx, w, derr.Wrap(err, "invalid request"))
		return
	}

	startBlockNum, _ := strconv.ParseUint(request.StartBlockNum, 10, 64)
	stopBlockNum, _ := strconv.ParseUint(request.StopBlockNum, 10, 64)

	dhttp.WriteHTML(ctx, w, statsTopAccountsTemplate, map[string]interface{}{
		"StartBlock": startBlockNum,
		"StopBlock":  stopBlockNum,
	})
	renderer.FlushWriter(w)

	stats := &StatsTopAccountsBlockHandler{
		writer:        w,
		startBlockNum: startBlockNum,
		stopBlockNum:  stopBlockNum,
		logInterval:   logIntervalInBlock(startBlockNum, stopBlockNum),

		ActionCountMap:      map[string]uint64{},
		TransactionCountMap: map[string]uint64{},
		ResultLineMap:       map[string]string{},

		MaxActionCount: 0,
		MinActionCount: math.MaxUint64,

		MaxTransactionCount: 0,
		MinTransactionCount: math.MaxUint64,

		LastHour: -1,
	}

	source := bstream.NewFileSource(
		pbbstream.Protocol_EOS,
		e.BlocksStore,
		startBlockNum,
		int(e.ParallelDownloadCount),
		nil,
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

	dhttp.WriteHTML(ctx, w, statsTopAccountsResultTemplate, stats)
}

type statsTopAccountsRequest struct {
	StartBlockNum string `schema:"start_block"`
	StopBlockNum  string `schema:"stop_block"`
}

var statsTopAccountsRequestValidationRules = validator.Rules{
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

type StatsTopAccountsBlockHandler struct {
	writer http.ResponseWriter

	BlockCount          uint64
	ActionCountMap      map[string]uint64
	TransactionCountMap map[string]uint64
	ResultLineMap       map[string]string
	OrderedResults      []string

	SumActionCount     uint64
	AverageActionCount uint64
	MaxActionCount     uint64
	MinActionCount     uint64

	SumTransactionCount     uint64
	AverageTransactionCount uint64
	MaxTransactionCount     uint64
	MinTransactionCount     uint64

	ElapsedTime time.Duration
	LastHour    int64

	startBlockNum               uint64
	stopBlockNum                uint64
	highestPreprocessedBlockNum uint64
	highestProcessBlockNum      uint64

	timeStart   time.Time
	logInterval int
}

func (s *StatsTopAccountsBlockHandler) handleBlock(block *bstream.Block, obj interface{}) error {
	blockNum := uint64(block.Num())
	if blockNum > s.stopBlockNum {
		return io.EOF
	}

	if blockNum > s.highestProcessBlockNum {
		s.highestProcessBlockNum = blockNum
	}

	if s.timeStart.IsZero() {
		s.timeStart = time.Now()
		renderer.PutLine(s.writer, "<small>Processing blocks...</small><br>")
	}

	s.accumulateBlockStats(block)

	if s.logInterval > 0 && blockNum%uint64(s.logInterval) == 0 {
		renderer.PutLine(s.writer, "<small>At block %d</small><br>", blockNum)
		s.logStats(false)
	}

	if blockNum == s.stopBlockNum {
		s.ElapsedTime = time.Since(s.timeStart)
		s.logStats(true)
	}

	return nil
}

func (s *StatsTopAccountsBlockHandler) accumulateBlockStats(block *bstream.Block) error {
	s.BlockCount++

	blk := block.ToNative().(*pbdeos.Block)
	for _, trx := range blk.GetTransactionTraces() {
		seenActionsMap := map[string]bool{}

		for _, action := range trx.ActionTraces {
			receiver := action.Receiver
			topAccountAuthorizer := findAuthorizingTopAccount(action)

			if topAccounts[receiver] {
				seenActionsMap[receiver] = true
				s.recordAccountActivity(receiver)
				continue
			}

			if topAccountAuthorizer != "" {
				seenActionsMap[topAccountAuthorizer] = true
				s.recordAccountActivity(topAccountAuthorizer)
				continue
			}
		}

		for seenAccount := range seenActionsMap {
			s.SumTransactionCount++

			transactionCount := s.TransactionCountMap[seenAccount]
			transactionCount++

			if transactionCount > s.MaxTransactionCount {
				s.MaxTransactionCount = transactionCount
			}

			if transactionCount < s.MinTransactionCount {
				s.MinTransactionCount = transactionCount
			}

			s.TransactionCountMap[seenAccount] = transactionCount
		}
	}

	return nil
}

func (s *StatsTopAccountsBlockHandler) recordAccountActivity(account string) {
	s.SumActionCount++

	actionCount := s.ActionCountMap[account]
	actionCount++

	if actionCount > s.MaxActionCount {
		s.MaxActionCount = actionCount
	}

	if actionCount < s.MinActionCount {
		s.MinActionCount = actionCount
	}

	s.ActionCountMap[account] = actionCount
}

func findAuthorizingTopAccount(actionTrace *pbdeos.ActionTrace) string {
	actors := actionTrace.GetData("act.authorization.#.actor")
	if !actors.IsArray() {
		return ""
	}

	for _, actorResult := range actors.Array() {
		actor := actorResult.String()
		if topAccounts[actor] {
			return actor
		}
	}

	return ""
}

func (s *StatsTopAccountsBlockHandler) computeOngoingStats(last bool) {
	s.AverageActionCount = uint64(float64(s.SumActionCount) / float64(len(s.ActionCountMap)))
	s.AverageTransactionCount = uint64(float64(s.SumTransactionCount) / float64(len(s.TransactionCountMap)))

	for account, transactionCount := range s.TransactionCountMap {
		s.ResultLineMap[account] = fmt.Sprintf("%d documents (%d actions)", transactionCount, s.ActionCountMap[account])
	}

	if last {
		var orderedAccounts []string
		for account := range s.TransactionCountMap {
			orderedAccounts = append(orderedAccounts, account)
		}

		sort.Slice(orderedAccounts, func(i int, j int) bool {
			left := orderedAccounts[i]
			right := orderedAccounts[j]

			return s.TransactionCountMap[right] < s.TransactionCountMap[left]
		})

		s.OrderedResults = make([]string, len(orderedAccounts))
		for i, account := range orderedAccounts {
			s.OrderedResults[i] = fmt.Sprintf("%s %d documents (%d actions)", account, s.TransactionCountMap[account], s.ActionCountMap[account])
		}
	}
}

func (s *StatsTopAccountsBlockHandler) logStats(last bool) {
	s.computeOngoingStats(last)

	fields := []zap.Field{
		zap.Uint64("block_count", s.BlockCount),
		zap.Uint64("active_account_count", uint64(len(s.ResultLineMap))),
		zap.Uint64("sum_action_count", s.SumActionCount),
		zap.Uint64("sum_transaction_count", s.SumTransactionCount),
		zap.Uint64("average_action_count", s.AverageActionCount),
		zap.Uint64("average_transaction_count", s.AverageTransactionCount),
		zap.Uint64("min_action_count", s.MinActionCount),
		zap.Uint64("max_action_count", s.MaxActionCount),
		zap.Uint64("min_transaction_count", s.MinTransactionCount),
		zap.Uint64("max_transaction_count", s.MaxTransactionCount),
	}

	elapsedTime := s.ElapsedTime
	if uint64(elapsedTime) == 0 {
		elapsedTime = time.Since(s.timeStart)
	}

	fields = append(fields, zap.Duration("elapsed_time", elapsedTime))

	zlog.Info("ongoing stats", fields...)

	if last || int64(elapsedTime.Hours()) > s.LastHour {
		s.LastHour = int64(elapsedTime.Hours())
		for account, resultLine := range s.ResultLineMap {
			zlog.Info(account+"_stats", zap.String("line", resultLine))
		}
	}
}

var statsTopAccountsTemplateContent = `
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

var statsTopAccountsResultsTemplateContent = `
	<h3>Stats Results</h3>
	<ul>
		<li>Block Count: {{ .BlockCount }}</li>
		<li>Active Account Count: {{ len .ResultLineMap }}</li>
		<li>Total Documents: {{ .SumTransactionCount }} ({{ .SumActionCount }} actions)</li>

		<br>

		<li>Average Action Count: {{ .AverageActionCount }}</li>
		<li>Average Transaction Count: {{ .AverageTransactionCount }}</li>

		<br>

		<li>Min Action Count: {{ .MinActionCount }}</li>
		<li>Max Action Count: {{ .MaxActionCount }}</li>

		<br>

		<li>Min Transaction Count: {{ .MinTransactionCount }}</li>
		<li>Max Transaction Count: {{ .MaxTransactionCount }}</li>

		<br>

		{{ range $i, $result := .OrderedResults }}
			<li>{{ $result }}</li>
		{{ end }}

		<li>Elapsed Time: {{ .ElapsedTime }}</li>
	</ul>
`

var topAccounts = map[string]bool{
	"eosio.stake":  true,
	"eosio.rex":    true,
	"binancecold1": true,
	"fepxecwzm41t": true,
	"vuniyuoxoeub": true,
	"osd1zfr4xn2a": true,
	"eospstotoken": true,
	"bittrexacct1": true,
	"winterishere": true,
	"eosio.saving": true,
	"okbtothemoon": true,
	"jts51xdlwoey": true,
	"vwgumvhjzkvn": true,
	"bpvoter.one":  true,
	"bitfinexcw55": true,
	"okexoffiline": true,
	"gy2dsnjugige": true,
	"gateiowallet": true,
	"b1":           true,
	"eosio.ram":    true,
	"dunamueoshot": true,
	"ycheosforzb4": true,
	"giytemzyhege": true,
	"krakenkraken": true,
	"poloniexeos1": true,
	"zbeoscharge1": true,
	"gm3tombuguge": true,
	"gzth5wlzwrth": true,
	"a2so5mzgelqu": true,
	"eosio.ramfee": true,
	"gq4dinigenes": true,
	"gu4dgojygege": true,
	"heztanrqgene": true,
	"fcoineosio11": true,
	"eosio.unregd": true,
	"bitfinexdep1": true,
	"ha2tanbqg4ge": true,
	"turkmanistan": true,
	"ljnop45jo3fi": true,
	"eosio.names":  true,
	"iav5zsmua15q": true,
	"athenafundom": true,
	"haytmmjrgige": true,
	"pxneosincome": true,
	"wotokencoldw": true,
	"otcbtcdotcom": true,
	"guydsnbsgige": true,
	"csccsccsccs1": true,
	"4e2uzhb5mo1f": true,
	"nbyeoswallet": true,
	"d1e5ujvg2z5k": true,
	"wohotwallet5": true,
	"eoswithmixin": true,
	"bithumbsend1": true,
	"eos32signhw1": true,
	"ppllgg122412": true,
	"hitbtcpayins": true,
	"abcdabcd2233": true,
	"sbcount33555": true,
	"gq3timbugene": true,
	"arbyiloveyou": true,
	"coinonekorea": true,
	"eosedeiofund": true,
	"plyeusxtkmav": true,
	"haztkmbvgqge": true,
	"ge2tsmbzg4ge": true,
	"hazdmmrvgige": true,
	"coinonewallt": true,
	"berithwallet": true,
	"hffyr5dfv3gk": true,
	"ta5wjbhgakay": true,
	"xjcg3ijovmls": true,
	"eoshoowallet": true,
	"sharethepzza": true,
	"eoszq3333333": true,
	"darwinism111": true,
	"g4ydgmjug4ge": true,
	"gopaxdeposit": true,
	"eosjianglban": true,
	"renrenbit135": true,
	"wallet4bixin": true,
	"gqzdknbsgage": true,
	"wotokencleos": true,
	"gm3tknrrguge": true,
	"zbeosforsend": true,
	"wallthoteos1": true,
	"chengcheng51": true,
	"mcdkcoidjowe": true,
	"askbankofeos": true,
	"ha2tcojygene": true,
	"dskjzyhdzw12": true,
	"kickmyfatass": true,
	"wolf11223344": true,
	"2cn4srotut3s": true,
	"wbxyejt5ezd3": true,
	"whaleextrust": true,
	"schussboomer": true,
	"youbankcleos": true,
	"radardeposit": true,
	"gmzdcobrg4ge": true,
	"zhangkai1324": true,
	"guyteojugage": true,
	"bitfinexeos1": true,
	"txjjzyzqbx12": true,
	"geztamjtguge": true,
	"itxuhotcleos": true,
	"jrzrhjz12345": true,
	"eosedeiocost": true,
	"pxnramtrader": true,
	"bcexcawallet": true,
	"gm2tqnjygyge": true,
	"bbbbbbb11111": true,
	"ainidaolao35": true,
	"bitzdeposit1": true,
	"eosbiggame21": true,
	"bitforexcoin": true,
	"q1tjyq5ts1s1": true,
	"worldfoxprin": true,
	"guztoojqgege": true,
	"chhbsvgu3433": true,
	"biboxexdepst": true,
	"wmyqcgzdq123": true,
	"eoszxcvbnmaa": true,
	"xxtmqyb3srce": true,
	"craigspys211": true,
	"bc1etheosua1": true,
	"heydcmrqgige": true,
	"bxcoldwallet": true,
	"eos4delegate": true,
	"fundboxeosio": true,
	"ov2epx1hverx": true,
	"kj1ffmdwgdyy": true,
	"g42tombug4ge": true,
	"eosplaybanks": true,
	"bkexeoscold1": true,
	"ge2danrxhage": true,
	"gq2dembvhege": true,
	"gq2tiobzguge": true,
	"hsgu55223554": true,
	"qiangqiang33": true,
	"fiz3wepjoykn": true,
	"w3utiu51jf5m": true,
	"brown3215435": true,
	"paribuwallet": true,
	"jwjw11111111": true,
	"changjiang35": true,
	"chainupsaas2": true,
	"knjkt3y5eqhb": true,
	"hyperpayeos3": true,
	"sensaysensay": true,
	"gi4tkmjwgige": true,
	"bybitzhou.tp": true,
	"jim23541234h": true,
	"jimk12345332": true,
	"ydfrzzxyecfy": true,
	"wallthoteos2": true,
	"eosofdeposit": true,
	"sosbp1111111": true,
	"brown3254123": true,
	"aofexdeposit": true,
	"songyan12345": true,
	"actresscarda": true,
	"brownh321542": true,
	"ge3tiobvgqge": true,
	"gy4dcmbvhege": true,
	"wallthoteos4": true,
	"brown3254213": true,
	"gi2demzsgyge": true,
	"atmitxeosc11": true,
	"atmitxeosc15": true,
	"atmitxeosc21": true,
	"gingin111112": true,
	"coincolasave": true,
	"gq4tonbrhege": true,
	"wallethoteos": true,
	"cobowalletio": true,
	"gm2tsmztgene": true,
	"lqapwalletc5": true,
	"mfmmccoygrlz": true,
	"kucoinsender": true,
	"geztaobygene": true,
	"findexeosabc": true,
	"eosfinance11": true,
	"hasoafgctbn3": true,
	"kvj3mw3ps1c3": true,
	"gq4tkmbsgqge": true,
	"eoslcoinling": true,
	"ge4tmmbvgage": true,
	"gezdqmzqgege": true,
	"atmitxeosc12": true,
	"coinexcoldn1": true,
	"eosplaybrand": true,
	"1515ezuptent": true,
	"miochain5555": true,
	"mikewwww1234": true,
	"gm3dgojvgmge": true,
	"wallthoteos5": true,
	"gurguruchusu": true,
	"athenarefuge": true,
	"ypnlfxyhbcwg": true,
	"ddjghss12233": true,
	"btcbtcbtc542": true,
	"gi2tsnbtguge": true,
	"gm4tinjvgene": true,
	"wang14411441": true,
	"eosbybit1111": true,
	"atmitxeosc13": true,
	"atmitxeosc14": true,
	"sunchaoeos55": true,
	"hbnderqwnvop": true,
	"blockcitygxc": true,
	"xiaotian5555": true,
	"meliodassama": true,
	"allenxingab2": true,
	"zd3bl2abauwe": true,
	"guytcmbqgege": true,
	"coind1adress": true,
	"he2dkmrrgene": true,
	"gqytqnjxg4ge": true,
	"sinkleaf3eos": true,
	"eoswallet123": true,
	"eosfinqaz222": true,
	"gqztamrzgyge": true,
	"sss111zzz111": true,
	"coinspoteos1": true,
	"bitmaxaddrz1": true,
	"gmztqnjvgage": true,
	"athenalinker": true,
	"shysh2333.tp": true,
	"eoschihuahua": true,
	"gi3denzxhege": true,
	"ge4dgnrqhege": true,
	"dalishuilong": true,
	"coinbenedepo": true,
	"kryptosystem": true,
	"gi2dsnzrgmge": true,
	"bweosdeposit": true,
	"atmitxeosc22": true,
	"liqinliyike5": true,
	"kucoincenter": true,
	"gqytenjwgqge": true,
	"eoshashhouse": true,
	"gu4domjxgmge": true,
	"blocktradeos": true,
	"yexuede22331": true,
	"hegenesis111": true,
	"wallthoteos3": true,
	"gyztenzsgqge": true,
	"giydknygenes": true,
	"readythecat1": true,
	"djbwallet123": true,
	"eosmoneyball": true,
	"eosioeosios3": true,
	"sanguoyanyi5": true,
	"mbaexdeposit": true,
	"xsz111xsz333": true,
	"jiexuprimary": true,
	"gu3diojyhage": true,
	"eoszxywjh311": true,
	"manoliwallet": true,
	"supermangaga": true,
	"rdscgfoztd1d": true,
	"12345aassdda": true,
	"dlwjtys1hllp": true,
	"guydmmryhage": true,
	"gi4tqmzwhege": true,
	"5iyewpv32.tp": true,
	"zos2eoscnvrt": true,
	"1me5yyisko4e": true,
	"exxeoscharge": true,
	"guxiaoweieos": true,
	"blockcityout": true,
	"aybihmn2ofxu": true,
	"eosantpoolbp": true,
	"amad3jgoux3l": true,
	"red5stndngby": true,
	"viceostokeni": true,
	"cbflulmceopi": true,
	"xku2epmcpevk": true,
	"freewalletin": true,
	"giytqobrgmge": true,
	"gy2teobyhege": true,
	"ha3tmnjvgige": true,
	"ah25u2crjgkl": true,
	"grunter21343": true,
	"giytamjygage": true,
	"ge2domzqgige": true,
	"btceos123542": true,
	"coinwwallet1": true,
	"eosbetbank11": true,
	"catdividends": true,
	"jxxyhljxxyhl": true,
	"gq2tmmrugage": true,
	"iamhumaneosu": true,
	"eostodam1111": true,
	"potus1111111": true,
	"cscscscs1111": true,
	"freetokeneos": true,
	"newdexiocold": true,
	"ncnjorxmid41": true,
	"myeosaddress": true,
	"coinsuperpos": true,
	"cateatpeanut": true,
	"ge2tknjugqge": true,
	"coinpaymteos": true,
	"jiaozi313dao": true,
	"genesisfund2": true,
	"eosbiggame55": true,
	"genesiscap11": true,
	"betdicetasks": true,
	"gq4tembvgqge": true,
	"wrwr11111111": true,
	"gyztknbtgyge": true,
	"gue3nfibst11": true,
	"qzwbi2m1znes": true,
	"ge3tgojzhage": true,
	"eosprospertk": true,
	"zhl134134zdl": true,
	"gy2tiobzgmge": true,
	"gu3tmnjtgmge": true,
	"eosiswallets": true,
	"big.one":      true,
	"eos5parkowen": true,
	"obe5ugevtvbe": true,
	"huhuashi2314": true,
	"ubebxakigid2": true,
	"bitvavostora": true,
	"g4ydamnxhege": true,
	"guzdimbtgage": true,
	"ge4demrrgage": true,
	"bnt2eoscnvrt": true,
	"fanbingbing5": true,
	"gmytiobzgmge": true,
	"eosurubucucu": true,
	"eosfomoplay1": true,
	"blacksheep22": true,
	"d3jfocw12345": true,
	"huhuashi1324": true,
	"huhuashi1234": true,
	"chulanfn2233": true,
	"weatherising": true,
	"yegouyegou12": true,
	"g43timrugege": true,
	"ht4dm1wqoock": true,
	"huhuashi1243": true,
	"woaiwojia112": true,
	"baoqiang4533": true,
	"wangfeng4533": true,
	"eosprivate11": true,
	"huhuashi4321": true,
	"gizdonbzgqge": true,
	"gu4tinbugige": true,
	"gq4dqnzrgage": true,
	"zhaojunfeng3": true,
	"gi3dsojrhage": true,
	"hyperpayeos1": true,
	"2ybksqywiggl": true,
	"biao55555555": true,
	"koineksadres": true,
	"astrohouseos": true,
	"liuhuan45333": true,
	"porschegt3rs": true,
	"soulrich1144": true,
	"viabtccoinex": true,
	"dhbfd1344221": true,
	"magrefa55555": true,
	"1vucvxtdjs4e": true,
	"asdfghjklhai": true,
	"hashqhashqha": true,
	"ycheosforzb3": true,
	"zhongtianxua": true,
	"bhtzqsbhdzqx": true,
	"esgserosdghs": true,
	"hanyusheng22": true,
	"likeqiang333": true,
	"g43danzzhage": true,
	"olololololon": true,
	"halizulibuli": true,
	"eoscoineggim": true,
	"oopaie3yitwp": true,
	"tocashierest": true,
	"he4dcmigenes": true,
	"bitmaxeos123": true,
	"bitebiyitai5": true,
	"eosblockteam": true,
	"sometimes123": true,
	"wcywangyinwy": true,
	"gyzdgnrtguge": true,
	"maxexchwarm2": true,
	"guztqnbshage": true,
	"bptbpt333333": true,
	"orange151515": true,
	"eosbetcldstr": true,
	"qiangsky1111": true,
	"gqytkobqgyge": true,
	"geytamrug4ge": true,
	"guanjm123123": true,
	"geytqnjzgene": true,
	"g44tqojthege": true,
	"geytsojzg4ge": true,
	"brisbane3344": true,
	"emiremir1234": true,
	"linkcoin2out": true,
	"eos4ourdamon": true,
	"yxiliang2233": true,
	"adalawrenloi": true,
	"charlie.x":    true,
	"wiseyewallet": true,
	"hellothere22": true,
	"lixiaohan123": true,
	"newdexpublic": true,
	"tangzuochen5": true,
	"eosioimufasa": true,
	"ge3demzzgege": true,
	"alleoscom1n1": true,
	"gm3tqmbrgige": true,
	"shishaoshuai": true,
	"iehldd3ijnxk": true,
	"g4ydkmrxhege": true,
	"gu1asnrqgmge": true,
	"goc3ktjhxcge": true,
	"gob2ktjhxcge": true,
	"goa1ktjhxcge": true,
	"heztamjugmge": true,
	"yshaopeng223": true,
	"ge3tgmzxhege": true,
	"dox111111111": true,
	"guytkmbrgyge": true,
	"yshungpeng23": true,
	"betdicehouse": true,
	"bcexdeposit1": true,
	"gyztkmjrgege": true,
	"oneswordyoon": true,
	"ge4dmnrwgmge": true,
	"gyytcmagenes": true,
	"librapokey32": true,
	"geztcnrxgyge": true,
	"mxprosperity": true,
	"bgbetwallet1": true,
	"freelinklink": true,
	"guydeojyg4ge": true,
	"n4nxgi1ffidi": true,
	"g4ydomnxhage": true,
	"gu3dgmjvguge": true,
	"atrefxdy1523": true,
	"horuspaycold": true,
	"chainriftcom": true,
	"bitpieeosoo3": true,
	"eoseouldotio": true,
	"lqapwalletd1": true,
	"huobideposit": true,
	"cryptomkteos": true,
	"mycb4h5wnwbe": true,
	"geytcobxgqge": true,
	"ab15131415ab": true,
	"gy3dcojtguge": true,
	"daybitwallet": true,
	"e43edknbnahk": true,
	"selinaselina": true,
	"account11112": true,
	"chintailease": true,
	"hezdsnjsgege": true,
	"haztcmzuguge": true,
	"gezdimbtgige": true,
	"qiasili11111": true,
	"eos4bank4pay": true,
	"gm2demzsg4ge": true,
	"eosplayagent": true,
	"hehongwei123": true,
	"hazdmobygqge": true,
	"fyqeos241315": true,
	"tianmei22335": true,
	"ge3dmnruguge": true,
	"siberiaeos12": true,
	"geydimjugene": true,
	"sinasina1111": true,
	"g44dcnzwgege": true,
	"passage1corp": true,
	"ljyeos112233": true,
	"hytju5321gyi": true,
	"bibox2wallet": true,
	"lmeulpb1vwct": true,
	"zgeosdeposit": true,
	"qhx54h4xvwu5": true,
	"skagitsamish": true,
	"gogogoheaven": true,
	"agdkeih12345": true,
	"hezdanbzgene": true,
	"zhaoyunhong3": true,
	"afan11111111": true,
	"eugtese12345": true,
	"hunseok35211": true,
	"xiaolifeibao": true,
	"chintaidevop": true,
	"qirockbamboo": true,
	"ge4tcnbwgege": true,
	"bidreamalpha": true,
	"gmztenjrgyge": true,
	"gizdambrgene": true,
	"huhuashi1314": true,
	"wxgpssm3wagd": true,
	"gmztkojsgyge": true,
	"ywengpeng223": true,
	"qe5clja4govn": true,
	"321abc123abc": true,
	"kickeosaccnt": true,
	"gy2tqmjwgqge": true,
	"moonthistime": true,
	"haojiaeos111": true,
	"bellewayeo5a": true,
	"gq4tgnrxgige": true,
	"geydenrvgqge": true,
	"lqapmauctis1": true,
	"gqzdamrtgyge": true,
	"ge2tgnztgige": true,
	"eosbybit1234": true,
	"llgforcoacoa": true,
	"hithereatlas": true,
	"wangyunxuqin": true,
	"gu3dcobsgege": true,
	"lbkexwthdraw": true,
	"thekatzenklo": true,
	"myeosfundcom": true,
	"geztmojzgqge": true,
	"gi2daobqgmge": true,
	"ge2tmnbrgyge": true,
	"gy3tkmzrgmge": true,
	"pedroalvarez": true,
	"hyperpaytech": true,
	"gezdmnryg4ge": true,
	"udkg3qo2jri5": true,
	"satoshisatos": true,
	"g44dimrwguge": true,
	"ha4tmmrsgage": true,
	"libertariank": true,
	"mynewchapter": true,
	"u55qev4bufil": true,
	"j43spdgyzpwq": true,
	"geztmmzxgene": true,
	"daibank11111": true,
	"ge2dinrsgege": true,
	"eosbiggame44": true,
	"helloworldwt": true,
	"heytgmrrgyge": true,
	"eoscybexiobp": true,
	"shcoineos123": true,
	"gi2doobwgmge": true,
	"xiaomashitu1": true,
	"eos123123xyq": true,
	"zxcvbnmtgsqz": true,
	"lqapmauctis2": true,
	"g44tsnbtgyge": true,
	"gyztqmzwgqge": true,
	"eosyeeeeeeee": true,
	"gyytsnigenes": true,
	"gqztgmrwgage": true,
	"minuminuminu": true,
	"aqlij23frrya": true,
	"yxizhong2233": true,
	"cannondevops": true,
	"hotbitioeoss": true,
	"g44dcnjzgige": true,
	"gyytqnzvguge": true,
	"qpalzmwoskxg": true,
	"gizdenrxgege": true,
	"hujingchao11": true,
	"gmztomagenes": true,
	"pokerwarbank": true,
	"eosaetherkwy": true,
	"skaredagain2": true,
	"just4cryptoc": true,
	"woexdeposits": true,
	"maxexchange1": true,
	"hezdinbsgege": true,
	"xfaddnttndiz": true,
	"ojbkeostoken": true,
	"scatterfunds": true,
	"ball11223344": true,
	"xiimtoken345": true,
	"ge4timrtgige": true,
	"gm3dkmrxgmge": true,
	"badmanbadman": true,
	"gi4tcnjqhage": true,
	"yifanxiong55": true,
	"gi4dknzqgene": true,
	"exmocleosdep": true,
	"gezdomzrguge": true,
	"bitcoinereo5": true,
	"gmzdemjxgyge": true,
	"ge3domjwgmge": true,
	"ftjh4yfgkued": true,
	"ge3dmmztguge": true,
	"tng2ladmextc": true,
	"geytkmrzgege": true,
	"gongle222222": true,
	"ge2tsojxgige": true,
	"wangsongqing": true,
	"llgtothemoon": true,
	"11111111eos1": true,
	"redhatteryte": true,
	"kinogreciate": true,
	"gu4temjqgyge": true,
	"miaoshan1234": true,
	"gmytknjqgene": true,
	"marqetomaqer": true,
	"tantrictrade": true,
	"g4ztamzqgmge": true,
	"shaozhenimtk": true,
	"victorsazhin": true,
	"g4zdgnbqgene": true,
	"asdfhxcvtwql": true,
	"ge2tgmbqg4ge": true,
	"alubinuresss": true,
	"jlffzp112233": true,
	"hansi1235321": true,
	"lkvtzqocgbxf": true,
	"lifangmylove": true,
	"jeremyrothi2": true,
	"g4ydkojwg4ge": true,
	"djwinjum1234": true,
	"ha4tmmzug4ge": true,
	"ha2tqmbsgene": true,
	"g44toobqhege": true,
	"solomonleder": true,
	"ha3dqnrwgmge": true,
	"w.io":         true,
	"livecoin4eos": true,
	"gu4tonzzgyge": true,
	"hezdanbtgege": true,
	"bihudeposits": true,
	"gy2tamzzhege": true,
	"ethome111111": true,
	"needyougiven": true,
	"gm3tkojwgqge": true,
	"eidnc131zcxz": true,
	"wangyongfeng": true,
	"evagination1": true,
	"cryptotaurus": true,
	"bgdividend11": true,
	"beqjpqx5f3oc": true,
	"bncrchecking": true,
	"biggamevip11": true,
	"exmocleosout": true,
	"indodaxcold1": true,
	"e2sqbbekkvd5": true,
	"mxcexdeposit": true,
	"eoseoszuiren": true,
	"ge3dgmbzgige": true,
	"gy3dinrwgyge": true,
	"tuyint132122": true,
	"shevzhao1125": true,
	"ilelenatonya": true,
	"1kafd4zmjiw4": true,
	"chaojimalili": true,
	"twotwotwo123": true,
	"dermatitises": true,
	"gopaxdaposit": true,
	"indepreserve": true,
	"scottjkasper": true,
	"quanmingyou1": true,
	"gy4dsmzzgege": true,
	"kjbtwoyrbtve": true,
	"eoseoszurich": true,
	"liondani1234": true,
	"epursedotcom": true,
	"oneotc123451": true,
	"eosasia11111": true,
	"oeoy51qdsecq": true,
	"yellowjersey": true,
	"he4dmnrtgene": true,
	"eostitanprod": true,
	"gq2dgobugqge": true,
	"abccexchange": true,
	"21blackjacks": true,
	"fcoindeposit": true,
	"gyytonbug4ge": true,
	"bitmarttopup": true,
	"jamesbrolant": true,
	"eosgameapple": true,
	"eosiotp.bp":   true,
	"pidster51eos": true,
	"flipclubnine": true,
	"shangzhen123": true,
	"ge3dambsg4ge": true,
	"anjing141552": true,
	"slcdgt1lb21a": true,
	"firemanactor": true,
	"charuizhaog1": true,
	"idaxsupereos": true,
	"gezdimbsgyge": true,
	"addeoswallet": true,
	"beijixing111": true,
	"geytemjtgene": true,
	"ge4timbzgege": true,
	"luckymehouse": true,
	"xdapsafebank": true,
	"aecdzfg22335": true,
	"eosweihongli": true,
	"haydimbzgige": true,
	"g44tcmrygage": true,
	"happyflylong": true,
	"eosiojackcom": true,
	"geydqmzwgene": true,
	"gy3dcnjtgyge": true,
	"eosbetstake1": true,
	"wsaking55555": true,
	"gi3tanruhage": true,
	"tg5ykeqsdx4n": true,
	"gy4tqnzsgqge": true,
	"gqydcnbygene": true,
	"u4sys1s4e4ym": true,
	"zb5555555.m":  true,
	"gezdoobwgene": true,
	"dogwalleteos": true,
	"bkexeosouter": true,
	"eoscrwallet2": true,
	"lcchecking11": true,
	"ge3dgnrsgage": true,
	"haytamrxguge": true,
	"bxhotwallet1": true,
	"biggamebuyer": true,
	"hezdambugene": true,
	"ya4sass2dl1c": true,
	"tkvvdwukadps": true,
	"haytcnbvgige": true,
	"fcincgyhvoc2": true,
	"luth12345123": true,
	"sumtoken5555": true,
	"parkbojung12": true,
	"gy2tsnzygene": true,
	"gu2dsnjygege": true,
	"yuzhefu12345": true,
	"eosio.regram": true,
	"geztcmbsgige": true,
	"5pczhtg2j3mm": true,
	"guozhibian12": true,
	"wazirxdoteos": true,
	"chezedude123": true,
	"ahmedmyounis": true,
	"mqcypsatbxeq": true,
	"aakkeos11111": true,
	"kazkoseos325": true,
	"charliealpha": true,
	"gu4dkmjqhage": true,
	"foodgoodrood": true,
	"windknight23": true,
	"g4ydgobthege": true,
	"ge2deojwhage": true,
	"altcoinomysa": true,
	"fucktheword1": true,
	"guytsnjuguge": true,
	"haytknrygige": true,
	"guzdoobygage": true,
	"haytinrzgene": true,
	"ge4tiobzgyge": true,
	"cryptoapples": true,
	"rfinex123455": true,
	"menghaigang1": true,
	"ge4tkmbsguge": true,
	"aexdepositbm": true,
	"stevesledger": true,
	"pieinstantoo": true,
	"sdrobert2eos": true,
	"iloveyoucyou": true,
	"commmmmmmmmm": true,
	"thesis434343": true,
	"sg1111111111": true,
	"gi2dqmryg4ge": true,
	"eosyyywallet": true,
	"koreainbexbp": true,
	"dcoindeposit": true,
	"bizuyunmoney": true,
	"eosaaa555555": true,
	"zuoshangming": true,
	"jiuchiroulin": true,
	"eosbiggame11": true,
	"kimchijuice3": true,
	"wudijimowang": true,
	"coinspoteos2": true,
	"vorobyeboris": true,
	"givegive1111": true,
	"youtong11111": true,
	"vbhdrqedwzxc": true,
	"tadaister125": true,
	"blackfalcon1": true,
	"alisanuresss": true,
	"tradeiotrade": true,
	"gy2tsmbzgene": true,
	"qazwsx412345": true,
	"lulisinyuiot": true,
	"ha3dimjwguge": true,
	"boom1221boom": true,
	"gi3dimygenes": true,
	"gy4tgmygenes": true,
	"wfm111333555": true,
	"seedcapital1": true,
	"bptbptbpt555": true,
	"gizdgnjrg4ge": true,
	"33fohswxzqnw": true,
	"jnfropqbrtor": true,
	"iloveyoumoon": true,
	"richrichtang": true,
	"yfl123321123": true,
	"giztmmbuhege": true,
	"eoswwwyptcom": true,
	"devilmao1111": true,
	"geydonjxhage": true,
	"he2daoagenes": true,
	"dongpeisong1": true,
	"guzdgmbrguge": true,
	"ox13qwertyui": true,
	"gq2tqoigenes": true,
	"ha2tomjsgmge": true,
	"cryptopiaeos": true,
	"ndaxexchange": true,
	"wdcnwdcnwdcn": true,
	"eh1eoswallet": true,
	"tmimicus5555": true,
	"vg2yyep1olze": true,
	"rudexgateway": true,
	"walletiorent": true,
	"gi2tqnjtgyge": true,
	"wowogege1111": true,
	"gi3dimrxgige": true,
	"c11223344551": true,
	"geydimbygige": true,
	"genereosbank": true,
	"angelmao1314": true,
	"ha3dgobugage": true,
	"gu3demrvgmge": true,
	"eosio.vpay":   true,
	"panagiotis11": true,
	"gq4tmnztgqge": true,
	"binancecleos": true,
	"eosfinzyqb11": true,
	"gqzdkoagenes": true,
	"lijiayi11111": true,
	"g44tcnzzhege": true,
	"tt.eos":       true,
	"gooddog12345": true,
	"eosdaoshuo11": true,
	"binancetleos": true,
	"eoslibyou412": true,
	"najieheng.m":  true,
	"gq4dcmqgenes": true,
	"haydmojvg4ge": true,
	"shenhuayan11": true,
	"hazdgojrgmge": true,
	"hj4uhj4mhj4a": true,
	"aabbdd554432": true,
	"gu4tomzygige": true,
	"gyztinjuhage": true,
	"hyebyekadal3": true,
	"guytqmrthege": true,
	"heytgobrgmge": true,
	"gq3tanzvgmge": true,
	"goodday12345": true,
	"eosyemin5sss": true,
	"whrcf1p1gtcn": true,
	"ccd221nuw5nk": true,
	"gmytcmjugage": true,
	"xqym2eqdc1kp": true,
	"eosswedenorg": true,
	"g4ydsnzzgmge": true,
	"binancewleos": true,
	"gi3taobrhage": true,
	"heydonjwgqge": true,
	"ha4tgnjwgqge": true,
	"dengyunlei2b": true,
	"coinonekarea": true,
	"helloeosmoon": true,
	"buldgoldness": true,
	"coingameplay": true,
	"floweringnew": true,
	"wasabieoseos": true,
	"wujiaoshou25": true,
	"ge2teobtguge": true,
	"ha3dcnbzgage": true,
	"aqawaqw12345": true,
	"dragonexabcd": true,
	"fjszzszpx111": true,
	"glsvwrxehhtz": true,
	"rsvliquideos": true,
	"cybexeospool": true,
	"krishnayogiy": true,
	"gmytgobzgege": true,
	"ha3dsmbxgige": true,
	"gmydiobsgyge": true,
	"ge4dqnzwgmge": true,
	"eosioeosios4": true,
	"crypticbit23": true,
	"he4dmnjugene": true,
	"loveyoubigbb": true,
	"gi4tqnjygqge": true,
	"g4ztgobwgqge": true,
	"gu2dmnrzgyge": true,
	"gu2bsnrqgmge": true,
	"oceansailing": true,
	"gi2tiojqgqge": true,
	"utilitymuffn": true,
	"zgzhylcjbssm": true,
	"guztcmbsg4ge": true,
	"e4q3hmateu2e": true,
	"zhangtian125": true,
	"thomas123123": true,
	"helloworldch": true,
	"gu3tiobrhage": true,
	"ge4dgmztgege": true,
	"bitewdeposit": true,
	"fcoinjp2moon": true,
	"eosiomeetone": true,
	"ge4tknzqgege": true,
	"danieleosbag": true,
	"211311jackie": true,
	"longbiteosto": true,
	"biboexchange": true,
	"likunyingeos": true,
	"coffeeman123": true,
	"newstart3155": true,
	"archaeoptryx": true,
	"gq2dmnrrgege": true,
	"nbyyfywallet": true,
	"ha4timztgmge": true,
	"haydagenesis": true,
	"ge4tknbrgyge": true,
	"1aowutong223": true,
	"giytcnbzgene": true,
	"ge4tqmbsgege": true,
	"bitbnsglobal": true,
	"gi2tcnjugige": true,
	"ftbcds3pwnb1": true,
	"ha3dqnjvgene": true,
	"bikicoineos5": true,
	"eos2trillion": true,
	"gy2dmnrwgege": true,
	"gq4tcobzgege": true,
	"eosi521soste": true,
	"songdobomber": true,
	"1v4i4tvsuer2": true,
	"antibreak111": true,
	"tobyxjohnson": true,
	"yanyli4ka123": true,
	"zxcvbzxcvbnn": true,
	"o4tpn1ja1zg3": true,
	"ge3dqmbqhege": true,
	"gu3dknjshage": true,
	"gu4temzzhage": true,
	"gizdanbvgyge": true,
	"cpdaxeoscold": true,
	"daibank12345": true,
	"ge2demjsguge": true,
	"gizdaojygige": true,
	"gyztsmzzgqge": true,
	"gy3dsmztg4ge": true,
	"hedzerfrwrda": true,
	"soigheososha": true,
	"gqztcnzugmge": true,
	"gq3tknrvhege": true,
	"gy2dinbsgage": true,
	"bosstoken.x":  true,
	"guzdonrrguge": true,
	"gu3dkmrzgmge": true,
	"gy3dkmrugmge": true,
	"ranktheworld": true,
	"jiuzixing111": true,
	"yangpeixin12": true,
	"eosreeladmin": true,
	"gmydqnjtgage": true,
	"hitbtcpayout": true,
	"rcajsmctvz2k": true,
	"gy3tgmrsgene": true,
	"hezdinzwgmge": true,
	"eoswinplayio": true,
	"gq4dkmjvhege": true,
	"ha4dcmagenes": true,
	"richgor12345": true,
	"zhangzong521": true,
	"tongzhengeos": true,
	"xexvmvgkz4ne": true,
	"gm3tsobsgyge": true,
	"gqydmmbugqge": true,
	"xuyijing1234": true,
	"gy4toobyhege": true,
	"eoshashbulls": true,
	"yisongchen11": true,
	"eosbihuhot11": true,
	"kcyf2n4ewgep": true,
	"guzdkmjygege": true,
	"geytgnbwgqge": true,
	"gmzdcmzug4ge": true,
	"gu4tkmjsgige": true,
	"mycoinkinbtc": true,
	"gu2tqnbsgige": true,
	"ffcryptoeosz": true,
	"gq2danbqgage": true,
	"g4yteoagenes": true,
	"pwrdbyobolus": true,
	"ge4dkmbuhage": true,
	"gqztanbwgege": true,
	"gi3dqnryhage": true,
}
