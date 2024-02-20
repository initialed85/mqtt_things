package aircons_client

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type State int64

const (
	Unknown State = -1
	Off     State = 0
	On      State = 1
)

type Codes struct {
	Name   string
	OnAt18 string
	OnAt23 string
	Off    string
}

var codesByName = map[string]Codes{
	"fujitsu": {
		Name:   "fujitsu",
		OnAt18: "1:1,0,37000,1,1,122,62,15,16,15,16,15,46,15,16,15,46,15,16,15,16,15,16,15,46,15,46,15,16,15,16,15,16,15,46,15,46,15,16,15,16,14,16,15,16,15,16,14,16,15,16,15,16,14,16,15,16,15,16,15,16,15,16,15,46,15,16,15,16,15,16,15,16,15,16,14,16,15,16,15,46,15,16,15,16,15,16,15,16,15,16,15,46,15,46,15,46,15,46,15,46,15,46,15,16,15,16,15,16,14,47,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,46,15,46,15,16,15,16,15,47,14,16,15,16,15,16,15,16,15,46,15,16,15,16,15,46,15,16,15,16,15,16,15,16,15,16,14,16,15,16,15,46,15,46,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,14,16,15,16,15,16,14,16,15,16,15,16,14,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,46,15,46,15,16,15,46,15,16,15,46,15,16,15,46,15,3692",
		OnAt23: "1:1,0,37000,1,1,121,62,15,16,15,16,15,46,15,16,15,46,15,16,15,16,15,16,15,46,15,47,14,16,15,16,15,16,14,47,15,47,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,46,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,46,15,16,15,16,15,16,15,16,15,16,14,47,15,47,15,46,15,46,15,46,15,46,15,16,15,16,15,16,15,46,15,16,15,16,15,16,15,16,15,16,15,16,15,16,14,16,15,47,14,47,15,16,15,16,15,46,15,16,15,16,15,16,14,47,15,47,15,46,15,16,15,47,14,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,14,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,46,15,46,15,46,15,46,15,16,15,46,15,16,15,3692",
		Off:    "1:1,0,37000,1,1,122,62,15,16,15,16,15,46,15,16,15,46,15,16,14,16,15,16,15,46,15,47,14,16,15,16,15,16,14,47,15,47,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,46,15,16,15,16,15,16,15,16,15,16,15,16,15,16,15,46,15,16,15,16,15,16,15,16,15,46,15,16,15,16,15,16,15,16,15,16,14,16,15,3692",
	},
	"mitsubishi": {
		Name:   "mitsubishi",
		OnAt18: "1:1,9935,38000,1,1,131,65,16,49B16,16CCBCCBBCBCCBBCBBCCBCCBCCCCCCCCCCCCCCCCCCCCBCCCCCBBCCCCCCCCCCCCBBCBBCCCCCCCCCBCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCBBCCCCCC16,490ABBCCCBCCBBCBCCBBCBBCCBCCBCCCCCCCCCCCCCCCCCCCCBCCCCCBBCCCCCCCCCCCCBBCBBCCCCCCCCCBCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCBBCCCCCC16,760",
		OnAt23: "1:1,8867,38000,1,1,131,65,16,49B16,16CCBCCBBCBCCBBCBBCCBCCBCCCCCCCCCCCCCCCCCCCCBCCCCCBBCCCBBBCCCCCCBBCBBCCCCCCCCCBCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCBCBCCCC16,490ABBCCCBCCBBCBCCBBCBBCCBCCBCCCCCCCCCCCCCCCCCCCCBCCCCCBBCCCBBBCCCCCCBBCBBCCCCCCCCCBCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCBCBCCCC16,760",
		Off:    "1:1,9255,38000,1,1,131,65,16,49B16,16CCBCCBBCBCCBBCBBCCBCCBCCCCCCCCCCCCCCCCCCCCCCCCCCBBCCCCCCCCCCCCBBCBBCCCCCCCCCBCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCBBCCCBBB16,490ABBCCCBCCBBCBCCBBCBBCCBCCBCCCCCCCCCCCCCCCCCCCCCCCCCCBBCCCCCCCCCCCCBBCBBCCCCCCCCCBCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCBBCCCBBB16,760",
	},
}

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

func sendIR(host, code string) error {
	var url string

	log.Printf("sending %v to %v", code, host)

	url = fmt.Sprintf("http://%v/uuid", host)

	log.Printf("getting uuid from %v", url)

	if TestMode {
		url = TestURL
	}

	resp, err := HTTPClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
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

type Aircon struct {
	Name             string
	State            State
	codes            Codes
	host             string
	firstInteraction bool
}

func NewAircon(host, name, codesName string) (Aircon, error) {
	codesName = strings.ToLower(strings.TrimSpace(codesName))
	codes, ok := codesByName[codesName]
	if !ok {
		possibleCodesNames := make([]string, 0)
		for possibleCodesName := range codesByName {
			possibleCodesNames = append(possibleCodesNames, possibleCodesName)
		}

		return Aircon{}, fmt.Errorf("failed to find codes for %v; options are %v", codesName, possibleCodesNames)
	}

	return Aircon{
		Name:             name,
		State:            Unknown,
		codes:            codes,
		host:             host,
		firstInteraction: true,
	}, nil
}

func (a *Aircon) On() error {
	log.Printf("setting %v:%v:%v to On", a.Name, a.codes.Name, a.host)

	if !a.firstInteraction {
		if a.State == On {
			log.Printf("already on")
			return nil
		}
	} else {
		a.firstInteraction = false
	}

	err := sendIR(a.host, a.codes.OnAt23)
	if err != nil {
		return err
	}

	a.State = On

	return nil
}

func (a *Aircon) Off() error {
	log.Printf("setting %v:%v:%v to Off", a.Name, a.codes.Name, a.host)

	if !a.firstInteraction {
		if a.State == Off {
			log.Printf("already off")
			return nil
		}
	} else {
		a.firstInteraction = false
	}

	err := sendIR(a.host, a.codes.Off)
	if err != nil {
		return err
	}

	a.State = Off

	return nil
}

type HostAndNameAndCodesName struct {
	Host      string
	Name      string
	CodesName string
}

type Client struct {
	airconByName map[string]*Aircon
}

func New(hostsAndNamesAndCodesNames []HostAndNameAndCodesName) (Client, error) {
	airconByName := make(map[string]*Aircon)

	for _, hostAndNameAndCodesName := range hostsAndNamesAndCodesNames {
		aircon, err := NewAircon(
			hostAndNameAndCodesName.Host,
			hostAndNameAndCodesName.Name,
			hostAndNameAndCodesName.CodesName,
		)
		if err != nil {
			return Client{}, err
		}

		// err = aircon.Off()
		// if err != nil {
		// 	return Client{}, err
		// }

		airconByName[hostAndNameAndCodesName.Name] = &aircon
	}

	return Client{
		airconByName: airconByName,
	}, nil
}

func (c *Client) GetAircons() ([]*Aircon, error) {
	aircons := make([]*Aircon, 0)
	for _, a := range c.airconByName {
		aircons = append(aircons, a)
	}

	return aircons, nil
}

func (c *Client) GetAircon(name string) (*Aircon, error) {
	a, ok := c.airconByName[name]
	if !ok {
		return nil, fmt.Errorf("failed to find switch for name %v", name)
	}

	return a, nil
}
