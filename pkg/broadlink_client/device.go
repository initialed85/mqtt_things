package broadlink_client

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"slices"
	"sync"
	"time"
)

type SensorData struct {
	Temperature float64
	Humidity    float64
}

type Device struct {
	mu       *sync.Mutex
	client   *Client
	Name     string
	Type     uint16
	MAC      net.HardwareAddr
	Addr     net.UDPAddr
	LastSeen time.Time
	ID       int32
	Key      []byte
}

func (d *Device) doCommand(commandType uint16, commandPayload []byte, key []byte, timeout time.Duration) (responseHeader []byte, responsePayload []byte, err error) {
	rawDeviceType, err := packShort(int16(d.Type))
	if err != nil {
		return nil, nil, err
	}

	rawcommandType, err := packShort(int16(commandType))
	if err != nil {
		return nil, nil, err
	}

	sequenceNumber := d.client.getNextSequenceNumber()
	rawSequenceNumber, err := packShort(sequenceNumber)
	if err != nil {
		return nil, nil, err
	}

	rawMac := bytes.Clone([]byte(d.MAC))
	slices.Reverse(rawMac)

	rawID, err := packInt(d.ID)
	if err != nil {
		return nil, nil, err
	}

	rawPayloadChecksum, err := packShort(int16(getChecksum(commandPayload)))
	if err != nil {
		return nil, nil, err
	}

	requestPayload := make([]byte, 0x38)

	requestPayload = setBytes(requestPayload, []byte{0x5a, 0xa5, 0xaa, 0x55, 0x5a, 0xa5, 0xaa, 0x55}, 0x00)
	requestPayload = setBytes(requestPayload, rawDeviceType, 0x24)
	requestPayload = setBytes(requestPayload, rawcommandType, 0x26)
	requestPayload = setBytes(requestPayload, rawSequenceNumber, 0x28)
	requestPayload = setBytes(requestPayload, rawMac, 0x2a)
	requestPayload = setBytes(requestPayload, rawID, 0x30)
	requestPayload = setBytes(requestPayload, rawPayloadChecksum, 0x34)

	encryptedCommandPayload, err := encrypt(commandPayload, key)
	if err != nil {
		return nil, nil, err
	}

	rawEncryptedCommandPayloadChecksum, err := packShort(int16(getChecksum(encryptedCommandPayload)))
	if err != nil {
		return nil, nil, err
	}

	requestPayload = append(requestPayload, encryptedCommandPayload...)

	requestPayload = setBytes(requestPayload, rawEncryptedCommandPayloadChecksum, 0x20)

	request := d.client.getRequest(&d.Addr, requestPayload, sequenceNumber, make(chan *Response, 1))

	response, err := d.client.doCall(request, timeout)
	if err != nil {
		return nil, nil, err
	}

	if response.Err != nil {
		return nil, nil, response.Err
	}

	if len(response.Payload) <= 0x38 {
		return nil, nil, fmt.Errorf(
			"response not long enough (need more than %v bytes, got %v bytes)",
			0x38, len(response.Payload),
		)
	}

	responseHeader = response.Payload[:0x38]
	responsePayload = response.Payload[0x38:]

	return
}

func (d *Device) Discover(timeout time.Duration) error {
	sequenceNumber := d.client.getNextSequenceNumber()

	commandPayload, err := getDiscoveryPayload(time.Now(), d.client.sourceAddr, sequenceNumber)
	if err != nil {
		return err
	}

	responses := make(chan *Response, 65536)

	err = d.client.send(d.client.getRequest(&d.Addr, commandPayload, sequenceNumber, responses))
	if err != nil {
		return err
	}

	select {
	case response := <-responses:
		if response.Payload == nil {
			return fmt.Errorf("response payload unexpectedly empty; socket failed?")
		}

		rawMac := response.Payload[0x3a:0x40]
		slices.Reverse(rawMac)
		mac := net.HardwareAddr(rawMac)

		if response.Err != nil {
			return response.Err
		}

		d.Name = string(bytes.Split(response.Payload[0x40:], []byte{0x00})[0])
		d.Type = binary.LittleEndian.Uint16(response.Payload[0x34:0x36])
		d.MAC = mac
		d.LastSeen = response.ReceivedAt

		return nil
	case <-time.After(timeout):
		break
	}

	return fmt.Errorf("call timed out waiting for response after %v", timeout)
}

