package smart_aircons_client

import (
	"fmt"
	"log"
	"sync"
	"time"
)

const defaultOn bool = false
const defaultMode string = "fan_only"
const defaultTemperature int64 = 24
const debounceDuration = time.Millisecond * 100

type Model struct {
	mu          sync.Mutex
	calls       int64
	on          bool
	mode        string // "off", "cool", "heat", "fan_only"
	temperature int64
	setState    func(on bool, mode string, temperature int64) error
}

func NewModel(
	setState func(on bool, mode string, temperature int64) error,
) *Model {
	a := Model{
		calls:       0,
		on:          defaultOn,
		mode:        defaultMode,
		temperature: defaultTemperature,
		setState:    setState,
	}

	return &a
}

func (a *Model) debouncedSetState() {
	a.calls++
	calls := a.calls

	go func() {
		time.Sleep(debounceDuration)

		a.mu.Lock()
		defer a.mu.Unlock()

		takeAction := calls == a.calls

		if !takeAction {
			log.Printf("state superseded (we're %#+v, now at %#+v); taking no action", calls, a.calls)
			return
		}

		err := a.setState(a.on, a.mode, a.temperature)
		if err != nil {
			log.Printf("warning: attempt to setState(%#+v, %#+v, %#+v) failed because: %v", a.on, a.mode, a.temperature, err)
		}
	}()
}

func (a *Model) SetOn(on bool) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if on == a.on {
		log.Printf("on already %#+v; no change", on)
		return nil
	}

	a.on = on

	a.debouncedSetState()

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

	a.mode = mode

	a.debouncedSetState()

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

	a.temperature = temperature

	a.debouncedSetState()

	return nil
}
