package net

import (
	"crypto/aes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	random "math/rand"

	"github.com/Tnze/go-mc/net/CFB8"
	"github.com/anchormc/anchor/src/api"
	"github.com/anchormc/anchor/src/api/util"
	"github.com/anchormc/anchor/src/impl/game"
	"github.com/anchormc/protocol"
	UUID "github.com/google/uuid"
)

var isOnline = true

func Login(server api.Server, client api.Client) error {
	var nickname protocol.String
	var hasSigData protocol.Boolean
	var timestamp protocol.Long
	var cpublicKey protocol.ByteArray
	var signature protocol.ByteArray

	if isOnline {
		if err := client.UnmarshalPacket(
			protocol.VarInt(0x00),
			&nickname,
			&hasSigData,
			&timestamp,
			&cpublicKey,
			&signature,
		); err != nil {
			return err
		}
	} else {
		if err := client.UnmarshalPacket(
			protocol.VarInt(0x00),
			&nickname,
		); err != nil {
			return err
		}
	}

	var username string
	var uuid protocol.UUID

	if server.GetConfig().OnlineMode {
		verifyToken := make([]byte, 4)

		if _, err := rand.Read(verifyToken); err != nil {
			return err
		}

		privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
		if err != nil {
			return err
		}

		publicKey, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
		if err != nil {
			return err
		}

		if err = client.MarshalPacket(
			protocol.VarInt(0x01),
			protocol.String(""),
			protocol.ByteArray(publicKey),
			protocol.ByteArray(verifyToken),
		); err != nil {
			return err
		}

		var sharedSecret protocol.ByteArray
		var encryptedVerifyToken protocol.ByteArray
		var hasVerifyToken protocol.Boolean
		//var salt protocol.Long

		if err = client.UnmarshalPacket(
			protocol.VarInt(0x01),
			&sharedSecret,
			&hasVerifyToken,
			&encryptedVerifyToken,
			//&salt,
		); err != nil {
			return err
		}

		decodedSharedSecret, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, sharedSecret)
		if err != nil {
			return err
		}

		//decodedVerifyToken, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, encryptedVerifyToken)
		//if err != nil {
		//	return err
		//}
		//
		//if !bytes.Equal(decodedVerifyToken, verifyToken) {
		//	return fmt.Errorf("decrypted verify token does not match server token")
		//}

		block, err := aes.NewCipher(decodedSharedSecret)

		if err != nil {
			return err
		}

		client.SetCipher(CFB8.NewCFB8Encrypt(block, decodedSharedSecret), CFB8.NewCFB8Decrypt(block, decodedSharedSecret))

		hash := util.AuthDigest("", decodedSharedSecret, publicKey)

		response, err := util.Authenticate(string(nickname), hash)

		if err != nil {
			return err
		}

		username = response.Name
		uuid = protocol.UUID(response.ID)
	} else {
		username = "OfflinePlayer"
		uuid = protocol.UUID(UUID.NewString())
	}

	if err := client.MarshalPacket(
		protocol.VarInt(0x02),
		uuid,
		protocol.String(username),
		protocol.VarInt(0),
	); err != nil {
		return err
	}

	client.SetPlayer(game.NewPlayer(random.Int63(), username, string(uuid), protocol.AbsolutePosition{}))

	return nil
}
