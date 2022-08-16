package game

import (
	"github.com/Tnze/go-mc/level/block"
	"github.com/anchormc/anchor/src/api/game"
)

type Region struct {
	x      int64
	z      int64
	chunks []game.Chunk
}

func NewEmptyRegion(x, z int64) game.Region {
	chunks := make([]game.Chunk, 32*32)

	for i := 0; i < 32*32; i++ {
		chunks[i] = nil
	}

	return &Region{
		x:      x,
		z:      z,
		chunks: chunks,
	}
}

func (r Region) GetChunk(x, z int64) game.Chunk {
	return r.chunks[x+z*32]
}

func (r *Region) SetChunk(x, z int64, chunk game.Chunk) {
	r.chunks[x+z*32] = chunk
}

func (r Region) GetBlock(x, y, z int64) *block.Block {
	chunk := r.GetChunk(x/16, z/16)

	if chunk == nil {
		return nil
	}

	return chunk.GetBlock(x%16, y, z%16)
}
