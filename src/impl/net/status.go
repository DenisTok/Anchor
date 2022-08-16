package net

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image/png"

	"github.com/anchormc/anchor/src/api"
	"github.com/anchormc/anchor/src/api/data"
	"github.com/anchormc/protocol"
)

func Status(server api.Server, client api.Client) error {
	requestPacket, err := client.ReadPacket()

	if err != nil {
		return err
	}

	if requestPacket.ID != protocol.VarInt(0x00) {
		return fmt.Errorf("packet ID mismatch, expected 0x00, got %02X", requestPacket.ID)
	}

	response := data.StatusResponse{
		Version: data.StatusVersion{
			Name:     "1.19",
			Protocol: 759,
		},
		Players: data.StatusPlayers{
			Online: server.PlayerCount(),
			Max:    server.GetConfig().MaxPlayers,
		},
		Description: server.GetConfig().MOTD,
	}

	if favicon := server.Favicon(); favicon != nil {
		buf := &bytes.Buffer{}

		if err = png.Encode(buf, favicon); err != nil {
			return err
		}

		response.Favicon = "data:image/png;base64," + base64.RawStdEncoding.EncodeToString(buf.Bytes())
	}

	data, err := json.Marshal(response)

	if err != nil {
		return err
	}

	if err = client.MarshalPacket(
		protocol.VarInt(0x00),
		protocol.String(string(data)),
	); err != nil {
		return err
	}

	var payload protocol.Long

	if err = client.UnmarshalPacket(
		protocol.VarInt(0x01),
		&payload,
	); err != nil {
		return err
	}

	return client.MarshalPacket(
		protocol.VarInt(0x01),
		payload,
	)
}
