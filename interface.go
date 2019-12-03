package main

import (
	"time"
)

const (
	WebsocketTypeBlockRange  = "BlockRange"
	WebsocketTypeTransaction = "Transaction"
	WebsocketTypeMessage     = "Message"
	WebsocketTypePeerEvent   = "PeerEvent"
	WebsocketTypeProgress    = "Progress"
)

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

type Transaction struct {
	Prefix   string `json:"prefix"`
	Id       string `json:"id"`
	BlockNum uint32 `json:"blockNum"`
}

type Message struct {
	Msg string `json:"message"`
}

type Progress struct {
	Elapsed          time.Duration `json:"elapsed"`
	TotalIteration   int32         `json:"totalIteration"`
	CurrentIteration int32         `json:"currentIteration"`
}
