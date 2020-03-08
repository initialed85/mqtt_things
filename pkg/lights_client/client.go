package lights_client

import (
	"fmt"
	"github.com/amimof/huego"
	"strings"
)

type State int64

const (
	Unknown State = -1
	Off     State = 0
	On      State = 1
)

type Client struct {
	bridge *huego.Bridge
}

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

	return Light{
		Name:     formatLightName(hueLight.Name),
		State:    state,
		hueLight: hueLight,
	}
}

func (l *Light) On() error {
	err := l.hueLight.On()
	if err != nil {
		return err
	}

	return nil
}

func (l *Light) Off() error {
	err := l.hueLight.Off()
	if err != nil {
		return err
	}

	return nil
}

func New(bridgeHost, appName, apiKey string) Client {
	return Client{
		bridge: huego.New(bridgeHost, appName).Login(apiKey),
	}
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

	return lights, nil
}

func (c *Client) GetLight(name string) (Light, error) {
	lights, err := c.GetLights()
	if err != nil {
		return Light{}, err
	}

	for _, light := range lights {
		// TODO: implicit behaviour is a little gross
		if light.Name == name || light.Name == formatLightName(name) {
			return light, nil
		}
	}

	return Light{}, fmt.Errorf("failed to find light for name %v", name)
}
