package switches_client

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	testResponseBody = ""
)

type TestHTTPServer struct {
	CallCount int
	Client    *http.Client
	Close     func()
	URL       string
}

func (t *TestHTTPServer) Handle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	_, err := fmt.Fprintln(w, testResponseBody)
	if err != nil {
		log.Fatal(err)
	}

	t.CallCount++
}

func getTestHTTPServer() *TestHTTPServer {
	t := TestHTTPServer{}

	s := httptest.NewServer(http.HandlerFunc(t.Handle))
	t.Client = s.Client()
	t.Close = s.Close
	t.URL = s.URL

	return &t
}

func TestOnOffAndGetState(t *testing.T) {
	testHTTPServer := getTestHTTPServer()
	defer testHTTPServer.Close()

	enableTestMode(testHTTPServer.Client, testHTTPServer.URL)

	testResponseBody = getOn
	err := on("192.168.137.15")
	if err != nil {
		log.Fatal(err)
	}
	assert.Equal(t, 1, testHTTPServer.CallCount)

	state, err := getState("192.168.137.15")
	if err != nil {
		log.Fatal(err)
	}
	assert.Equal(t, 2, testHTTPServer.CallCount)
	assert.Equal(t, On, state)

	testResponseBody = getOff
	err = off("192.168.137.15")
	if err != nil {
		log.Fatal(err)
	}
	assert.Equal(t, 3, testHTTPServer.CallCount)

	state, err = getState("192.168.137.15")
	if err != nil {
		log.Fatal(err)
	}
	assert.Equal(t, 4, testHTTPServer.CallCount)
	assert.Equal(t, Off, state)
}

func TestNew(t *testing.T) {
	testHTTPServer := getTestHTTPServer()
	defer testHTTPServer.Close()

	enableTestMode(testHTTPServer.Client, testHTTPServer.URL)

	client := New([]HostAndName{
		{
			Host: "192.168.137.15",
			Name: "tuya_1",
		},
		{
			Host: "192.168.137.16",
			Name: "tuya_2",
		},
	})

	testResponseBody = getOn
	s, err := client.GetSwitch("tuya_1")
	if err != nil {
		log.Fatal(err)
	}
	err = s.On()
	if err != nil {
		log.Fatal(err)
	}
	assert.Equal(t, 2, testHTTPServer.CallCount)
	assert.Equal(t, On, s.State)

	testResponseBody = getOff
	s, err = client.GetSwitch("tuya_2")
	if err != nil {
		log.Fatal(err)
	}
	err = s.Off()
	if err != nil {
		log.Fatal(err)
	}
	assert.Equal(t, 4, testHTTPServer.CallCount)
	assert.Equal(t, Off, s.State)

	switches, err := client.GetSwitches()
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(
		t,
		&Switch{Name: "tuya_1", State: 0, host: "192.168.137.15"},
		switches[0],
	)

	assert.Equal(
		t,
		&Switch{Name: "tuya_2", State: 0, host: "192.168.137.16"},
		switches[1],
	)
}
