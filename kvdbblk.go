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
	"github.com/eoscanada/kvdb/eosdb"
	"github.com/eoscanada/kvdb/ethdb"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func (d *Diagnose) EOSKVDBBlocks(w http.ResponseWriter, req *http.Request) {
	kvdbInfo, db := d.getEOSDatabase(w, req)
	if kvdbInfo == nil || db == nil {
		return
	}

	zlog.Info("diagnose - EOS  - KVDB Block Hole Checker", zap.Reflect("connection_info", kvdbInfo))
	conn, err := d.upgrader.Upgrade(w, req, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(req.Context())

	count := int64(0)
	started := false
	previousNum := int64(0)
	batchHighBlockNum := int64(0)
	currentBlockNum := int64(0)

	startTime := time.Now()
	batchStartTime := time.Now()

	go readWebsocket(conn, cancel)

	maybeSendWebsocket(conn, WebsocketTypeProgress, Progress{Elapsed: time.Now().Sub(startTime)})
	db.Blocks.BaseTable.ReadRows(ctx, bt.InfiniteRange(""), func(row bt.Row) bool {
		count++

		currentBlockNum = int64(math.MaxUint32 - kvdb.BlockNum(row.Key()))

		if count%5000 == 0 {
			maybeSendWebsocket(conn, WebsocketTypeProgress, Progress{Elapsed: time.Now().Sub(startTime)})
		}

		if !started {
			previousNum = currentBlockNum + 1
			batchStartTime = time.Now()
			batchHighBlockNum = currentBlockNum

		}

		difference := previousNum - currentBlockNum

		if difference > 1 && started {

			msg := fmt.Sprintf("%d rows read", (uint32(batchHighBlockNum) - uint32(currentBlockNum+1)))
			maybeSendWebsocket(conn, WebsocketTypeBlockRange, &BlockRange{
				StarBlock: uint32(currentBlockNum + 1),
				EndBlock:  uint32(batchHighBlockNum),
				Message:   "",
				Status:    BlockRangeStatusValid,
			})
			msg = fmt.Sprintf("Found block hole %d rows", (uint32(previousNum-1) - uint32(currentBlockNum+1)))
			maybeSendWebsocket(conn, WebsocketTypeBlockRange, &BlockRange{
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
			msg := fmt.Sprintf("%d rows read (batch %s, total %s)", (uint32(batchHighBlockNum) - uint32(currentBlockNum)), now.Sub(batchStartTime), now.Sub(startTime))
			maybeSendWebsocket(conn, WebsocketTypeBlockRange, &BlockRange{
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
	msg := fmt.Sprintf("%d rows read (batch %s, total %s)", (uint32(batchHighBlockNum) - uint32(currentBlockNum)), now.Sub(batchStartTime), now.Sub(startTime))
	maybeSendWebsocket(conn, WebsocketTypeBlockRange, &BlockRange{
		StarBlock: uint32(currentBlockNum),
		EndBlock:  uint32(batchHighBlockNum),
		Message:   msg,
		Status:    BlockRangeStatusValid,
	})
	zlog.Info("diagnose - EOS  - KVDB Block Hole Checker - Completed")
}

func (d *Diagnose) EOSKVDBBlocksValidation(w http.ResponseWriter, req *http.Request) {
	kvdbInfo, db := d.getEOSDatabase(w, req)
	if kvdbInfo == nil || db == nil {
		return
	}

	zlog.Info("diagnose - EOS  - KVDB Block Validation", zap.Reflect("connection_info", kvdbInfo))

	conn, err := d.upgrader.Upgrade(w, req, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(req.Context())

	count := int64(0)
	started := false
	previousNum := int64(0)
	batchHighBlockNum := int64(0)
	currentBlockNum := int64(0)

	startTime := time.Now()
	batchStartTime := time.Now()

	go readWebsocket(conn, cancel)

	maybeSendWebsocket(conn, WebsocketTypeProgress, Progress{Elapsed: time.Now().Sub(startTime)})

	db.Blocks.BaseTable.ReadRows(ctx, bt.InfiniteRange(""), func(row bt.Row) bool {
		count++

		currentBlockNum = int64(math.MaxUint32 - kvdb.BlockNum(row.Key()))

		isValid := utils.HasAllColumns(row, db.Blocks.ColBlock, db.Blocks.ColMetaIrreversible, db.Blocks.ColMetaWritten, db.Blocks.ColTransactionRefs, db.Blocks.ColTransactionTraceRefs)
		if count%5000 == 0 {
			maybeSendWebsocket(conn, WebsocketTypeProgress, Progress{Elapsed: time.Now().Sub(startTime)})
		}

		if !started {
			previousNum = currentBlockNum + 1
			batchStartTime = time.Now()
			batchHighBlockNum = currentBlockNum
		}

		difference := previousNum - currentBlockNum

		if difference > 1 && started && isValid {
			msg := fmt.Sprintf("Found missing columns(s) %d rows", (uint32(previousNum-1) - uint32(currentBlockNum+1)))
			maybeSendWebsocket(conn, WebsocketTypeBlockRange, &BlockRange{
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
			msg := fmt.Sprintf("%d rows read (batch %s, total %s)", (uint32(batchHighBlockNum) - uint32(currentBlockNum)), now.Sub(batchStartTime), now.Sub(startTime))
			maybeSendWebsocket(conn, WebsocketTypeBlockRange, &BlockRange{
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
	msg := fmt.Sprintf("%d rows read (batch %s, total %s)", (uint32(batchHighBlockNum) - uint32(currentBlockNum)), now.Sub(batchStartTime), now.Sub(startTime))
	maybeSendWebsocket(conn, WebsocketTypeBlockRange, &BlockRange{
		StarBlock: uint32(currentBlockNum),
		EndBlock:  uint32(batchHighBlockNum),
		Message:   msg,
		Status:    BlockRangeStatusValid,
	})
	zlog.Info("diagnose - EOS  - KVDB Block Validation - Completed")

}

func (d *Diagnose) ETHKVDBBlocks(w http.ResponseWriter, req *http.Request) {
	kvdbInfo, db := d.getETHDatabase(w, req)
	if kvdbInfo == nil || db == nil {
		return
	}

	zlog.Info("diagnose - ETH  - KVDB Block Hole Checker", zap.Reflect("connection_info", kvdbInfo))

	conn, err := d.upgrader.Upgrade(w, req, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(req.Context())

	count := int64(0)
	started := false
	previousNum := uint64(0)
	batchHighBlockNum := uint64(0)
	currentBlockNum := uint64(0)

	startTime := time.Now()
	batchStartTime := time.Now()

	go readWebsocket(conn, cancel)
	maybeSendWebsocket(conn, WebsocketTypeProgress, Progress{Elapsed: time.Now().Sub(startTime)})
	// You can test on a lower range with `bt.NewRange("ff76abbf", "ff76abcf")`
	db.Blocks.BaseTable.ReadRows(ctx, bt.InfiniteRange("blkn:"), func(row bt.Row) bool {
		count++

		currentBlockNum, _, err = ethdb.Keys.ReadBlockNum(row.Key())
		if err != nil {
			return false
		}

		if count%5000 == 0 {
			maybeSendWebsocket(conn, WebsocketTypeProgress, Progress{Elapsed: time.Now().Sub(startTime)})
		}

		if !started {
			previousNum = currentBlockNum + 1
			batchStartTime = time.Now()
			batchHighBlockNum = currentBlockNum
		}

		difference := previousNum - currentBlockNum

		if difference > 1 && started {
			msg := fmt.Sprintf("Found block hole %d rows", (uint32(previousNum-1) - uint32(currentBlockNum+1)))
			maybeSendWebsocket(conn, WebsocketTypeBlockRange, &BlockRange{
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
			msg := fmt.Sprintf("%d rows read (batch %s, total %s)", (uint32(batchHighBlockNum) - uint32(currentBlockNum)), now.Sub(batchStartTime), now.Sub(startTime))
			maybeSendWebsocket(conn, WebsocketTypeBlockRange, &BlockRange{
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
	msg := fmt.Sprintf("%d rows read (batch %s, total %s)", (uint32(batchHighBlockNum) - uint32(currentBlockNum)), now.Sub(batchStartTime), now.Sub(startTime))
	maybeSendWebsocket(conn, WebsocketTypeBlockRange, &BlockRange{
		StarBlock: uint32(currentBlockNum),
		EndBlock:  uint32(batchHighBlockNum),
		Message:   msg,
		Status:    BlockRangeStatusValid,
	})

	zlog.Info("diagnose - ETH  - KVDB Block Hole Checker  - Completed")
}

func (d *Diagnose) ETHKVDBBlockValidation(w http.ResponseWriter, req *http.Request) {
	kvdbInfo, db := d.getETHDatabase(w, req)
	if kvdbInfo == nil || db == nil {
		return
	}

	zlog.Info("diagnose - ETH  - KVDB Block Validation", zap.Reflect("connection_info", kvdbInfo))

	conn, err := d.upgrader.Upgrade(w, req, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(req.Context())

	count := int64(0)
	started := false
	previousNum := uint64(0)
	batchHighBlockNum := uint64(0)
	currentBlockNum := uint64(0)

	startTime := time.Now()
	batchStartTime := time.Now()

	go readWebsocket(conn, cancel)

	maybeSendWebsocket(conn, WebsocketTypeProgress, Progress{Elapsed: time.Now().Sub(startTime)})
	db.Blocks.BaseTable.ReadRows(ctx, bt.InfiniteRange("blkn:"), func(row bt.Row) bool {
		count++

		currentBlockNum, _, err = ethdb.Keys.ReadBlockNum(row.Key())
		if err != nil {
			return false
		}

		isValid := utils.HasAllColumns(row, db.Blocks.ColHeaderProto, db.Blocks.ColMetaIrreversible, db.Blocks.ColMetaMapping, db.Blocks.ColMetaWritten, db.Blocks.ColTrxRefsProto, db.Blocks.ColUnclesProto)
		if count%5000 == 0 {
			maybeSendWebsocket(conn, WebsocketTypeProgress, Progress{Elapsed: time.Now().Sub(startTime)})
		}

		if !started {
			previousNum = currentBlockNum + 1
			batchStartTime = time.Now()
			batchHighBlockNum = currentBlockNum
		}

		difference := previousNum - currentBlockNum

		if difference > 1 && started && isValid {
			msg := fmt.Sprintf("%d rows read", (uint32(batchHighBlockNum) - uint32(currentBlockNum+1)))
			maybeSendWebsocket(conn, WebsocketTypeBlockRange, &BlockRange{
				StarBlock: uint32(currentBlockNum + 1),
				EndBlock:  uint32(batchHighBlockNum),
				Message:   "",
				Status:    BlockRangeStatusValid,
			})
			msg = fmt.Sprintf("Found missing column(s) %d rows\n", (uint32(previousNum-1) - uint32(currentBlockNum+1)))
			maybeSendWebsocket(conn, WebsocketTypeBlockRange, &BlockRange{
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
			maybeSendWebsocket(conn, WebsocketTypeBlockRange, &BlockRange{
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
	maybeSendWebsocket(conn, WebsocketTypeBlockRange, &BlockRange{
		StarBlock: uint32(currentBlockNum),
		EndBlock:  uint32(batchHighBlockNum),
		Message:   msg,
		Status:    BlockRangeStatusValid,
	})
	zlog.Info("diagnose - ETH  - KVDB Block Validation - Completed")
}

func (d *Diagnose) extractConnectionInfo(w http.ResponseWriter, req *http.Request) *kvdb.ConnectionInfo {
	params := mux.Vars(req)

	connectionInfo := params["connection_info"]
	if connectionInfo == "" {
		connectionInfo = d.KvdbConnectionInfo
	}

	kvdbInfo, err := kvdb.NewConnectionInfo(connectionInfo)
	if err != nil {
		fmt.Fprintf(w, "invalid connection info: %s", err)
		return nil
	}

	return kvdbInfo
}

func (d *Diagnose) getEOSDatabase(w http.ResponseWriter, req *http.Request) (*kvdb.ConnectionInfo, *eosdb.EOSDatabase) {
	kvdbInfo := d.extractConnectionInfo(w, req)
	if kvdbInfo == nil {
		return nil, nil
	}

	db, err := eosdb.New(kvdbInfo.TablePrefix, kvdbInfo.Project, kvdbInfo.Instance, false)
	if err != nil {
		fmt.Fprintf(w, "unable to create EOS database: %s", err)
		return nil, nil
	}

	return kvdbInfo, db
}

func (d *Diagnose) getETHDatabase(w http.ResponseWriter, req *http.Request) (*kvdb.ConnectionInfo, *ethdb.ETHDatabase) {
	kvdbInfo := d.extractConnectionInfo(w, req)
	if kvdbInfo == nil {
		return nil, nil
	}

	db, err := ethdb.New(kvdbInfo.TablePrefix, kvdbInfo.Project, kvdbInfo.Instance, false)
	if err != nil {
		fmt.Fprintf(w, "unable to create ETH database: %s", err)
		return nil, nil
	}

	return kvdbInfo, db
}
