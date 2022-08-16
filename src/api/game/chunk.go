package game

import "github.com/Tnze/go-mc/level/block"

type Chunk interface {
	GetSection(i int64) Section
	GetBlock(x, y, z int64) *block.Block
	SetBlock(x, y, z int64, state int)
}
