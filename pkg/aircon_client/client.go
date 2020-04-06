package aircon_client

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

const OnCode string = "1:1,0,37000,1,1,122,62,15,16,15,16,15,46,15,16,15,46,15,16,15,16,15,16,15,46,15,46,15,16,15,16,15,16,15,46,15,46,15,16,15,16,14,16,15,16,15,16,14,16,15,16,15,16,14,16,15,16,15,16,15,16,15,16,15,46,15,16,15,16,15,16,15,16,15,16,14,16,15,16,15,46,15,16,15,16,15,16,15,16,15,16,15,46,15,46,15,46,15,46,15,46,15,46,15,16,15,16,15,16,14,47,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,46,15,46,15,16,15,16,15,47,14,16,15,16,15,16,15,16,15,46,15,16,15,16,15,46,15,16,15,16,15,16,15,16,15,16,14,16,15,16,15,46,15,46,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,14,16,15,16,15,16,14,16,15,16,15,16,14,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,46,15,46,15,16,15,46,15,16,15,46,15,16,15,46,15,3692"
const OffCode string = "1:1,0,37000,1,1,122,62,15,16,15,16,15,46,15,16,15,46,15,16,14,16,15,16,15,46,15,47,14,16,15,16,15,16,14,47,15,47,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,46,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,46,15,16,15,16,15,16,15,16,15,46,15,16,15,16,15,16,15,16,15,16,14,16,15,3692"

var TestMode = false
var TestURL string
var NetClient = &http.Client{
	Timeout: time.Second * 5,
}

func EnableTestMode(client *http.Client, url string) {
	TestMode = true
	TestURL = url
	NetClient = client
}

func sendIR(host, code string) error {
	var url string

	log.Printf("sending %v to %v", code, host)

	url = fmt.Sprintf("http://%v/uuid", host)

	log.Printf("getting uuid from %v", url)

	if TestMode {
		url = TestURL
	}

	resp, err := NetClient.Get(url)
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
	defer resp.Body.Close()

	return nil
}

func on(host string) error {
	return sendIR(host, OnCode)
}

func off(host string) error {
	return sendIR(host, OffCode)
}

type Client struct {
	host             string
	firstInteraction bool
	on               bool
}

func New(host string) Client {
	client := Client{
		host:             host,
		firstInteraction: true,
	}

	log.Printf("created %+v", client)

	return client
}

func (a *Client) On() error {
	log.Printf("on requested")
	if !a.firstInteraction {
		if a.on {
			log.Printf("already on")

			return nil
		}
	} else {
		a.firstInteraction = false
	}

	a.on = true
	return on(a.host)
}

func (a *Client) Off() error {
	log.Printf("off requested")
	if !a.firstInteraction {
		if !a.on {
			log.Printf("already off")

			return nil
		}
	} else {
		a.firstInteraction = false
	}

	a.on = false
	return off(a.host)
}
