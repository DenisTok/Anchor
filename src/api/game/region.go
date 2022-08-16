package game

import "github.com/Tnze/go-mc/level/block"

type Region interface {
	GetChunk(x, z int64) Chunk
	SetChunk(x, z int64, chunk Chunk)
	GetBlock(x, y, z int64) *block.Block
}
