package mqtt_action_router

import (
	"fmt"
	"github.com/initialed85/mqtt_things/pkg/mqtt_client"
	"log"
	"strconv"
	"sync"
	"time"
)

type State int64

const (
	Unknown State = -1
	Off     State = 0
	On      State = 1
)

type action struct {
	setTopic  string
	arguments interface{}
	on        func(interface{}) error
	off       func(interface{}) error
	baseState State
	debounce  time.Duration
	client    mqtt_client.Client
	getTopic  string
	mutex     sync.Mutex
}

func parseBinaryState(payload string) (State, error) {
	state, err := strconv.ParseInt(payload, 10, 64)
	if err != nil {
		return Unknown, err
	}

	if state != 0 && state != 1 {
		return Unknown, fmt.Errorf("failed to parse payload of %v to binary state", payload)
	}

	return State(state), nil
}

func newAction(setTopic string, arguments interface{}, on func(interface{}) error, off func(interface{}) error, debounce time.Duration, client mqtt_client.Client, baseState State, getTopic string) action {
	action := action{
		setTopic:  setTopic,
		arguments: arguments,
		on:        on,
		off:       off,
		baseState: baseState,
		debounce:  debounce,
		client:    client,
		getTopic:  getTopic,
	}

	if setTopic == getTopic {
		log.Fatalf("setTopic %v is the same as getTopic %v; this could cause loops", setTopic, getTopic)
	}

	log.Printf("created action %+v", action)

	return action
}

func (a *action) actuate(state State) error {
	log.Printf("actuate called with state %+v; grabbing lock", state)

	a.mutex.Lock()
	defer a.mutex.Unlock()

	var actuate func(interface{}) error

	if state == Off {
		actuate = a.off
	} else if state == On {
		actuate = a.on
	} else {
		return fmt.Errorf("expected state of 0 or 1 but got %v", state)
	}

	log.Printf("calling actuate with %+v", a.arguments)
	err := actuate(a.arguments)

	go func() {
		payload := fmt.Sprintf("%v", state)
		log.Printf("publishing %v to %v ", payload, a.getTopic)
		err := a.client.Publish(a.getTopic, mqtt_client.ExactlyOnce, true, payload)
		if err != nil {
			log.Printf("failed to publish %v to %v because %v", payload, a.getTopic, err)
		}
	}()

	log.Printf("debouncing for %+v, lock will be released", a.debounce)
	time.Sleep(a.debounce)

	return err
}

func (a *action) setup() error {
	log.Printf("setup called for %v, establishing base state of %v", a.setTopic, a.baseState)
	err := a.actuate(a.baseState)
	if err != nil {
		return err
	}

	log.Printf("subscribing to %v", a.setTopic)
	err = a.client.Subscribe(a.setTopic, mqtt_client.ExactlyOnce, a.callback)
	if err != nil {
		return err
	}

	return err
}

func (a *action) handleBinaryState(incomingPayload string) error {
	state, err := parseBinaryState(incomingPayload)
	if err != nil {
		return err
	}

	return a.actuate(state)
}

func (a *action) callback(message mqtt_client.Message) {
	log.Printf("callback for %v called with %+v", a.setTopic, message)

	err := a.handleBinaryState(message.Payload)
	if err != nil {
		log.Printf("handleBinaryState for %+v caused %+v", message, err)
	}
}

func (a *action) teardown() error {
	log.Printf("teardown called for %v, establishing base state of %v", a.setTopic, a.baseState)
	actuateErr := a.actuate(a.baseState)

	log.Printf("unsubscribing from %v", a.setTopic)
	mqttErr := a.client.Unsubscribe(a.setTopic)

	if actuateErr != nil && mqttErr != nil {
		return fmt.Errorf("actuate caused %+v and unsubscribe caused %+v", actuateErr, mqttErr)
	} else if actuateErr != nil {
		return actuateErr
	} else if mqttErr != nil {
		return mqttErr
	}

	return nil
}

type Router struct {
	client          mqtt_client.Client
	debounce        time.Duration
	actions         map[string]action
	actionsMapMutex sync.Mutex
	actionsMutex    sync.Mutex
	useActionsMutex bool
}

func New(client mqtt_client.Client, debounce time.Duration, allowConcurrentActions bool) Router {
	router := Router{
		client:          client,
		debounce:        debounce,
		useActionsMutex: !allowConcurrentActions,
		actions:         make(map[string]action),
	}

	log.Printf("created router %v", router)

	return router
}

func (a *Router) RemoveAction(setTopic string) error {
	log.Printf("removing action for %v", setTopic)

	a.actionsMapMutex.Lock()
	defer a.actionsMapMutex.Unlock()

	action, ok := a.actions[setTopic]
	if ok {
		err := action.teardown()
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("no action for topic %v", setTopic)
	}

	return nil
}

func (a *Router) RemoveAllActions() error {
	log.Printf("removing all actions")

	a.actionsMapMutex.Lock()
	defer a.actionsMapMutex.Unlock()

	var errors []error
	for _, action := range a.actions {
		err := action.teardown()
		if err == nil {
			continue
		}

		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("action teardowns caused some errors: %+v", errors)
	}

	return nil
}

func (a *Router) AddAction(setTopic string, arguments interface{}, on func(interface{}) error, off func(interface{}) error, baseState State, getTopic string) error {
	log.Printf("adding action for %v, arguments are %+v, on func is %p, off func is %p", setTopic, arguments, on, off)

	a.actionsMapMutex.Lock()
	defer a.actionsMapMutex.Unlock()

	_, ok := a.actions[setTopic]
	if ok {
		return fmt.Errorf("action for topic %v already exists", setTopic)
	}

	action := newAction(setTopic, arguments, on, off, a.debounce, a.client, baseState, getTopic)

	err := action.setup()
	if err != nil {
		return err
	}

	a.actions[setTopic] = action

	return nil
}
