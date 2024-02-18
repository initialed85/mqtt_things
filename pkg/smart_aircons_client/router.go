package smart_aircons_client

import (
	"fmt"
	"log"
	"strings"
	"sync"

	mqtt "github.com/initialed85/mqtt_things/pkg/mqtt_client"
)

const (
	topicSetSuffix        = "set"
	topicGetSuffix        = "get"
	topicOnInfix          = "power"
	topicModeInfix        = "mode"
	topicTemperatureInfix = "temperature"
)

type Router struct {
	mu sync.Mutex

	topicPrefix string

	invokeCallbacksForGets bool
	onHandler              func(bool) error
	modeHandler            func(string) error
	temperatureHandler     func(int64) error
}

func NewRouter(
	topicPrefix string,
	onHandler func(bool) error,
	modeHandler func(string) error,
	temperatureHandler func(int64) error,
) *Router {
	r := Router{
		topicPrefix:            strings.TrimRight(topicPrefix, "/") + "/",
		invokeCallbacksForGets: false,
		onHandler:              onHandler,
		modeHandler:            modeHandler,
		temperatureHandler:     temperatureHandler,
	}

	return &r
}

func (r *Router) handleOn(payload interface{}) {
	on, err := PayloadToOn(payload.(string))
	if err != nil {
		log.Printf("warning: ignoring %#+v for %#+v topic because: %v", payload, "on", err)
		return
	}

	log.Printf("invoking %#+v handler with %#+v", "on", on)
	err = r.onHandler(on)
	if err != nil {
		log.Printf("warning: failed to invoke %#+v handler with %#+v becuase: %v", "on", on, err)
	}
}

func (r *Router) handleMode(payload interface{}) {
	mode, err := PayloadToMode(payload.(string))
	if err != nil {
		log.Printf("warning: ignoring %#+v for %#+v topic because: %v", payload, "mode", err)
		return
	}

	log.Printf("invoking %#+v handler with %#+v", "mode", mode)
	err = r.modeHandler(mode)
	if err != nil {
		log.Printf("warning: failed to invoke %#+v handler with %#+v becuase: %v", "mode", mode, err)
	}
}

func (r *Router) handleTemperature(payload interface{}) {
	temperature, err := PayloadToTemperature(payload.(string))
	if err != nil {
		log.Printf("warning: ignoring %#+v for %#+v topic because: %v", payload, "temperature", err)
		return
	}

	log.Printf("invoking %#+v handler with %#+v", "temperature", temperature)
	err = r.temperatureHandler(temperature)
	if err != nil {
		log.Printf("warning: failed to invoke %#+v handler with %#+v becuase: %v", "temperature", temperature, err)
	}
}

func (r *Router) EnableGetCallbacks() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.invokeCallbacksForGets = true
}

func (r *Router) DisableGetCallbacks() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.invokeCallbacksForGets = false
}

func (r *Router) IsGetCallbacks() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.invokeCallbacksForGets
}

func (r *Router) Handle(message mqtt.Message) (mqtt.Message, bool) {
	log.Printf("handling %#+v for %#+v", message.Payload, message.Topic)

	if !strings.HasPrefix(message.Topic, r.topicPrefix) {
		log.Printf("warning: ignoring message with unrecognized topic %#+v", message.Topic)
		return mqtt.Message{}, false
	}

	infixAndSuffix := strings.Split(message.Topic[len(r.topicPrefix):], "/")
	if len(infixAndSuffix) != 2 {
		log.Printf("warning: ignoring message with unexpected sub-topic pattern %#+v", infixAndSuffix)
		return mqtt.Message{}, false
	}

	var isGet bool

	// identify if this was a /get or /set message
	suffix := strings.ToLower(strings.TrimSpace(infixAndSuffix[1]))
	if suffix == topicGetSuffix {
		isGet = true
	} else if suffix == topicSetSuffix {
		isGet = false
	} else {
		log.Printf("warning: ignoring message with unexpected suffix %#+v", suffix)
		return mqtt.Message{}, false
	}

	// only handle /get messages in special cases (restore state from broker on startup)
	if isGet {
		r.mu.Lock()
		invokeCallbacks := r.invokeCallbacksForGets
		r.mu.Unlock()

		if !invokeCallbacks {
			return mqtt.Message{}, false
		}
	}

	// route the message
	infix := strings.ToLower(strings.TrimSpace(infixAndSuffix[0]))

	if infix == topicOnInfix {
		r.handleOn(message.Payload)
	} else if infix == topicModeInfix {
		r.handleMode(message.Payload)
	} else if infix == topicTemperatureInfix {
		r.handleTemperature(message.Payload)
	} else {
		log.Printf("warning: ignoring message with unexpected infix %#+v", infix)
		return mqtt.Message{}, false
	}

	if isGet {
		return mqtt.Message{}, false
	}

	return mqtt.Message{
		Topic:   fmt.Sprintf("%v%v/%v", r.topicPrefix, infix, topicGetSuffix),
		Payload: message.Payload,
	}, true
}
