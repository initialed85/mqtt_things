package broadlink_client

import (
	"bytes"
	"encoding/binary"
	"net"
	"time"
)

func getDiscoveryPayload(now time.Time, sourceAddr *net.UDPAddr, sequenceNumber int16) (payload []byte, err error) {
	rawSequenceNumber, err := packShort(sequenceNumber)
	if err != nil {
		return nil, err
	}

	payload = make([]byte, 0x30)

	payload = setBytes(payload, packTime(now), 0x08)
	payload = setBytes(payload, sourceAddr.IP.To4()[0:4], 0x18)
	binary.LittleEndian.PutUint16(payload[0x1c:0x1d+1], uint16(sourceAddr.Port))
	payload = setBytes(payload, []byte{0x06}, 0x26)
	payload = setBytes(payload, rawSequenceNumber, 0x28)

	return payload, nil
}

func getAuthPayload() (payload []byte) {
	payload = make([]byte, 0x50)
	payload = setBytes(payload, getPadding(0x31, 16), 0x04)
	payload[0x1e] = 0x01
	payload[0x2d] = 0x01
	payload = setBytes(payload, bytes.NewBufferString("Test 1\x00").Bytes(), 0x30)

	return
}

func getSensorPayload() (payload []byte, err error) {
	payload = make([]byte, 0)

	rawDataLength, err := packUnsignedShort(4)
	if err != nil {
		return nil, err
	}

	rawCommand, err := packUnsignedInt(0x24)
	if err != nil {
		return nil, err
	}

	payload = append(payload, rawDataLength...)
	payload = append(payload, rawCommand...)

	return payload, nil
}

func getLearnPayload() (payload []byte, err error) {
	payload = make([]byte, 0)

	rawDataLength, err := packUnsignedShort(4)
	if err != nil {
		return nil, err
	}

	rawCommand, err := packUnsignedInt(0x03)
	if err != nil {
		return nil, err
	}

	payload = append(payload, rawDataLength...)
	payload = append(payload, rawCommand...)

	return payload, nil
}

func getLastCodePayload() (payload []byte, err error) {
	payload = make([]byte, 0)

	rawDataLength, err := packUnsignedShort(4)
	if err != nil {
		return nil, err
	}

	rawCommand, err := packUnsignedInt(0x04)
	if err != nil {
		return nil, err
	}

	payload = append(payload, rawDataLength...)
	payload = append(payload, rawCommand...)

	return payload, nil
}

func getSendIRPayload(data []byte) (payload []byte, err error) {
	payload = make([]byte, 0)

	rawDataLength, err := packUnsignedShort(4 + uint16(len(data)))
	if err != nil {
		return nil, err
	}

	rawCommand, err := packUnsignedInt(0x02)
	if err != nil {
		return nil, err
	}

	payload = append(payload, rawDataLength...)
	payload = append(payload, rawCommand...)
	payload = append(payload, data...)

	return payload, nil
}
