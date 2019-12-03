package main

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"time"

	bt "cloud.google.com/go/bigtable"
	"github.com/eoscanada/diagnose/utils"
	"github.com/eoscanada/kvdb"
	"github.com/eoscanada/kvdb/ethdb"
	"go.uber.org/zap"
)

func (d *Diagnose) EOSKVDBBlocks(w http.ResponseWriter, req *http.Request) {
	zlog.Info("diagnose - EOS  - KVDB Block Hole Checker",
		zap.String("kvdb_connection_info", d.KvdbConnectionInfo))

	conn, err := d.upgrader.Upgrade(w, req, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	ctx := req.Context()

	count := int64(0)
	started := false
	previousNum := int64(0)
	//previousValidNum := int64(0)
	batchHighBlockNum := int64(0)
	currentBlockNum := int64(0)

	startTime := time.Now()
	batchStartTime := time.Now()

	blocksTable := d.EOSdb.Blocks

	go websocketRead(conn)

	blocksTable.BaseTable.ReadRows(ctx, bt.InfiniteRange(""), func(row bt.Row) bool {
		count++

		currentBlockNum = int64(math.MaxUint32 - kvdb.BlockNum(row.Key()))

		if !started {
			previousNum = currentBlockNum + 1
			//		previousValidNum = currentBlockNum + 1
			batchStartTime = time.Now()
			batchHighBlockNum = currentBlockNum
		}

		difference := previousNum - currentBlockNum

		if difference > 1 && started {
			msg := fmt.Sprintf("%d rows read\n", (uint32(batchHighBlockNum) - uint32(currentBlockNum+1)))
			_ = sendMessage(conn, &BlockRange{
				StarBlock: uint32(currentBlockNum + 1),
				EndBlock:  uint32(batchHighBlockNum),
				Message:   "",
				Status:    BlockRangeStatusValid,
			})
			msg = fmt.Sprintf("Found block hole %d rows\n", (uint32(previousNum-1) - uint32(currentBlockNum+1)))
			_ = sendMessage(conn, &BlockRange{
				StarBlock: uint32(currentBlockNum + 1),
				EndBlock:  uint32(previousNum - 1),
				Message:   msg,
				Status:    BlockRangeStatusHole,
			})
			batchHighBlockNum = currentBlockNum
		}

		previousNum = currentBlockNum

		if count%200000 == 0 {
			now := time.Now()
			msg := fmt.Sprintf("%d rows read (batch %s, total %s)\n", (uint32(batchHighBlockNum) - uint32(currentBlockNum)), now.Sub(batchStartTime), now.Sub(startTime))
			_ = sendMessage(conn, &BlockRange{
				StarBlock: uint32(currentBlockNum),
				EndBlock:  uint32(batchHighBlockNum),
				Message:   msg,
				Status:    BlockRangeStatusValid,
			})
			batchStartTime = time.Now()
			batchHighBlockNum = (currentBlockNum - 1)
		}

		started = true

		return true
	}, bt.RowFilter(bt.StripValueFilter()))
	now := time.Now()
	msg := fmt.Sprintf("%d rows read (batch %s, total %s)\n", (uint32(batchHighBlockNum) - uint32(currentBlockNum)), now.Sub(batchStartTime), now.Sub(startTime))
	_ = sendMessage(conn, &BlockRange{
		StarBlock: uint32(currentBlockNum),
		EndBlock:  uint32(batchHighBlockNum),
		Message:   msg,
		Status:    BlockRangeStatusValid,
	})
	zlog.Info("diagnose - EOS  - KVDB Block Hole Checker - Completed")
}

func (d *Diagnose) EOSKVDBBlocksValidation(w http.ResponseWriter, req *http.Request) {
	zlog.Info("diagnose - EOS  - KVDB Block Validation",
		zap.String("kvdb_connection_info", d.KvdbConnectionInfo))

	conn, err := d.upgrader.Upgrade(w, req, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	ctx := req.Context()

	count := int64(0)
	started := false
	previousNum := int64(0)
	//previousValidNum := int64(0)
	batchHighBlockNum := int64(0)
	currentBlockNum := int64(0)

	startTime := time.Now()
	batchStartTime := time.Now()

	blocksTable := d.EOSdb.Blocks

	go websocketRead(conn)

	blocksTable.BaseTable.ReadRows(ctx, bt.InfiniteRange(""), func(row bt.Row) bool {
		count++

		currentBlockNum = int64(math.MaxUint32 - kvdb.BlockNum(row.Key()))

		isValid := utils.HasAllColumns(row, blocksTable.ColBlock, blocksTable.ColMetaIrreversible, blocksTable.ColMetaWritten, blocksTable.ColTransactionRefs, blocksTable.ColTransactionTraceRefs)

		if !started {
			previousNum = currentBlockNum + 1
			//previousValidNum = currentBlockNum + 1
			batchStartTime = time.Now()
			batchHighBlockNum = currentBlockNum
		}

		difference := previousNum - currentBlockNum

		if difference > 1 && started && isValid {
			msg := fmt.Sprintf("Found missing columns(s) %d rows\n", (uint32(previousNum-1) - uint32(currentBlockNum+1)))
			_ = sendMessage(conn, &BlockRange{
				StarBlock: uint32(currentBlockNum + 1),
				EndBlock:  uint32(previousNum - 1),
				Message:   msg,
				Status:    BlockRangeStatusHole,
			})
			batchHighBlockNum = currentBlockNum
		}

		if isValid {
			previousNum = currentBlockNum
		}

		if count%200000 == 0 {
			now := time.Now()
			msg := fmt.Sprintf("%d rows read (batch %s, total %s)\n", (uint32(batchHighBlockNum) - uint32(currentBlockNum)), now.Sub(batchStartTime), now.Sub(startTime))
			_ = sendMessage(conn, &BlockRange{
				StarBlock: uint32(currentBlockNum),
				EndBlock:  uint32(batchHighBlockNum),
				Message:   msg,
				Status:    BlockRangeStatusValid,
			})
			batchStartTime = time.Now()
			batchHighBlockNum = (currentBlockNum - 1)
		}

		started = true

		return true
	}, bt.RowFilter(bt.StripValueFilter()))
	now := time.Now()
	msg := fmt.Sprintf("%d rows read (batch %s, total %s)\n", (uint32(batchHighBlockNum) - uint32(currentBlockNum)), now.Sub(batchStartTime), now.Sub(startTime))
	_ = sendMessage(conn, &BlockRange{
		StarBlock: uint32(currentBlockNum),
		EndBlock:  uint32(batchHighBlockNum),
		Message:   msg,
		Status:    BlockRangeStatusValid,
	})
	zlog.Info("diagnose - EOS  - KVDB Block Validation - Completed")

}

func (d *Diagnose) ETHKVDBBlocks(w http.ResponseWriter, req *http.Request) {
	zlog.Info("diagnose - ETH  - KVDB Block Hole Checker",
		zap.String("kvdb_connection_info", d.KvdbConnectionInfo))

	conn, err := d.upgrader.Upgrade(w, req, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(req.Context())

	count := int64(0)
	started := false
	previousNum := uint64(0)
	//previousValidNum := uint64(0)
	batchHighBlockNum := uint64(0)
	currentBlockNum := uint64(0)

	startTime := time.Now()
	batchStartTime := time.Now()

	blocksTable := d.ETHdb.Blocks

	go websocketCloser(conn, cancel)

	// You can test on a lower range with `bt.NewRange("ff76abbf", "ff76abcf")`
	blocksTable.BaseTable.ReadRows(ctx, bt.InfiniteRange("blkn:"), func(row bt.Row) bool {
		count++

		currentBlockNum, _, err = ethdb.Keys.ReadBlockNum(row.Key())
		if err != nil {
			return false
		}

		//isValid := utils.HasAllColumns(row, blocksTable.ColHeaderProto, blocksTable.ColMetaIrreversible, blocksTable.ColMetaMapping, blocksTable.ColMetaWritten, blocksTable.ColTrxRefsProto, blocksTable.ColUnclesProto)

		if !started {
			previousNum = currentBlockNum + 1
			//previousValidNum = currentBlockNum + 1
			batchStartTime = time.Now()
			batchHighBlockNum = currentBlockNum
		}

		difference := previousNum - currentBlockNum
		//differenceInvalid := previousValidNum - currentBlockNum

		if difference > 1 && started {
			msg := fmt.Sprintf("Found block hole %d rows\n", (uint32(previousNum-1) - uint32(currentBlockNum+1)))
			_ = sendMessage(conn, &BlockRange{
				StarBlock: uint32(currentBlockNum + 1),
				EndBlock:  uint32(previousNum - 1),
				Message:   msg,
				Status:    BlockRangeStatusHole,
			})
			batchHighBlockNum = currentBlockNum
		}

		//if differenceInvalid > 1 && started && isValid {
		//	holeFound = true
		//	renderer.PutLine(w, "<p><strong>Found missing column(s) hole: [%d, %d]</strong></p>\n", currentBlockNum+1, previousValidNum-1)
		//}

		previousNum = currentBlockNum

		if count%200000 == 0 {
			now := time.Now()
			msg := fmt.Sprintf("%d rows read (batch %s, total %s)\n", (uint32(batchHighBlockNum) - uint32(currentBlockNum)), now.Sub(batchStartTime), now.Sub(startTime))
			_ = sendMessage(conn, &BlockRange{
				StarBlock: uint32(currentBlockNum),
				EndBlock:  uint32(batchHighBlockNum),
				Message:   msg,
				Status:    BlockRangeStatusValid,
			})
			batchStartTime = time.Now()
			batchHighBlockNum = (currentBlockNum - 1)
		}

		started = true

		return true
	}, bt.RowFilter(bt.StripValueFilter()))
	now := time.Now()
	msg := fmt.Sprintf("%d rows read (batch %s, total %s)\n", (uint32(batchHighBlockNum) - uint32(currentBlockNum)), now.Sub(batchStartTime), now.Sub(startTime))
	_ = sendMessage(conn, &BlockRange{
		StarBlock: uint32(currentBlockNum),
		EndBlock:  uint32(batchHighBlockNum),
		Message:   msg,
		Status:    BlockRangeStatusValid,
	})

	zlog.Info("diagnose - ETH  - KVDB Block Hole Checker  - Completed")
}

func (d *Diagnose) ETHKVDBBlockValidation(w http.ResponseWriter, req *http.Request) {
	zlog.Info("diagnose - ETH  - KVDB Block Validation",
		zap.String("kvdb_connection_info", d.KvdbConnectionInfo))

	conn, err := d.upgrader.Upgrade(w, req, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(req.Context())

	count := int64(0)
	started := false
	previousNum := uint64(0)
	//previousValidNum := uint64(0)
	batchHighBlockNum := uint64(0)
	currentBlockNum := uint64(0)

	startTime := time.Now()
	batchStartTime := time.Now()

	blocksTable := d.ETHdb.Blocks

	go websocketCloser(conn, cancel)

	blocksTable.BaseTable.ReadRows(ctx, bt.InfiniteRange("blkn:"), func(row bt.Row) bool {
		count++

		currentBlockNum, _, err = ethdb.Keys.ReadBlockNum(row.Key())
		if err != nil {
			return false
		}

		isValid := utils.HasAllColumns(row, blocksTable.ColHeaderProto, blocksTable.ColMetaIrreversible, blocksTable.ColMetaMapping, blocksTable.ColMetaWritten, blocksTable.ColTrxRefsProto, blocksTable.ColUnclesProto)

		if !started {
			previousNum = currentBlockNum + 1
			//previousValidNum = currentBlockNum + 1
			batchStartTime = time.Now()
			batchHighBlockNum = currentBlockNum
		}

		difference := previousNum - currentBlockNum

		if difference > 1 && started && isValid {
			msg := fmt.Sprintf("%d rows read\n", (uint32(batchHighBlockNum) - uint32(currentBlockNum+1)))
			_ = sendMessage(conn, &BlockRange{
				StarBlock: uint32(currentBlockNum + 1),
				EndBlock:  uint32(batchHighBlockNum),
				Message:   "",
				Status:    BlockRangeStatusValid,
			})
			msg = fmt.Sprintf("Found missing column(s) %d rows\n", (uint32(previousNum-1) - uint32(currentBlockNum+1)))
			_ = sendMessage(conn, &BlockRange{
				StarBlock: uint32(currentBlockNum + 1),
				EndBlock:  uint32(previousNum - 1),
				Message:   msg,
				Status:    BlockRangeStatusHole,
			})
			batchHighBlockNum = currentBlockNum
		}

		if isValid {
			previousNum = currentBlockNum
		}

		if count%200000 == 0 {
			now := time.Now()
			msg := fmt.Sprintf("%d rows read (batch %s, total %s)\n", (uint32(batchHighBlockNum) - uint32(currentBlockNum)), now.Sub(batchStartTime), now.Sub(startTime))
			_ = sendMessage(conn, &BlockRange{
				StarBlock: uint32(currentBlockNum),
				EndBlock:  uint32(batchHighBlockNum),
				Message:   msg,
				Status:    BlockRangeStatusValid,
			})
			batchStartTime = time.Now()
			batchHighBlockNum = (currentBlockNum - 1)
		}

		started = true

		return true
	}, bt.RowFilter(bt.StripValueFilter()))
	now := time.Now()
	msg := fmt.Sprintf("%d rows read (batch %s, total %s)\n", (uint32(batchHighBlockNum) - uint32(currentBlockNum)), now.Sub(batchStartTime), now.Sub(startTime))
	_ = sendMessage(conn, &BlockRange{
		StarBlock: uint32(currentBlockNum),
		EndBlock:  uint32(batchHighBlockNum),
		Message:   msg,
		Status:    BlockRangeStatusValid,
	})
	zlog.Info("diagnose - ETH  - KVDB Block Validation - Completed")
}
