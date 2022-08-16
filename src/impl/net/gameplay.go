package net

import (
	_ "embed"
	"github.com/Tnze/go-mc/data/packetid"
	"github.com/Tnze/go-mc/level"
	"github.com/Tnze/go-mc/nbt"
	"github.com/anchormc/anchor/src/api"
	"github.com/anchormc/protocol"
	"io"
	"math/rand"
)

//go:embed DimensionCodec.snbt
var dimensionCodecSNBT []byte

//go:embed Dimension.snbt
var dimensionSNBT []byte

func Gameplay(server api.Server, client api.Client) error {
	var arr []protocol.DataType

	for _, s := range []string{
		"minecraft:overworld",
		"minecraft:nether",
		"minecraft:the_end",
		"minecraft:overworld_caves",
	} {
		id := protocol.Identifier(s)
		arr = append(arr, &id)
	}

	if err := client.MarshalPacket(
		packetid.ClientboundLogin,
		protocol.Int(client.GetPlayer().EntityID()),
		protocol.Boolean(false),  // Hardcore
		protocol.UnsignedByte(3), // Game mode
		protocol.Byte(-1),        // Previous game mode
		protocol.Array(arr),      // Dimensions
		protocol.NBT{Value: nbt.StringifiedMessage(dimensionCodecSNBT)},
		protocol.Identifier("minecraft:overworld"), // Dimension type
		protocol.Identifier("minecraft:temp"),      // Dimension
		protocol.Long(rand.Int63()),                // Seed
		protocol.VarInt(123),                       // Max players
		protocol.VarInt(1),                         // View distance
		protocol.VarInt(1),                         // Simulation distance
		protocol.Boolean(false),                    // Reduce debug info
		protocol.Boolean(true),                     // Enable respawn screen
		protocol.Boolean(false),                    // Is debug
		protocol.Boolean(true),                     // Is flat
		protocol.Boolean(false),                    // Has death location
	); err != nil {
		return err
	}

	//if err := client.MarshalPacket(
	//	protocol.VarInt(0x0B),
	//	protocol.UnsignedByte(0), // TODO server difficulty
	//	protocol.Boolean(true),
	//); err != nil {
	//	return err
	//}

	var locale protocol.String
	var viewDistance protocol.Byte
	var chatMode protocol.VarInt
	var chatColors protocol.Boolean
	var displayedSkinParts protocol.UnsignedByte
	var mainHand protocol.VarInt
	var enableTextFiltering protocol.Boolean
	var allowServerListings protocol.Boolean

	if err := client.UnmarshalPacket(
		packetid.ClientboundBlockEntityData,
		&locale,
		&viewDistance,
		&chatMode,
		&chatColors,
		&displayedSkinParts,
		&mainHand,
		&enableTextFiltering,
		&allowServerListings,
	); err != nil {
		return err
	}

	position := client.GetPlayer().Position()

	if err := client.MarshalPacket(
		packetid.ClientboundPlayerPosition,
		protocol.Double(position.X),
		protocol.Double(position.Y),
		protocol.Double(position.Z),
		protocol.Float(0),
		protocol.Float(90),
		protocol.Byte(0),
		protocol.VarInt(rand.Int31()),
		protocol.Boolean(true),
	); err != nil {
		return err
	}

	err := writeChunk(client, 0, 0, 256)
	if err != nil {
		return err
	}

	return nil
}

func writeChunk(c api.Client, x int, z int, height int) error {
	chunk := level.EmptyChunk(height)
	data, _ := chunk.Data()
	return c.MarshalPacket(
		packetid.ClientboundLevelChunkWithLight,
		protocol.Int(x), protocol.Int(z),
		protocol.NBT{Value: nbt.StringifiedMessage("{}")},
		protocol.ByteArray(data),
		protocol.ByteArray{}, // pretending block entity array
		protocol.Boolean(false),
		BitSet{},
		BitSet{},
		BitSet{},
		BitSet{},
		protocol.ByteArray{},
		protocol.ByteArray{},
	)
}

type BitSet []int64

func (b BitSet) Encode(w io.Writer) (n int64, err error) {
	n, err = protocol.VarInt(len(b)).Encode(w)
	if err != nil {
		return
	}
	for i := range b {
		n2, err := protocol.Long(b[i]).Encode(w)
		if err != nil {
			return n + n2, err
		}
		n += n2
	}
	return
}
