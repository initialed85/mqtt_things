package smart_aircons_client

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

var TestMode = false
var TestURL string
var HTTPClient = &http.Client{
	Timeout: time.Second * 5,
}

func enableTestMode(client *http.Client, url string) {
	TestMode = true
	TestURL = url
	HTTPClient = client
}

func SendIR(host, code string) error {
	log.Printf("sending %v to %v", code, host)

	url := fmt.Sprintf("http://%v/uuid", host)

	log.Printf("getting uuid from %v", url)

	if TestMode {
		url = TestURL
	}

	resp, err := HTTPClient.Get(url)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	parts := strings.Split(strings.TrimSpace(string(body)), ",")
	if len(parts) != 2 {
		return fmt.Errorf("expected []string of length 2, got %v", parts)
	}

	uuid := parts[1]

	log.Printf("got uuid %v", uuid)

	buf := []byte(fmt.Sprintf("sendir,%v", code))

	url = fmt.Sprintf("http://%v/v2/%v", host, uuid)
	if TestMode {
		url = TestURL
	}

	log.Printf("sending %v to %v", code, url)

	resp, err = http.Post(
		url,
		"text/plain",
		bytes.NewBuffer(buf),
	)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	return nil
}
