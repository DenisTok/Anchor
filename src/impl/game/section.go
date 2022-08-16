package game

import (
	"github.com/Tnze/go-mc/level"
	"github.com/Tnze/go-mc/level/block"
	"github.com/anchormc/anchor/src/api/game"
)

type Section struct {
	index   int64
	storage *level.BitStorage
}

func NewEmptySection(i int64) game.Section {
	return &Section{
		index:   i,
		storage: level.NewBitStorage(8, 16*16*16, make([]uint64, 16*16*16)),
	}
}

func (s Section) GetBlock(x, y, z int64) *block.Block {
	//state := s.storage.Get(int(x) + int(z)*16 + int(y)*16*16)
	//
	//blockID := level.StateID[uint32(state)]
	//
	//return block.ByID[blockID]
	return nil
}

func (s *Section) SetBlock(x, y, z int64, state int) {
	s.storage.Set(int(x)+int(z)*16+int(y)*16*16, state)
}
