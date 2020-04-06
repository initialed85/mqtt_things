package relays_client

import (
	"fmt"
	"github.com/tarm/serial"
	"log"
	"strings"
	"sync"
	"time"
)

var TestMode = false
var TestPortInstance TestPort

func enableTestMode() {
	TestMode = true
}

type PortInterface interface {
	Read([]byte) (int, error)
	Write([]byte) (int, error)
	Close() error
}

type TestPort struct {
	Config     *serial.Config
	ReadCursor int
	ReadData   []byte
	WriteData  []byte
}

func OpenTestPort(config *serial.Config) (*TestPort, error) {
	TestPortInstance.Config = config

	return &TestPortInstance, nil
}

func (tp *TestPort) Read(buf []byte) (int, error) {
	if tp.ReadCursor >= len(tp.ReadData) {
		return -1, fmt.Errorf("out of ReadData at %v", tp.ReadCursor)
	}

	buf[0] = tp.ReadData[tp.ReadCursor]

	tp.ReadCursor++

	return len(buf), nil
}

func (tp *TestPort) Write(buf []byte) (int, error) {
	TestPortInstance.WriteData = buf

	return len(buf), nil
}

func (tp *TestPort) Close() error {
	return nil
}

func Read(port PortInterface) (string, error) {
	log.Printf("reading from %+v", port)

	buf := make([]byte, 1)

	_, err := port.Read(buf)
	if err != nil {
		return string(buf), err
	}

	return string(buf), nil
}

func ReadUntil(port PortInterface, until string) (string, error) {
	log.Printf("reading until %v from %+v", until, port)

	var data string

	for {
		if strings.Contains(data, until) {
			break
		}

		s, err := Read(port)
		if err != nil {
			return data, err
		}

		data += s
	}

	return data, nil
}

func Write(port PortInterface, data string) error {
	log.Printf("writing %v to %+v", data, port)

	_, err := port.Write([]byte(data))

	return err
}

func Close(port PortInterface) error {
	log.Printf("closing %+v", port)

	return port.Close()
}

const Delimiter = "relay %v change to state %v\r\n"

type Client struct {
	port PortInterface
	mu   sync.Mutex
}

func New(port string, bitRate int) (Client, error) {
	r := Client{}

	var p PortInterface
	var err error

	c := serial.Config{
		Name:        port,
		Baud:        bitRate,
		ReadTimeout: time.Second,
	}

	if !TestMode {
		p, err = serial.OpenPort(&c)
	} else {
		p, err = OpenTestPort(&c)
	}

	if err != nil {
		return r, err
	}

	r.port = p

	log.Printf("created %+v", r)

	return r, nil
}

func (c *Client) setState(relay int64, state string) error {
	log.Printf("setting relay %v to %v", relay, state)

	c.mu.Lock()
	defer c.mu.Unlock()

	err := Write(c.port, fmt.Sprintf("%v,%v\r\n", relay, state))
	if err != nil {
		return err
	}

	delimiter := fmt.Sprintf(Delimiter, relay, state)

	data, err := ReadUntil(c.port, delimiter)
	if err != nil {
		return err
	}

	if !strings.HasSuffix(data, delimiter) {
		return fmt.Errorf(
			"insane response; expected \"%v\" but got \"%v\"",
			delimiter,
			data,
		)
	}

	// some debounce for the relay
	time.Sleep(time.Second)

	return nil
}

func (c *Client) On(relay int64) error {
	return c.setState(relay, "on")
}

func (c *Client) Off(relay int64) error {
	return c.setState(relay, "off")

}

func (c *Client) Close() error {
	return Close(c.port)
}
