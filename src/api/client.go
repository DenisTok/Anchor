package api

import (
	"context"
	"crypto/cipher"

	"github.com/anchormc/anchor/src/api/game"
	"github.com/anchormc/anchor/src/api/proto"
	"github.com/anchormc/protocol"
)

type Client interface {
	UUID() string
	ReadPacket() (*proto.Packet, error)
	WritePacket(*proto.Packet) error
	UnmarshalPacket(protocol.VarInt, ...protocol.DataTypeReader) error
	MarshalPacket(protocol.VarInt, ...protocol.DataTypeWriter) error
	HandlePackets(ctx context.Context, server Server)
	RemoteAddr() string
	GetPlayer() game.Player
	SetPlayer(game.Player)
	SetCipher(cipher.Stream, cipher.Stream)
	Close() error
}
