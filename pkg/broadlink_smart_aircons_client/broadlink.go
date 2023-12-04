package broadlink_smart_aircons_client

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"fmt"
	"github.com/initialed85/glue/pkg/network"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const (
	timeout = time.Second * 2
)

var (
	DefaultKey = []byte{0x09, 0x76, 0x28, 0x34, 0x3f, 0xe9, 0x9e, 0x23, 0x76, 0x5c, 0x15, 0x13, 0xac, 0xcf, 0x8b, 0x02}
	DefaultIV  = []byte{0x56, 0x2e, 0x17, 0x99, 0x6d, 0x09, 0x3d, 0x28, 0xdd, 0xb3, 0xba, 0x69, 0x5a, 0x2e, 0x6f, 0x58}

	mu                 = new(sync.Mutex)
	conn               *net.UDPConn
	sourceAddr         *net.UDPAddr
	sourceHardwareAddr *net.HardwareAddr
	packetNumber       uint16
)

type DiscoveryResponse struct {
	Type uint16
	MAC  net.HardwareAddr
	IP   net.IP
	Port int
}

type AuthorizationResponse struct {
	DeviceID []byte
	Key      []byte
}

func init() {
	mu.Lock()
	defer mu.Unlock()

	listenAddr, err := net.ResolveUDPAddr("udp4", "0.0.0.0:0")
	if err != nil {
		panic(err)
	}

	conn, err = network.GetReceiverConn(listenAddr, nil)
	if err != nil {
		panic(err)
	}

	srcAddr, err := net.ResolveUDPAddr("udp4", conn.LocalAddr().String())
	if err != nil {
		panic(err)
	}

	sourceAddr = srcAddr

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		_ = <-sigs
		_ = conn.Close()
	}()

	ifaces, err := net.Interfaces()
	if err != nil {
		panic(err)
	}

	sourceHardwareAddr = &net.HardwareAddr{0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

loop:
	for _, iface := range ifaces {
		ifaceAddrs, err := iface.Addrs()
		if err != nil {
			panic(err)
		}

		for _, ifaceAddr := range ifaceAddrs {
			ipNet := ifaceAddr.(*net.IPNet).IP
			if ipNet == nil {
				continue
			}

			if ifaceAddr.(*net.IPNet).IP.To4().Equal(sourceAddr.IP.To4()) {
				sourceHardwareAddr = &iface.HardwareAddr
				break loop
			}
		}
	}

	packetNumber = uint16(rand.Uint32())
}

