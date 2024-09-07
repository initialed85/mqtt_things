package smart_aircons_client

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

var HTTPClient = &http.Client{
	Timeout: time.Second * 5,
}

func ZmoteSendIR(host string, rawCode []byte) error {
	host = strings.ToLower(strings.TrimSpace(host))

	code := strings.TrimSpace(string(rawCode))

	if strings.Contains(code, "\n") {
		parts := strings.Split(code, "\n")
		code = strings.TrimSpace(parts[len(parts)-1])
	}

	log.Printf("sending %#+v to zmote %v", code, host)

	url := fmt.Sprintf("http://%v/uuid", host)

	resp, err := HTTPClient.Get(url)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	parts := strings.Split(strings.TrimSpace(string(body)), ",")
	if len(parts) != 2 {
		return fmt.Errorf("expected []string of length 2, got %v", parts)
	}

	uuid := parts[1]

	buf := []byte(code)

	url = fmt.Sprintf("http://%v/v2/%v", host, uuid)

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

func ZmoteLearnIR(host string) ([]byte, error) {
	host = strings.ToLower(strings.TrimSpace(host))

	url := fmt.Sprintf("http://%v/uuid", host)

	resp, err := HTTPClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(strings.TrimSpace(string(body)), ",")
	if len(parts) != 2 {
		return nil, fmt.Errorf("expected []string of length 2, got %v", parts)
	}

	uuid := parts[1]

	buf := []byte("get_IRL")

	url = fmt.Sprintf("http://%v/v2/%v", host, uuid)

	resp, err = http.Post(
		url,
		"text/plain",
		bytes.NewBuffer(buf),
	)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("wanted %v, got %v", http.StatusOK, resp.StatusCode)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if len(b) <= 0 {
		return nil, fmt.Errorf("wanted non-empty response body, got %#+v", string(b))
	}

	return b, nil
}
