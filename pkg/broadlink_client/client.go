package broadlink_client

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net"
	"runtime"
	"slices"
	"strings"
	"sync"
	"time"

	binarypack "github.com/canhlinh/go-binary-pack"
	"github.com/initialed85/glue/pkg/network"
)

var errorMessages = map[int16]string{
	// Firmware-related errors
	-1:  "Authentication failed",
	-2:  "You have been logged out",
	-3:  "The device is offline",
	-4:  "Command not supported",
	-5:  "The device storage is full",
	-6:  "Structure is abnormal",
	-7:  "Control key is expired",
	-8:  "Send error",
	-9:  "Write error",
	-10: "Read error",
	-11: "SSID could not be found in AP configuration",
}

var (
	defaultKey = []byte{0x09, 0x76, 0x28, 0x34, 0x3f, 0xe9, 0x9e, 0x23, 0x76, 0x5c, 0x15, 0x13, 0xac, 0xcf, 0x8b, 0x02}
	defaultIV  = []byte{0x56, 0x2e, 0x17, 0x99, 0x6d, 0x09, 0x3d, 0x28, 0xdd, 0xb3, 0xba, 0x69, 0x5a, 0x2e, 0x6f, 0x58}
)

type RequestID struct {
	DstAddr        string
	SequenceNumber int16
}

type Response struct {
	RequestID  RequestID
	SrcAddr    *net.UDPAddr
	Payload    []byte
	Err        error
	ReceivedAt time.Time
}

type Request struct {
	ID        RequestID
	Payload   []byte
	DstAddr   *net.UDPAddr
	Responses chan *Response
	SentAt    time.Time
}

type Client struct {
	ctx                context.Context
	cancel             context.CancelFunc
	mu                 *sync.Mutex
	conn               *net.UDPConn
	sourceAddr         *net.UDPAddr
	sourceHardwareAddr *net.HardwareAddr
	uhandledRequests   chan *Request
	requestByRequestID map[RequestID]*Request
	sequenceNumber     int16
}

func NewClient() (c *Client, err error) {
	c = &Client{
		mu:                 new(sync.Mutex),
		uhandledRequests:   make(chan *Request, 1024),
		requestByRequestID: make(map[RequestID]*Request),
		sequenceNumber:     int16(rand.Float32() * math.MaxInt16),
	}

	c.ctx, c.cancel = context.WithCancel(context.Background())

	listenAddr, err := net.ResolveUDPAddr("udp4", "0.0.0.0:0")
	if err != nil {
		return nil, err
	}

	c.conn, err = network.GetReceiverConn(listenAddr, nil)
	if err != nil {
		return nil, err
	}

	c.sourceAddr, err = net.ResolveUDPAddr("udp4", c.conn.LocalAddr().String())
	if err != nil {
		return nil, err
	}

	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	c.sourceHardwareAddr = &net.HardwareAddr{0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

loop:
	for _, iface := range ifaces {
		ifaceAddrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}

		for _, ifaceAddr := range ifaceAddrs {
			ipNet := ifaceAddr.(*net.IPNet).IP
			if ipNet == nil {
				continue
			}

			if ifaceAddr.(*net.IPNet).IP.To4().Equal(c.sourceAddr.IP.To4()) {
				c.sourceHardwareAddr = &iface.HardwareAddr
				break loop
			}
		}
	}

	go func() {
		defer func() {
			r := recover()
			if r != nil {
				log.Printf("warning: recovered from panic: %v", r)
			}
		}()

		for {
			select {
			case <-c.ctx.Done():
				log.Printf("warning: read goroutine canceled")
				return
			default:
			}

			c.mu.Lock()
			conn := c.conn
			c.mu.Unlock()

			if conn == nil {
				log.Printf("warning: read goroutine found conn to be nil; skipping...")
				time.Sleep(time.Second * 1)
				continue
			}

			b := make([]byte, 1024)

			c.conn.SetReadDeadline(time.Now().Add(time.Second * 5))
			n, srcAddr, err := c.conn.ReadFromUDP(b)
			if err != nil {

				if strings.Contains(err.Error(), "i/o timeout") {
					continue
				}

				log.Printf("warning: failed to read from conn: %v; skipping...", err)
				time.Sleep(time.Second * 1)
				continue
			}

			payload := b[:n]

			if len(payload) < 0x2a {
				log.Printf("warning: message not long enough to contain sequence number")
				continue
			}

			unpackedSequenceNumber, err := new(binarypack.BinaryPack).UnPack([]string{"h"}, payload[0x28:0x2a])
			if err != nil {
				log.Printf("warning: failed to unpack sequence number: %v", err)
				continue
			}
			sequenceNumber := int16(unpackedSequenceNumber[0].(int))

			c.mu.Lock()
			request := c.requestByRequestID[RequestID{
				DstAddr:        srcAddr.String(),
				SequenceNumber: sequenceNumber,
			}]

			if request == nil {
				request = c.requestByRequestID[RequestID{
					DstAddr:        fmt.Sprintf("255.255.255.255:%v", srcAddr.Port),
					SequenceNumber: sequenceNumber,
				}]
			}
			c.mu.Unlock()

			if request == nil {
				log.Printf(
					"warning: failed to find request for %v bytes from %v (sequence number %v)",
					len(payload), srcAddr.String(), sequenceNumber,
				)
				continue
			}

			response := &Response{
				RequestID:  request.ID,
				SrcAddr:    srcAddr,
				Payload:    payload,
				Err:        nil,
				ReceivedAt: time.Now(),
			}

			errorCode, err := unpackShort(payload[0x22:0x24])
			if err != nil {
				log.Printf("warning: failed to unpack sequence number: %v", err)
				continue
			}

			if errorCode != 0 {
				errorMessage := errorMessages[errorCode]
				if errorMessage == "" {
					errorMessage = "Unknown error"
				}

				response.Err = fmt.Errorf(
					"received error code %v (%#+v); %v",
					errorCode, payload[0x22:0x24], errorMessage,
				)
			}

			if request.Responses == nil {
				if response.Err != nil {
					log.Printf("warning: response specified error but requester did not want response; %v", err)
				}
				continue
			}

			request.Responses <- response
		}
	}()
	runtime.Gosched()

	go func() {
		for {
			select {
			case <-c.ctx.Done():
				log.Printf("warning: write goroutine canceled")
				return
			case request := <-c.uhandledRequests:
				c.mu.Lock()
				conn := c.conn
				c.mu.Unlock()

				if conn == nil {
					log.Printf("warning: write goroutine found conn to be nil; skipping...")
					time.Sleep(time.Second * 1)
					continue
				}

				c.mu.Lock()
				c.requestByRequestID[request.ID] = request
				c.mu.Unlock()

				_, err = c.conn.WriteToUDP(request.Payload, request.DstAddr)
				if err != nil {
					if request.Responses != nil {
						request.Responses <- &Response{
							RequestID:  request.ID,
							SrcAddr:    nil,
							Payload:    nil,
							Err:        err,
							ReceivedAt: time.Now(),
						}
					}
					continue
				}
			}
		}
	}()
	runtime.Gosched()

	return c, nil
}