func (d *Device) Auth(timeout time.Duration) error {
	commandPayload := getAuthPayload()

	_, encryptedResponsePayload, err := d.doCommand(0x65, commandPayload, defaultKey, timeout)
	if err != nil {
		return err
	}

	responsePayload, err := decrypt(encryptedResponsePayload, defaultKey)
	if err != nil {
		return err
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	d.ID, err = unpackInt(responsePayload[:0x04])
	if err != nil {
		return err
	}

	d.Key = responsePayload[0x04:0x14]

	_, err = encrypt(getPadding(0x69, 64), d.Key)
	if err != nil {
		return fmt.Errorf("key %#+v failed self-test: %v", d.Key, err)
	}

	return nil
}

func (d *Device) send(commandType uint16, commandPayload []byte, timeout time.Duration) (uint16, []byte, error) {
	_, encryptedResponsePayload, err := d.doCommand(0x6a, commandPayload, d.Key, timeout)
	if err != nil {
		return 0, nil, err
	}

	responsePayload, err := decrypt(encryptedResponsePayload, d.Key)
	if err != nil {
		return 0, nil, err
	}

	length, err := unpackUnsignedShort(responsePayload[:0x02])
	if err != nil {
		return 0, nil, err
	}

	return length - 0x06, responsePayload[0x06:], nil
}

func (d *Device) checkAuthentication() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if len(d.Key) == 0 {
		return fmt.Errorf("not authenticated; have you called .Auth()?")
	}

	return nil
}

func (d *Device) GetSensorData(timeout time.Duration) (*SensorData, error) {
	err := d.checkAuthentication()
	if err != nil {
		return nil, err
	}

	commandPayload, err := getSensorPayload()
	if err != nil {
		return nil, err
	}

	length, responsePayload, err := d.send(0x6a, commandPayload, timeout)
	if err != nil {
		return nil, err
	}

	rawSensorData := responsePayload[:length+2]

	rawTemp, err := unpackDoubleChar(rawSensorData[:0x02])
	if err != nil {
		return nil, err
	}

	s := SensorData{
		Temperature: float64(rawTemp[0x00]) + float64(rawTemp[0x01])/100.0,
		Humidity:    float64(rawSensorData[0x02]) + float64(rawSensorData[0x03])/100.0,
	}

	return &s, nil
}

func (d *Device) Learn(timeout time.Duration) ([]byte, error) {
	deadline := time.Now().Add(timeout)

	err := d.checkAuthentication()
	if err != nil {
		return nil, err
	}

	commandPayload, err := getLearnPayload()
	if err != nil {
		return nil, err
	}

	_, _, err = d.send(0x6a, commandPayload, timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to enter learn mode: %v", err)
	}

	commandPayload, err = getLastCodePayload()
	if err != nil {
		return nil, err
	}

	for time.Now().Before(deadline) {
		_, responsePayload, err := d.send(0x6a, commandPayload, timeout)
		if err != nil {
			time.Sleep(time.Millisecond * 100)
			continue
		}

		return responsePayload, nil
	}

	return nil, fmt.Errorf("timed out after %v waiting to receive an IR code", timeout)
}

func (d *Device) SendIR(data []byte, timeout time.Duration) error {
	err := d.checkAuthentication()
	if err != nil {
		return err
	}

	commandPayload, err := getSendIRPayload(data)
	if err != nil {
		return err
	}

	_, _, err = d.send(0x6a, commandPayload, timeout)
	if err != nil {
		return fmt.Errorf("failed to send IR: %v", err)
	}

	return nil
}
