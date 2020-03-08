package aircon_client

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

const OnCode string = "1:1,0,37000,1,1,122,62,15,16,15,16,15,46,15,16,15,46,15,16,15,16,15,16,15,46,15,46,15,16,15,16,15,16,15,46,15,46,15,16,15,16,14,16,15,16,15,16,14,16,15,16,15,16,14,16,15,16,15,16,15,16,15,16,15,46,15,16,15,16,15,16,15,16,15,16,14,16,15,16,15,46,15,16,15,16,15,16,15,16,15,16,15,46,15,46,15,46,15,46,15,46,15,46,15,16,15,16,15,16,14,47,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,46,15,46,15,16,15,16,15,47,14,16,15,16,15,16,15,16,15,46,15,16,15,16,15,46,15,16,15,16,15,16,15,16,15,16,14,16,15,16,15,46,15,46,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,14,16,15,16,15,16,14,16,15,16,15,16,14,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,46,15,46,15,16,15,46,15,16,15,46,15,16,15,46,15,3692"
const OffCode string = "1:1,0,37000,1,1,122,62,15,16,15,16,15,46,15,16,15,46,15,16,14,16,15,16,15,46,15,47,14,16,15,16,15,16,14,47,15,47,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,46,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,46,15,16,15,16,15,16,15,16,15,46,15,16,15,16,15,16,15,16,15,16,14,16,15,3692"

var TestMode = false
var TestURL string
var DefaultClient = http.DefaultClient

func EnableTestMode(client *http.Client, url string) {
	TestMode = true
	TestURL = url
	DefaultClient = client
}

func sendIR(host, code string) error {
	var url string

	url = fmt.Sprintf("http://%v/uuid", host)
	if TestMode {
		url = TestURL
	}

	resp, err := DefaultClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	parts := strings.Split(strings.TrimSpace(string(body)), ",")
	if len(parts) != 2 {
		return fmt.Errorf("expected []string of length 2, got %v", parts)
	}

	uuid := parts[1]

	buf := []byte(fmt.Sprintf("sendir,%v", code))

	url = fmt.Sprintf("http://%v/v2/%v", host, uuid)
	if TestMode {
		url = TestURL
	}

	resp, err = http.Post(
		url,
		"text/plain",
		bytes.NewBuffer(buf),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func TurnOn(host string) error {
	return sendIR(host, OnCode)
}

func TurnOff(host string) error {
	return sendIR(host, OffCode)
}

type Client struct {
	host             string
	firstInteraction bool
	on               bool
}

func New(host string) Client {
	return Client{
		host:             host,
		firstInteraction: true,
	}
}

func (a *Client) TurnOn() error {
	if !a.firstInteraction {
		if a.on {
			return nil
		}
	} else {
		a.firstInteraction = false
	}

	a.on = true
	return TurnOn(a.host)
}

func (a *Client) TurnOff() error {
	if !a.firstInteraction {
		if !a.on {
			return nil
		}
	} else {
		a.firstInteraction = false
	}

	a.on = false
	return TurnOff(a.host)
}
