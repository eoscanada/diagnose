package main

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
