package aircon_client

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

type TestHTTPServer struct {
	CallCount int
	Client    *http.Client
	Close     func()
	URL       string
}

func (t *TestHTTPServer) Handle(w http.ResponseWriter, r *http.Request) {
	responseBody := "uuid,CI123a1234"

	w.Header().Set("Content-Type", "text/plain")
	_, err := fmt.Fprintln(w, responseBody)
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

func TestTurnOn(t *testing.T) {
	testHTTPServer := getTestHTTPServer()
	defer testHTTPServer.Close()

	EnableTestMode(testHTTPServer.Client, testHTTPServer.URL)

	err := TurnOn("192.168.137.20")
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, nil, err)
	assert.Equal(t, 2, testHTTPServer.CallCount)
}

func TestTurnOff(t *testing.T) {
	testHTTPServer := getTestHTTPServer()
	defer testHTTPServer.Close()

	EnableTestMode(testHTTPServer.Client, testHTTPServer.URL)

	err := TurnOff("192.168.137.20")
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, nil, err)
	assert.Equal(t, 2, testHTTPServer.CallCount)
}

func TestNew(t *testing.T) {
	ts := getTestHTTPServer()
	defer ts.Close()

	EnableTestMode(ts.Client, ts.URL)

	a := New("192.168.137.20")

	var err error

	for i := 0; i < 16; i++ {
		err = a.TurnOn()
		if err != nil {
			log.Fatal(err)
		}
	}

	assert.Equal(t, nil, err)
	assert.Equal(t, 2, ts.CallCount)

	ts.CallCount = 0

	for i := 0; i < 16; i++ {
		err = a.TurnOff()
		if err != nil {
			log.Fatal(err)
		}
	}

	assert.Equal(t, nil, err)
	assert.Equal(t, 2, ts.CallCount)
}
