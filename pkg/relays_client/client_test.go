package relays_client

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tarm/serial"
)

const TestMessage = "relay 3 change to state on\r\n"

func TestNew(t *testing.T) {
	enableTestMode()

	r, err := New("/dev/ttyACM0", 9600)
	if err != nil {
		log.Fatal(err)
	}
	assert.Nil(t, err)
	assert.Equal(
		t,
		&serial.Config{
			Name:        "/dev/ttyACM0",
			Baud:        9600,
			ReadTimeout: 1000000000,
		},
		TestPortInstance.Config,
	)

	TestPortInstance.ReadData = []byte(TestMessage)
	err = r.On(3)
	if err != nil {
		log.Fatal(err)
	}
	assert.Nil(t, err)

	err = r.Close()
	if err != nil {
		log.Fatal(err)
	}
	assert.Nil(t, err)
}
