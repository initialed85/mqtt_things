package lights_client

import (
	"fmt"
	"github.com/amimof/huego"
	"log"
	"strings"
)

type State int64

const (
	Unknown State = -1
	Off     State = 0
	On      State = 1
)

type Light struct {
	Name     string
	State    State
	hueLight huego.Light
}

func formatLightName(lightName string) string {
	return strings.ReplaceAll(strings.ToLower(lightName), " ", "-")
}

func newLight(hueLight huego.Light) Light {
	state := Unknown
	if hueLight.State.On {
		state = On
	} else {
		state = Off
	}

	light := Light{
		Name:     formatLightName(hueLight.Name),
		State:    state,
		hueLight: hueLight,
	}

	log.Printf("created %+v from %+v", light, hueLight)

	return light
}

func (l *Light) On() error {
	log.Printf("requesting on for %v", l.Name)

	if l.hueLight.IsOn() {
		log.Printf("%v already on", l.Name)
	}

	err := l.hueLight.On()
	if err != nil {
		return err
	}

	return nil
}

func (l *Light) Off() error {
	log.Printf("requesting off for %v", l.Name)

	if !l.hueLight.IsOn() {
		log.Printf("%v already off", l.Name)
	}

	err := l.hueLight.Off()
	if err != nil {
		return err
	}

	return nil
}

type Client struct {
	bridge *huego.Bridge
}

func New(bridgeHost, appName, apiKey string) Client {
	client := Client{
		bridge: huego.New(bridgeHost, appName).Login(apiKey),
	}

	log.Printf("created %+v", client)

	return client
}

func (c *Client) GetLights() ([]Light, error) {
	hueLights, err := c.bridge.GetLights()
	if err != nil {
		return []Light{}, err
	}

	var lights []Light
	for _, hueLight := range hueLights {
		lights = append(lights, newLight(hueLight))
	}

	log.Printf("built %+v from %+v", lights, hueLights)

	return lights, nil
}

func (c *Client) GetLight(name string) (Light, error) {
	log.Printf("getting light for %v", name)

	lights, err := c.GetLights()
	if err != nil {
		return Light{}, err
	}

	for _, light := range lights {
		// TODO: implicit behaviour is a little gross
		if light.Name == name || light.Name == formatLightName(name) {
			log.Printf("found %+v for %v", light, light)
			return light, nil
		}
	}

	return Light{}, fmt.Errorf("failed to find light for name %v", name)
}
