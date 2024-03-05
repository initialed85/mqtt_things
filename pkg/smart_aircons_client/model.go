package smart_aircons_client

import (
	"fmt"
	"log"
	"sync"
)

const defaultOn bool = false
const defaultMode string = "fan_only"
const defaultTemperature int64 = 24

type Model struct {
	mu          sync.Mutex
	on          bool
	mode        string // "off", "cool", "heat", "fan_only"
	temperature int64
	setState    func(on bool, mode string, temperature int64) error
}

func NewModel(
	setState func(on bool, mode string, temperature int64) error,
) *Model {
	a := Model{
		on:          defaultOn,
		mode:        defaultMode,
		temperature: defaultTemperature,
		setState:    setState,
	}

	return &a
}

func (a *Model) SetOn(on bool) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if on == a.on {
		log.Printf("on already %#+v; no change", on)
		return nil
	}

	err := a.setState(on, a.mode, a.temperature)
	if err != nil {
		return fmt.Errorf("warning: attempt to setState(%#+v, %#+v, %#+v) failed because: %v", on, a.mode, a.temperature, err)
	}

	a.on = on

	return nil
}

func (a *Model) SetMode(mode string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if mode != "off" && mode != "cool" && mode != "heat" && mode != "fan_only" {
		return fmt.Errorf("%#+v not one of %#+v, %#+v, %#+v or %#+v", mode, "off", "cool", "heat", "fan_only")
	}

	if mode == a.mode {
		log.Printf("mode already %#+v; no change", mode)
		return nil
	}

	on := mode != "off"

	err := a.setState(on, mode, a.temperature)
	if err != nil {
		return fmt.Errorf("warning: attempt to setState(%#+v, %#+v, %#+v) failed because: %v", a.on, mode, a.temperature, err)
	}

	a.on = on
	a.mode = mode

	return nil
}

func (a *Model) SetTemperature(temperature int64) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if temperature < 18 || temperature > 30 {
		return fmt.Errorf("%#+v out of range 18 - 30 inclusive", temperature)
	}

	if temperature == a.temperature {
		log.Printf("temperature already %#+v; no change", temperature)
		return nil
	}

	err := a.setState(true, a.mode, temperature)
	if err != nil {
		return fmt.Errorf("warning: attempt to setState(%#+v, %#+v, %#+v) failed because: %v", a.on, a.mode, temperature, err)
	}

	a.temperature = temperature
	a.on = true

	return nil
}

func (a *Model) GetState() (bool, string, int64) {
	a.mu.Lock()
	a.mu.Unlock()

	return a.on, a.mode, a.temperature
}
