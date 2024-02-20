package broadlink_client

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"fmt"
	"time"

	binarypack "github.com/canhlinh/go-binary-pack"
)

func setBytes(payload []byte, bytes []byte, start int) []byte {
	adjustedPayload := payload

	for i, v := range bytes {
		payload[start+i] = v
	}

	return adjustedPayload
}

func getPadding(value byte, length int) []byte {
	padding := make([]byte, length)

	for i := 0; i < length; i++ {
		padding = append(padding, value)
	}

	return padding
}

func getChecksum(payload []byte) uint16 {
	var checkSum uint16
	checkSum = 0xbeaf
	for i := 0; i < len(payload); i++ {
		checkSum += uint16(payload[i])

		if checkSum > 0xffff {
			checkSum -= 0xffff
		}
	}

	return checkSum
}

func encrypt(plainText []byte, key []byte) (cipherText []byte, err error) {
	for len(plainText)%aes.BlockSize != 0 {
		plainText = append(plainText, 0x00)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	cipherText = make([]byte, len(plainText))

	cbc := cipher.NewCBCEncrypter(block, defaultIV)
	cbc.CryptBlocks(cipherText, plainText)

	return
}

func decrypt(cipherText []byte, key []byte) (plainText []byte, err error) {
	var block cipher.Block

	if block, err = aes.NewCipher(key); err != nil {
		return
	}

	if len(cipherText) < aes.BlockSize {
		err = fmt.Errorf("cipherText too short")
		return
	}

	includedPadding := len(cipherText) % aes.BlockSize
	if includedPadding > 0 && includedPadding < aes.BlockSize {
		cipherText = cipherText[aes.BlockSize:]
	}

	plainText = make([]byte, len(cipherText))
	cbc := cipher.NewCBCDecrypter(block, defaultIV)
	cbc.CryptBlocks(plainText, cipherText)

	return
}

func packShort(value int16) (payload []byte, err error) {
	payload, err = new(binarypack.BinaryPack).Pack([]string{"h"}, []interface{}{int(value)})
	if err != nil {
		return
	}

	return
}

func unpackShort(payload []byte) (int16, error) {
	unpackedValue, err := new(binarypack.BinaryPack).UnPack([]string{"h"}, payload)
	if err != nil {
		return 0, err
	}

	return int16(unpackedValue[0].(int)), nil
}

func packUnsignedShort(value uint16) (payload []byte, err error) {
	payload, err = new(binarypack.BinaryPack).Pack([]string{"H"}, []interface{}{int(value)})
	if err != nil {
		return
	}

	return
}

func packInt(value int32) (payload []byte, err error) {
	payload, err = new(binarypack.BinaryPack).Pack([]string{"i"}, []interface{}{int(value)})
	if err != nil {
		return
	}

	return
}

func packUnsignedInt(value uint32) (payload []byte, err error) {
	payload, err = new(binarypack.BinaryPack).Pack([]string{"I"}, []interface{}{int(value)})
	if err != nil {
		return
	}

	return
}

func unpackUnsignedShort(payload []byte) (uint16, error) {
	unpackedValue, err := new(binarypack.BinaryPack).UnPack([]string{"H"}, payload)
	if err != nil {
		return 0, err
	}

	return uint16(unpackedValue[0].(int)), nil
}

func unpackInt(payload []byte) (int32, error) {
	unpackedValue, err := new(binarypack.BinaryPack).UnPack([]string{"i"}, payload)
	if err != nil {
		return 0, err
	}

	return int32(unpackedValue[0].(int)), nil
}

func unpackDoubleChar(payload []byte) ([]byte, error) {
	unpackedValue, err := new(binarypack.BinaryPack).UnPack([]string{"b", "b"}, payload)
	if err != nil {
		return nil, err
	}

	return []byte{uint8(unpackedValue[0].(int)), uint8(unpackedValue[1].(int))}, nil
}

func packTime(now time.Time) (payload []byte) {
	payload = make([]byte, 12)

	_, rawOffset := now.Zone()
	offsetHours := rawOffset / 3600

	binary.LittleEndian.PutUint32(payload[:0x04], uint32(offsetHours))
	binary.LittleEndian.PutUint16(payload[0x04:0x06], uint16(now.Year()))
	payload[0x06] = uint8(now.Minute())
	payload[0x07] = uint8(now.Hour())
	payload[0x08] = uint8(now.Year() - 2000)
	payload[0x09] = uint8(now.Weekday())
	payload[0x0a] = uint8(now.Day())
	payload[0x0b] = uint8(now.Month())

	return payload
}