func encrypt(plainText []byte, key []byte) (cipherText []byte, err error) {
	requiredPadding := aes.BlockSize - (len(plainText) % aes.BlockSize)
	if requiredPadding != 0 {
		for i := 0; i < requiredPadding; i++ {
			plainText = append(plainText, 0x00)
		}
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	cipherText = make([]byte, aes.BlockSize+len(plainText))

	cbc := cipher.NewCBCEncrypter(block, DefaultIV)
	cbc.CryptBlocks(cipherText[aes.BlockSize:], plainText)

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

	cipherText = cipherText[aes.BlockSize:]

	cbc := cipher.NewCBCDecrypter(block, DefaultIV)
	cbc.CryptBlocks(cipherText, cipherText)

	plainText = cipherText

	return
}

func setPadding(payload []byte, start int, end int) {
	for i := start; i < end+1; i++ {
		payload[i] = 0x00
	}
}

func setBytes(payload []byte, bytes []byte, start int) {
	for i := start; i < len(bytes); i++ {
		payload[i] = bytes[i-start]
	}
}

func getChecksum(payload []byte) int {
	checkSum := 0xbeaf
	for i := 0; i < len(payload); i++ {
		checkSum += int(payload[i])

		if checkSum > 0xffff {
			checkSum -= 0xffff
		}
	}

	return checkSum
}

// Discover returns a discovery response for each found device (ref.: https://github.com/mjg59/python-broadlink/blob/master/protocol.md#network-discovery)
func Discover() (discoveryResponses []DiscoveryResponse, err error) {
	mu.Lock()
	defer mu.Unlock()

	discoveryResponses = make([]DiscoveryResponse, 0)

	srcAddr, err := net.ResolveUDPAddr("udp4", conn.LocalAddr().String())
	if err != nil {
		panic(err)
	}

	sourceAddr = srcAddr

	payload := make([]byte, 48)

	// padding
	setPadding(payload, 0x00, 0x07)

	// UTC offset
	now := time.Now()
	_, rawOffset := now.Zone()
	offsetHours := rawOffset / 3600
	binary.LittleEndian.PutUint32(payload[0x08:0x0b+1], uint32(offsetHours))

	// year
	binary.LittleEndian.PutUint16(payload[0x0c:0x0d+1], uint16(now.Year()))

	// seconds past the minute
	payload[0x0e] = uint8(now.Second())

	// minutes past the hour
	payload[0x0f] = uint8(now.Minute())

	// hours past midnight
	payload[0x10] = uint8(now.Hour())

	// day of the week
	payload[0x11] = uint8(now.Weekday())

	// day of the month
	payload[0x12] = uint8(now.Day())

	// month
	payload[0x13] = uint8(now.Month())

	// padding
	setPadding(payload, 0x14, 0x17)

	// source IP
	setBytes(payload, sourceAddr.IP.To4()[0:4], 0x18)

	// source port
	binary.LittleEndian.PutUint16(payload[0x1c:0x1d+1], uint16(sourceAddr.Port))

	// padding
	setPadding(payload, 0x1e, 0x1f)

	// checksum
	binary.LittleEndian.PutUint16(payload[0x20:0x21+1], uint16(getChecksum(payload)))

	// padding
	setPadding(payload, 0x22, 0x25)

	// not sure
	payload[0x26] = 0x06

	// padding
	setPadding(payload, 0x27, 0x2f)

	dstAddr, err := net.ResolveUDPAddr("udp4", "255.255.255.255:80")
	if err != nil {
		return
	}

	err = conn.SetWriteDeadline(time.Now().Add(timeout))
	if err != nil {
		return
	}

	_, err = conn.WriteTo(payload, dstAddr)
	if err != nil {
		return
	}

	err = conn.SetReadDeadline(time.Now().Add(timeout))
	if err != nil {
		return
	}

	for {
		response := make([]byte, 64)
		_, receiveAddr, err := conn.ReadFromUDP(response)
		if err != nil { // a timeout means probably no more responses are coming
			break
		}

		macBytes := make([]byte, 0)

		for i := 0x3f; i >= 0x3a; i-- {
			macBytes = append(macBytes, response[i])
		}

		mac := net.HardwareAddr(macBytes)

		discoveryResponses = append(
			discoveryResponses,
			DiscoveryResponse{
				Type: binary.LittleEndian.Uint16(response[0x34 : 0x35+1]),
				MAC:  mac,
				IP:   receiveAddr.IP,
				Port: receiveAddr.Port,
			},
		)
	}

	return
}

// Command sends a payload to the given device and returns the response bytes (ref.: https://github.com/mjg59/python-broadlink/blob/master/protocol.md#command-packet-format)
func Command(deviceType uint16, commandCode uint16, deviceID []byte, deviceIP net.IP, devicePort int, plainTextPayload []byte, key []byte) (response []byte, err error) {
	mu.Lock()
	defer mu.Unlock()

	response = make([]byte, 0)

	header := make([]byte, 56)

	// magic
	header[0x00] = 0x5a
	header[0x01] = 0xa5
	header[0x02] = 0xaa
	header[0x03] = 0x55
	header[0x04] = 0x5a
	header[0x05] = 0xa5
	header[0x06] = 0xaa
	header[0x07] = 0x55

	// padding
	setPadding(header, 0x08, 0x1f)

	// checksum at 0x20 -> 0x21 later

	// padding
	setPadding(header, 0x22, 0x23)

	// device type
	binary.LittleEndian.PutUint16(header[0x24:0x25+1], deviceType)

	// command code
	binary.LittleEndian.PutUint16(header[0x26:0x27+1], commandCode)

	// packet packetNumber
	binary.LittleEndian.PutUint16(header[0x28:0x29+1], packetNumber)
	packetNumber++

	// local MAC
	setBytes(header, (*sourceHardwareAddr)[0:6], 0x2a)

	// device ID
	setBytes(header, deviceID[0:4], 0x30)

	// padding
	setPadding(header, 0x36, 0x37)

	// checksum of unencrypted plainTextPayload
	binary.LittleEndian.PutUint16(header[0x34:0x35+1], uint16(getChecksum(plainTextPayload)))

	cipherTextPayload, err := encrypt(plainTextPayload, key)
	if err != nil {
		return
	}

	request := append(header, cipherTextPayload...)

	// checksum of the whole request
	binary.LittleEndian.PutUint16(plainTextPayload[0x20:0x21+1], uint16(getChecksum(request)))

	dstAddr := net.UDPAddr{
		IP:   deviceIP,
		Port: devicePort,
	}

	log.Printf("sending %v bytes of %#+v to %#+v", len(request), request, dstAddr)

	err = conn.SetWriteDeadline(time.Now().Add(timeout))
	if err != nil {
		return
	}

	_, err = conn.WriteTo(request, &dstAddr)
	if err != nil {
		return
	}

	err = conn.SetReadDeadline(time.Now().Add(timeout))
	if err != nil {
		return
	}

	var n int
	for {
		b := make([]byte, 65536)

		n, _, err = conn.ReadFromUDP(b)
		if err != nil {
			err = nil
			break
		}

		if n == 0 {
			break
		}

		response = append(response, b[0:n]...)
	}

	log.Printf("received %v bytes of %#+v", len(response), response)

	return
}

// Authorize gets an authorization response for the given device (ref.: https://github.com/mjg59/python-broadlink/blob/master/protocol.md#command-packet-format)
func Authorize(deviceType uint16, deviceIP net.IP, devicePort int) (authorizationResponse AuthorizationResponse, err error) {
	payload := make([]byte, 80)

	// padding
	setPadding(payload, 0x00, 0x03)

	// this device identifier
	setBytes(payload, []byte("initialed85!!!1"), 0x04)

	// not sure
	payload[0x13] = 0x01

	// padding
	setPadding(payload, 0x14, 0x2c)

	// not sure
	payload[0x2d] = 0x01

	// this device name
	deviceName := []byte("github.com/initialed85")
	deviceName = append(deviceName, 0x00)
	setBytes(payload, deviceName, 0x30)

	response, err := Command(deviceType, 0x65, []byte{0x00, 0x00, 0x00, 0x00}, deviceIP, devicePort, payload, DefaultKey)
	if err != nil {
		return
	}

	if len(response) <= 0x38 {
		err = fmt.Errorf("reponse not long enough at %v bytes (need at least 0x38)", len(response))
		return
	}

	responseCipherText := response[0x38:]

	responsePlainText, err := decrypt(responseCipherText, DefaultKey)
	if err != nil {
		return
	}

	log.Print(responsePlainText)

	return
}
