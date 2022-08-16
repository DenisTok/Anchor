package game

import "github.com/Tnze/go-mc/level/block"

type Section interface {
	GetBlock(x, y, z int64) *block.Block
	SetBlock(x, y, z int64, state int)
}
