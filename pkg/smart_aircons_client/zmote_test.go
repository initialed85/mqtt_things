package smart_aircons_client

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestSendIR(t *testing.T) {
	ts := getTestHTTPServer()
	defer ts.Close()

	enableTestMode(ts.Client, ts.URL)

	err := SendIR("some_host", "some_code")
	handleError(err)

	assert.Equal(t, 2, ts.CallCount)
}
