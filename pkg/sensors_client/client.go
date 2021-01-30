package sensors_client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

var TestMode = false
var TestURL string
var HTTPClient = &http.Client{
	Timeout: time.Second * 5,
}

type State struct {
	Presence    bool `json:"presence"`
	LightLevel  int  `json:"lightlevel"`
	Dark        bool `json:"dark"`
	Daylight    bool `json:"daylight"`
	Temperature int  `json:"temperature"`
}

type Sensor struct {
	ID       int64
	State    State  `json:"state"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	UniqueID string `json:"uniqueid"`
}

type Response map[string]Sensor

type BlendedSensor struct {
	ID          int64
	Name        string
	Presence    bool
	LightLevel  int
	Dark        bool
	Daylight    bool
	Temperature float64
}

func enableTestMode(client *http.Client, url string) {
	TestMode = true
	TestURL = url
	HTTPClient = client
}

func get(url string) ([]byte, error) {
	if TestMode {
		url = TestURL
	}

	resp, err := HTTPClient.Get(url)
	if err != nil {
		return []byte{}, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}

	log.Printf("%v", string(body))

	return body, nil
}

func getSensors(host string, appID string) ([]BlendedSensor, error) {
	sensorsJSON, err := get(fmt.Sprintf(
		"http://%v/api/%v/sensors",
		host,
		appID,
	))
	if err != nil {
		return []BlendedSensor{}, err
	}

	sensorByID := make(map[string]Sensor)
	err = json.Unmarshal(sensorsJSON, &sensorByID)
	if err != nil {
		return []BlendedSensor{}, err
	}

	sensorsByAddress := make(map[string][]Sensor)
	for idString, sensor := range sensorByID {
		id, err := strconv.ParseInt(idString, 10, 64)
		if err != nil {
			return []BlendedSensor{}, err
		}

		sensor.ID = id

		address := strings.Split(sensor.UniqueID, "-")[0]

		_, ok := sensorsByAddress[address]
		if !ok {
			sensorsByAddress[address] = make([]Sensor, 0)
		}

		sensorsByAddress[address] = append(sensorsByAddress[address], sensor)
	}

	blendedSensors := make([]BlendedSensor, 0)

	for _, sensors := range sensorsByAddress {
		blendedSensor := BlendedSensor{}
		a := false
		b := false
		c := false
		for _, sensor := range sensors {
			switch sensor.Type {
			case "ZLLPresence":
				blendedSensor.ID = sensor.ID
				blendedSensor.Name = sensor.Name
				blendedSensor.Presence = sensor.State.Presence
				a = true
			case "ZLLLightLevel":
				blendedSensor.LightLevel = sensor.State.LightLevel
				blendedSensor.Dark = sensor.State.Dark
				blendedSensor.Daylight = sensor.State.Daylight
				b = true
			case "ZLLTemperature":
				blendedSensor.Temperature = float64(sensor.State.Temperature) / 100
				c = true
			}
		}

		if !(a && b && c) {
			continue
		}

		blendedSensors = append(blendedSensors, blendedSensor)
	}

	sort.Slice(
		blendedSensors,
		func(i, j int) bool {
			return blendedSensors[i].ID < blendedSensors[j].ID
		},
	)

	return blendedSensors, nil
}

type Client struct {
	host  string
	appID string
}

func New(host string, appID string) Client {
	return Client{
		host:  host,
		appID: appID,
	}
}

func (c *Client) GetSensors() ([]BlendedSensor, error) {
	return getSensors(c.host, c.appID)
}
