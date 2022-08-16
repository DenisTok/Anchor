package net

import (
	"bytes"
	"context"
	"crypto/cipher"
	"fmt"
	"github.com/Tnze/go-mc/data/packetid"
	"io"
	"net"
	"sync"
	"time"

	"github.com/anchormc/anchor/src/api"
	"github.com/anchormc/anchor/src/api/enum"
	"github.com/anchormc/anchor/src/api/game"
	"github.com/anchormc/anchor/src/api/log"
	"github.com/anchormc/anchor/src/api/proto"
	"github.com/anchormc/protocol"
	"github.com/google/uuid"
)

type Client struct {
	ctx context.Context

	uuid   uuid.UUID
	player game.Player
	conn   net.Conn
	r      io.Reader
	w      io.Writer
	m      *sync.Mutex

	keepAlive time.Time
}

func NewClient(conn net.Conn) (api.Client, error) {
	id, err := uuid.NewRandom()

	if err != nil {
		return nil, err
	}

	return &Client{
		ctx:  nil,
		uuid: id,
		conn: conn,
		r:    conn,
		w:    conn,
		m:    &sync.Mutex{},
	}, nil
}

func (c Client) UUID() string {
	return c.uuid.String()
}

func (c Client) ReadPacket() (*proto.Packet, error) {
	c.m.Lock()

	defer c.m.Unlock()

	var packetLength protocol.VarInt

	if _, err := packetLength.Decode(c.r); err != nil {
		return nil, err
	}

	data := make([]byte, packetLength)

	n, err := c.r.Read(data)

	if err != nil {
		return nil, err
	}

	if n < int(packetLength) {
		return nil, fmt.Errorf("short read when reading packet, expected %d bytes, only read %d bytes", packetLength, n)
	}

	buf := bytes.NewBuffer(data)

	var packetType protocol.VarInt

	if _, err := packetType.Decode(buf); err != nil {
		return nil, err
	}

	return proto.NewPacket(packetType, buf.Bytes()), nil
}

func (c *Client) WritePacket(packet *proto.Packet) error {
	dataBuffer := &bytes.Buffer{}

	if _, err := packet.ID.Encode(dataBuffer); err != nil {
		return err
	}

	if _, err := io.Copy(dataBuffer, packet.Buffer); err != nil {
		return err
	}

	packetBuffer := &bytes.Buffer{}

	if _, err := protocol.VarInt(dataBuffer.Len()).Encode(packetBuffer); err != nil {
		return err
	}

	if _, err := io.Copy(packetBuffer, dataBuffer); err != nil {
		return err
	}

	_, err := io.Copy(c.w, packetBuffer)

	return err
}

func (c Client) UnmarshalPacket(id protocol.VarInt, values ...protocol.DataTypeReader) error {
	packet, err := c.ReadPacket()

	if err != nil {
		return err
	}

	if packet.ID != id {
		return fmt.Errorf("packet ID mismatch, expected %02X, got %02X", id, packet.ID)
	}

	for _, value := range values {
		if _, err = value.Decode(packet.Buffer); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) MarshalPacket(id protocol.VarInt, values ...protocol.DataTypeWriter) error {
	buf := &bytes.Buffer{}

	for _, value := range values {
		if _, err := value.Encode(buf); err != nil {
			return err
		}
	}

	return c.WritePacket(proto.NewPacket(id, buf.Bytes()))
}

func (c *Client) HandlePackets(ctx context.Context, server api.Server) {
	nctx, can := context.WithCancel(ctx)

	c.ctx = nctx

	defer (func() {
		if err := c.Close(); err != nil {
			log.Error(err)
		}

		server.RemoveClient(c.UUID())

		can()

		log.Infof("Client %s has disconnected (connected: %d)\n", c.RemoteAddr(), len(server.GetAllClients()))
	})()

	var protocolVersion protocol.VarInt
	var serverAddress protocol.String
	var serverPort protocol.UnsignedShort
	var nextState protocol.VarInt

	if err := c.UnmarshalPacket(
		protocol.VarInt(0x00),
		&protocolVersion,
		&serverAddress,
		&serverPort,
		&nextState,
	); err != nil {
		log.Error(err)

		return
	}

	switch enum.ClientState(nextState) {
	case enum.ClientStateStatus:
		{
			if err := Status(server, c); err != nil {
				log.Error(err)
			}

			break
		}
	case enum.ClientStateLogin:
		{
			if err := Login(server, c); err != nil {
				log.Error(err)

				break
			}

			if err := Gameplay(server, c); err != nil {
				log.Error(err)

				break
			}

			c.runKeepAlive()

			for {
				p, err := c.ReadPacket()
				if err != nil {
					log.Error(err)
					break
				}

				err = c.handleClientPlayPacket(p)
				if err != nil {
					log.Error(err)
					break
				}

				log.Infof("got packet id: %v\n", p.ID)
			}
		}
	default:
		{
			log.Errorf("received unknown next state value from client: %d\n", nextState)
		}
	}
}

func (c *Client) runKeepAlive() {
	go func() {
		t := time.NewTicker(time.Second * 5)
		for {
			select {
			case <-c.ctx.Done():
				return
			case <-t.C:

			}

			err := c.MarshalPacket(packetid.ClientboundKeepAlive, protocol.Long(time.Now().UnixMilli()))
			if err != nil {
				log.Error(err)
			}

			if time.Since(c.keepAlive) >= time.Second*30 {
				// todo disconnect
				log.Infof("disconnected: keep alive timeout\n")
			}
		}
	}()
}

func (c *Client) handleClientPlayPacket(p *proto.Packet) error {
	switch p.ID {
	case packetid.ServerboundKeepAlive:
		log.Info("got keep alive")
		keepAliveID := protocol.Long(0)
		_, err := keepAliveID.Decode(p.Buffer)
		if err != nil {
			return err
		}

		c.keepAlive = time.Now()
	}
	return nil
}

func (c Client) GetPlayer() game.Player {
	return c.player
}

func (c *Client) SetPlayer(player game.Player) {
	c.player = player
}

func (c *Client) RemoteAddr() string {
	return c.conn.RemoteAddr().String()
}

func (c *Client) SetCipher(encode, decode cipher.Stream) {
	c.r = cipher.StreamReader{
		S: decode,
		R: c.conn,
	}

	c.w = cipher.StreamWriter{
		S: encode,
		W: c.conn,
	}
}

func (c *Client) Close() error {
	return c.conn.Close()
}

var _ api.Client = &Client{}
