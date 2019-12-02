package main

import (
	"net/http"
)

var processingDbHoles bool

//EOS
func (e *Diagnose) DBHoles(w http.ResponseWriter, r *http.Request) {
	//renderer.PutPreambule(w, "Checking block holes in EOSDB")
	//if processingDbHoles {
	//	renderer.PutLine(w, "<h1>Already running, try later</h1>")
	//	return
	//}
	//renderer.PutLine(w, "<p>Bigtable: %s</p>\n", e.KvdbConnectionInfo)
	//
	//processingDbHoles = true
	//defer func() { processingDbHoles = false }()
	//
	//count := int64(0)
	//holeFound := false
	//started := false
	//previousNum := int64(0)
	//previousValidNum := int64(0)
	//
	//startTime := time.Now()
	//batchStartTime := time.Now()
	//
	//blocksTable := e.EOSdb.Blocks
	//
	//// You can test on a lower range with `bt.NewRange("ff76abbf", "ff76abcf")`
	//err := blocksTable.BaseTable.ReadRows(r.Context(), bt.InfiniteRange(""), func(row bt.Row) bool {
	//	count++
	//
	//	num := int64(math.MaxUint32 - kvdb.BlockNum(row.Key()))
	//
	//	isValid := utils.HasAllColumns(row, blocksTable.ColBlock, blocksTable.ColTransactionRefs, blocksTable.ColTransactionTraceRefs, blocksTable.ColMetaWritten, blocksTable.ColMetaIrreversible)
	//
	//	if !started {
	//		previousNum = num + 1
	//		previousValidNum = num + 1
	//		batchStartTime = time.Now()
	//
	//		renderer.PutLine(w, "<p><strong>Start block %d</strong></p>\n", num)
	//	}
	//
	//	difference := previousNum - num
	//	differenceInvalid := previousValidNum - num
	//
	//	if difference > 1 && started {
	//		holeFound = true
	//		renderer.PutLine(w, "<p><strong>Found block hole: [%d, %d]</strong></p>\n", num+1, previousNum-1)
	//	}
	//
	//	if differenceInvalid > 1 && started && isValid {
	//		holeFound = true
	//		renderer.PutLine(w, "<p><strong>Found missing column(s) hole: [%d, %d]</strong></p>\n", num+1, previousValidNum-1)
	//	}
	//
	//	previousNum = num
	//	if isValid {
	//		previousValidNum = num
	//	}
	//
	//	if count%200000 == 0 {
	//		now := time.Now()
	//		renderer.PutLine(w, "<p>200K rows read @ #%d (batch %s, total %s) ...</p>\n", num, now.Sub(batchStartTime), now.Sub(startTime))
	//		batchStartTime = time.Now()
	//	}
	//
	//	started = true
	//
	//	return true
	//}, bt.RowFilter(bt.StripValueFilter()))
	//
	//differenceInvalid := previousValidNum - previousNum
	//if differenceInvalid > 1 && started {
	//	holeFound = true
	//	renderer.PutLine(w, "<p><strong>Found missing column(s) hole: [%d, %d]</strong></p>\n", previousNum, previousValidNum-1)
	//}
	//
	//if err != nil {
	//	renderer.PutLine(w, "<p><strong>Error: %s</strong></p>\n", err.Error())
	//	return
	//}
	//
	//if !holeFound {
	//	renderer.PutLine(w, "<p><strong>NO HOLE FOUND!</strong></p>\n")
	//}
	//
	//renderer.PutLine(w, "<p><strong>Completed at block num %d (%d blocks seen) in %s</strong></p>\n", previousNum, count, time.Now().Sub(startTime))
}

//ETH
//func (e *Diagnose) DBHoles(w http.ResponseWriter, r *http.Request) {
//	//renderer.PutPreambule(w, "Checking block holes in ETHDB")
//	//if processingDBHoles {
//	//	renderer.PutLine(w, "<h1>Already running, try later</h1>")
//	//	return
//	//}
//	//renderer.PutLine(w, "<p>Bigtable: %s</p>\n", e.KvdbConnectionInfo)
//	//
//	//processingDBHoles = true
//	//defer func() { processingDBHoles = false }()
//	//
//	//count := int64(0)
//	//holeFound := false
//	//started := false
//	//previousNum := uint64(0)
//	//previousValidNum := uint64(0)
//	//
//	//startTime := time.Now()
//	//batchStartTime := time.Now()
//	//
//	//blocksTable := e.ETHdb.Blocks
//	//
//	//// You can test on a lower range with `bt.NewRange("ff76abbf", "ff76abcf")`
//	//err := blocksTable.BaseTable.ReadRows(r.Context(), bt.InfiniteRange("blkn:"), func(row bt.Row) bool {
//	//	count++
//	//
//	//	num, _, e := ethdb.Keys.ReadBlockNum(row.Key())
//	//	if e != nil {
//	//		renderer.PutLine(w, "<p><strong>Error: %s</strong></p>\n", e.Error())
//	//		return false
//	//	}
//	//
//	//	isValid := utils.HasAllColumns(row, blocksTable.ColHeaderProto, blocksTable.ColMetaIrreversible, blocksTable.ColMetaMapping, blocksTable.ColMetaWritten, blocksTable.ColTrxRefsProto, blocksTable.ColUnclesProto)
//	//
//	//	if !started {
//	//		previousNum = num + 1
//	//		previousValidNum = num + 1
//	//		batchStartTime = time.Now()
//	//
//	//		renderer.PutLine(w, "<p><strong>Start block %d</strong></p>\n", num)
//	//	}
//	//
//	//	difference := previousNum - num
//	//	differenceInvalid := previousValidNum - num
//	//
//	//	if difference > 1 && started {
//	//		holeFound = true
//	//		renderer.PutLine(w, "<p><strong>Found block hole: [%d, %d]</strong></p>\n", num+1, previousNum-1)
//	//	}
//	//
//	//	if differenceInvalid > 1 && started && isValid {
//	//		holeFound = true
//	//		renderer.PutLine(w, "<p><strong>Found missing column(s) hole: [%d, %d]</strong></p>\n", num+1, previousValidNum-1)
//	//	}
//	//
//	//	previousNum = num
//	//	if isValid {
//	//		previousValidNum = num
//	//	}
//	//
//	//	if count%200000 == 0 {
//	//		now := time.Now()
//	//		renderer.PutLine(w, "<p>200K rows read @ #%d (batch %s, total %s) ...</p>\n", num, now.Sub(batchStartTime), now.Sub(startTime))
//	//		batchStartTime = time.Now()
//	//	}
//	//
//	//	started = true
//	//
//	//	return true
//	//}, bt.RowFilter(bt.StripValueFilter()))
//	//
//	//differenceInvalid := previousValidNum - previousNum
//	//if differenceInvalid > 1 && started {
//	//	holeFound = true
//	//	renderer.PutLine(w, "<p><strong>Found missing column(s) hole: [%d, %d]</strong></p>\n", previousNum, previousValidNum-1)
//	//}
//	//
//	//if err != nil {
//	//	renderer.PutLine(w, "<p><strong>Error: %s</strong></p>\n", err.Error())
//	//	return
//	//}
//	//
//	//if !holeFound {
//	//	renderer.PutLine(w, "<p><strong>NO HOLE FOUND!</strong></p>\n")
//	//}
//	//
//	//renderer.PutLine(w, "<p><strong>Completed at block num %d (%d blocks seen) in %s</strong></p>\n", previousNum, count, time.Now().Sub(startTime))
//}
