package game

import (
	"github.com/Tnze/go-mc/level/block"
	"math"

	"github.com/anchormc/anchor/src/api/game"
)

type Chunk struct {
	x        int64
	z        int64
	sections []game.Section
}

func NewEmptyChunk(x, z, height int64) game.Chunk {
	sections := make([]game.Section, 0)

	for i := 0; i < int(math.Ceil(float64(height)/16.0)); i++ {
		sections = append(sections, NewEmptySection(int64(i)))
	}

	return &Chunk{
		x:        x,
		z:        z,
		sections: sections,
	}
}

func (c Chunk) GetSection(i int64) game.Section {
	return c.sections[i]
}

func (c Chunk) GetBlock(x, y, z int64) *block.Block {
	section := c.sections[y/16]

	if section == nil {
		return nil
	}

	return section.GetBlock(x, y%16, z)
}

func (c Chunk) SetBlock(x, y, z int64, block int) {
	section := c.sections[y/16]

	if section == nil {
		return
	}

	section.SetBlock(x, y%16, z, block)
}
