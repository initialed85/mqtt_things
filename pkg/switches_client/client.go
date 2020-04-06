package switches_client

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

type State int64

const (
	urlPrefix       = "http://"
	urlSuffix       = "/cm?cmnd=Power"
	getOn           = `{"POWER":"ON"}`
	getOff          = `{"POWER":"OFF"}`
	setOn           = "%20On"
	setOff          = "%20Off"
	Unknown   State = -1
	Off       State = 0
	On        State = 1
)

var TestMode = false
var TestURL string
var NetClient = &http.Client{
	Timeout: time.Second * 5,
}

func enableTestMode(client *http.Client, url string) {
	TestMode = true
	TestURL = url
	NetClient = client
}

func get(url string) (string, error) {
	if TestMode {
		url = TestURL
	}

	resp, err := NetClient.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	bodyString := strings.TrimSpace(string(body))

	log.Printf("body is %v", bodyString)

	return bodyString, nil
}

func getState(host string) (State, error) {
	url := fmt.Sprintf("%v%v%v", urlPrefix, host, urlSuffix)

	log.Printf("getting state from %v", url)

	body, err := get(url)
	if err != nil {
		return Unknown, err
	}

	if body == getOn {
		return On, nil
	} else if body == getOff {
		return Off, nil
	}

	return Unknown, fmt.Errorf("unable to interpret state from '%v'", body)
}

func on(host string) error {
	url := fmt.Sprintf("%v%v%v%v", urlPrefix, host, urlSuffix, setOn)

	log.Printf("setting on state for %v", url)

	body, err := get(url)
	if err != nil {
		return err
	}

	if body != getOn {
		return fmt.Errorf("failed to interpret on state from '%v'", body)
	}

	return nil
}

func off(host string) error {
	url := fmt.Sprintf("%v%v%v%v", urlPrefix, host, urlSuffix, setOff)

	log.Printf("setting on state for %v", url)

	body, err := get(url)
	if err != nil {
		return err
	}

	if body != getOff {
		return fmt.Errorf("failed to interpret off state from '%v'", body)
	}

	return nil
}

type Switch struct {
	Name  string
	State State
	host  string
}

func NewSwitch(host, name string) Switch {
	return Switch{
		Name:  name,
		State: Unknown,
		host:  host,
	}
}

func (s *Switch) Update() error {
	log.Printf("getting state for %v:%v", s.Name, s.host)

	state, err := getState(s.host)
	if err != nil {
		return err
	}

	s.State = state

	log.Printf("state for %v:%v is %v", s.Name, s.host, s.State)

	return nil
}

func (s *Switch) On() error {
	log.Printf("setting %v:%v to On", s.Name, s.host)
	err := on(s.host)
	if err != nil {
		return err
	}

	return s.Update()
}

func (s *Switch) Off() error {
	log.Printf("setting %v:%v to Off", s.Name, s.host)
	err := off(s.host)
	if err != nil {
		return err
	}

	return s.Update()
}

type Client struct {
	switchByName map[string]Switch
}

func New(hostByName map[string]string) Client {
	switchByName := make(map[string]Switch)

	for name, host := range hostByName {
		switchByName[name] = NewSwitch(host, name)
	}

	return Client{
		switchByName: switchByName,
	}
}

func (c *Client) GetSwitches() ([]Switch, error) {
	switches := make([]Switch, 0)
	for _, s := range c.switchByName {
		err := s.Update()
		if err != nil {
			return []Switch{}, err
		}
		switches = append(switches, s)
	}

	return switches, nil
}

func (c *Client) GetSwitch(name string) (Switch, error) {
	s, ok := c.switchByName[name]
	if !ok {
		return Switch{}, fmt.Errorf("failed to find switch for name %v", name)
	}

	return s, nil
}
