package aircons_client

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

func TestNew(t *testing.T) {
	ts := getTestHTTPServer()
	defer ts.Close()

	enableTestMode(ts.Client, ts.URL)

	hostAndNameAndCodesName := []HostAndNameAndCodesName{
		{
			Host:      "192.168.137.20",
			Name:      "fujitsu_1",
			CodesName: "fujitsu",
		},
	}

	client, err := New(hostAndNameAndCodesName)
	if err != nil {
		log.Fatal(err)
	}

	a, err := client.GetAircon("fujitsu_1")
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < 16; i++ {
		err = a.On()
		if err != nil {
			log.Fatal(err)
		}
	}

	assert.Equal(t, nil, err)
	assert.Equal(t, 4, ts.CallCount)

	ts.CallCount = 0

	for i := 0; i < 16; i++ {
		err = a.Off()
		if err != nil {
			log.Fatal(err)
		}
	}

	assert.Equal(t, nil, err)
	assert.Equal(t, 2, ts.CallCount)
}