func (c *Client) Close() error {
	c.cancel()

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return fmt.Errorf("cannot close, already closed")
	}

	_ = c.conn.Close()
	c.conn = nil

	return nil
}

func (c *Client) getNextSequenceNumber() int16 {
	c.mu.Lock()
	sequenceNumber := c.sequenceNumber
	c.sequenceNumber++
	if c.sequenceNumber <= 0 {
		c.sequenceNumber = 0
	}
	c.mu.Unlock()

	return sequenceNumber
}

func (c *Client) getRequest(dstAddr *net.UDPAddr, payload []byte, sequenceNumber int16, responses chan *Response) *Request {
	return &Request{
		ID: RequestID{
			DstAddr:        dstAddr.String(),
			SequenceNumber: sequenceNumber,
		},
		Payload:   payload,
		DstAddr:   dstAddr,
		Responses: responses,
		SentAt:    time.Now(),
	}
}

func (c *Client) send(request *Request) error {
	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()

	if conn == nil {
		return fmt.Errorf("send found conn to be unexpectedly nil; client dead?")
	}

	select {
	case c.uhandledRequests <- request:
	default:
		return fmt.Errorf("unhandled request channel unexpectedly full")
	}

	return nil
}

func (c *Client) doCall(request *Request, timeout time.Duration) (*Response, error) {
	err := c.send(request)
	if err != nil {
		return nil, err
	}

	select {
	case <-time.After(timeout):
		break
	case response := <-request.Responses:
		return response, nil
	}

	return nil, fmt.Errorf("call timed out waiting for response after %v", timeout)
}

func (c *Client) Discover(timeout time.Duration) ([]*Device, error) {
	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()

	if conn == nil {
		return nil, fmt.Errorf("discover found conn to be unexpectedly nil; client dead?")
	}

	dstAddr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%v:%v", "255.255.255.255", "80"))
	if err != nil {
		return nil, err
	}

	deadline := time.Now().Add(timeout)

	responses := make(chan *Response, 65536)

	for time.Now().Before(deadline) {
		sleepUntil := time.Now().Add(time.Millisecond * 500)

		sequenceNumber := c.getNextSequenceNumber()

		payload, err := getDiscoveryPayload(time.Now(), c.sourceAddr, sequenceNumber)
		if err != nil {
			return nil, err
		}

		err = c.send(c.getRequest(dstAddr, payload, sequenceNumber, responses))
		if err != nil {
			return nil, err
		}

		time.Sleep(time.Until(sleepUntil))
	}

	devices := make([]*Device, 0)

drain:
	for {
		select {
		case response := <-responses:
			if response.Payload == nil {
				return nil, fmt.Errorf("response payload unexpectedly empty; socket failed?")
			}

			rawMac := response.Payload[0x3a:0x40]
			slices.Reverse(rawMac)
			mac := net.HardwareAddr(rawMac)

			if response.Err != nil {
				return nil, response.Err
			}

			device := Device{
				mu:       new(sync.Mutex),
				client:   c,
				Name:     string(bytes.Split(response.Payload[0x40:], []byte{0x00})[0]),
				Type:     binary.LittleEndian.Uint16(response.Payload[0x34:0x36]),
				MAC:      mac,
				Addr:     *response.SrcAddr,
				LastSeen: response.ReceivedAt,
			}

			for _, otherDevice := range devices {
				if otherDevice.MAC.String() == device.MAC.String() {
					continue drain
				}
			}

			devices = append(devices, &device)
		default:
			break drain
		}
	}

	return devices, nil
}
