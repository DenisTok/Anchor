package game

import (
	"fmt"
	"github.com/Tnze/go-mc/level/block"
	"io/ioutil"
	"path"

	"github.com/Tnze/go-mc/save"
	"github.com/Tnze/go-mc/save/region"
	"github.com/anchormc/anchor/src/api/game"
)

type World struct {
	name      string
	folder    string
	generator game.Generator
	regions   map[string]game.Region
}

func NewWorld(name, folder string, generator game.Generator) game.World {
	return &World{
		name:      name,
		folder:    folder,
		generator: generator,
		regions:   make(map[string]game.Region),
	}
}

func (w World) Name() string {
	return w.name
}

func (w World) Folder() string {
	return w.folder
}

func (w *World) LoadAllRegions() error {
	regions, err := ioutil.ReadDir(path.Join(w.folder, "region"))

	if err != nil {
		return err
	}

	for _, file := range regions {
		r, err := region.Open(path.Join(w.folder, "region", file.Name()))

		if err != nil {
			return err
		}

		for x := 0; x < 32; x++ {
			for z := 0; z < 32; z++ {
				if !r.ExistSector(x, z) {
					continue
				}

				data, err := r.ReadSector(x, z)

				if err != nil {
					return err
				}

				rawChunk := &save.Chunk{}

				if err = rawChunk.Load(data); err != nil {
					return err
				}

				// TODO finish region file loading
			}
		}
	}

	return nil
}

func (w World) GetRegion(x, z int64) game.Region {
	return w.regions[fmt.Sprintf("%d:%d", x, z)]
}

func (w World) GetChunk(x, z int64) game.Chunk {
	region := w.GetRegion(x/32, z/32)

	if region == nil {
		return nil
	}

	return region.GetChunk(x%32, z%32)
}

func (w World) GetBlock(x, y, z int64) *block.Block {
	chunk := w.GetChunk(x/16, z/16)

	if chunk == nil {
		return nil
	}

	return chunk.GetBlock(x%16, y, z%16)
}

func (w *World) GenerateChunk(x, z int64) error {
	chunk, err := w.generator.GenerateChunk(x, z, 256)

	if err != nil {
		return err
	}

	region, ok := w.regions[fmt.Sprintf("%d:%d", x/32, z/32)]

	if !ok {
		region = NewEmptyRegion(x/32, z/32)

		w.regions[fmt.Sprintf("%d:%d", x/32, z/32)] = region
	}

	region.SetChunk(x%32, z%32, chunk)

	return nil
}
