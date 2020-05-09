package mqtt_action_router

import (
	"github.com/initialed85/mqtt_things/pkg/mqtt_client"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
	"time"
)

const (
	setTopic1 = "some_action_topic_1/set"
	getTopic1 = "some_action_topic_1/get"
	setTopic2 = "some_action_topic_2/set"
	getTopic2 = "some_action_topic_2/get"
)

type ActuatableThing struct {
	Index int
	IsOn  bool
}

type ActuatableThingArguments struct {
	Index int
}

func (a *ActuatableThing) On(arguments interface{}) error {
	actuatableThingArguments := arguments.(ActuatableThingArguments)

	a.Index = actuatableThingArguments.Index
	a.IsOn = true

	return nil
}

func (a *ActuatableThing) Off(arguments interface{}) error {
	actuatableThingArguments := arguments.(ActuatableThingArguments)

	a.Index = actuatableThingArguments.Index
	a.IsOn = false

	return nil
}

var (
	lastPayload1     string
	lastPayload2     string
	actuatableThing1 ActuatableThing
	actuatableThing2 ActuatableThing
)

func callback1(message mqtt_client.Message) {
	lastPayload1 = message.Payload
}

func callback2(message mqtt_client.Message) {
	lastPayload2 = message.Payload
}

func setupActionTest(state State) *action {
	actuatableThing1.Index = 0
	actuatableThing1.IsOn = false

	// NOTE: requires a local eclipse-mosquitto Docker container to be running
	client := mqtt_client.New("127.0.0.1", "", "")

	err := client.Connect()
	if err != nil {
		log.Fatal(err)
	}

	err = client.Subscribe(getTopic1, mqtt_client.ExactlyOnce, callback1)
	if err != nil {
		log.Fatal(err)
	}

	return newAction(
		setTopic1,
		ActuatableThingArguments{Index: 6291},
		actuatableThing1.On,
		actuatableThing1.Off,
		time.Millisecond*100,
		client,
		state,
		getTopic1,
	)
}

func teardownActionTest(action *action) {
	err := action.client.Disconnect()
	if err != nil {
		log.Fatal(err)
	}
}

func TestAction_setup_Off(t *testing.T) {
	action := setupActionTest(Off)

	assert.Equal(t, false, actuatableThing1.IsOn)
	assert.Equal(t, 0, actuatableThing1.Index)

	err := action.setup()
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(time.Millisecond * 100)

	assert.Equal(t, false, actuatableThing1.IsOn)
	assert.Equal(t, 6291, actuatableThing1.Index)
	assert.Equal(t, "0", lastPayload1)

	teardownActionTest(action)
}

func TestAction_setup_On(t *testing.T) {
	action := setupActionTest(On)

	assert.Equal(t, false, actuatableThing1.IsOn)
	assert.Equal(t, 0, actuatableThing1.Index)

	err := action.setup()
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(time.Millisecond * 100)

	assert.Equal(t, true, actuatableThing1.IsOn)
	assert.Equal(t, 6291, actuatableThing1.Index)
	assert.Equal(t, "1", lastPayload1)

	teardownActionTest(action)
}

func TestAction_handleBinaryState_Off(t *testing.T) {
	action := setupActionTest(Off)
	err := action.setup()
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, false, actuatableThing1.IsOn)
	assert.Equal(t, 6291, actuatableThing1.Index)

	err = action.handleBinaryState("0")
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(time.Millisecond * 100)

	assert.Equal(t, false, actuatableThing1.IsOn)
	assert.Equal(t, 6291, actuatableThing1.Index)
	assert.Equal(t, "0", lastPayload1)

	teardownActionTest(action)
}

func TestAction_handleBinaryState_On(t *testing.T) {
	action := setupActionTest(Off)
	err := action.setup()
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, false, actuatableThing1.IsOn)
	assert.Equal(t, 6291, actuatableThing1.Index)

	err = action.handleBinaryState("1")
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(time.Millisecond * 100)

	assert.Equal(t, true, actuatableThing1.IsOn)
	assert.Equal(t, 6291, actuatableThing1.Index)
	assert.Equal(t, "1", lastPayload1)

	teardownActionTest(action)
}

func TestAction_callback_Off(t *testing.T) {
	action := setupActionTest(Off)
	err := action.setup()
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(time.Millisecond * 100)

	assert.Equal(t, false, actuatableThing1.IsOn)
	assert.Equal(t, 6291, actuatableThing1.Index)

	err = action.client.Publish(setTopic1, mqtt_client.ExactlyOnce, false, "0")
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(time.Second)

	assert.Equal(t, false, actuatableThing1.IsOn)
	assert.Equal(t, 6291, actuatableThing1.Index)
	assert.Equal(t, "0", lastPayload1)

	teardownActionTest(action)
}

func TestAction_callback_On(t *testing.T) {
	action := setupActionTest(Off)
	err := action.setup()
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(time.Millisecond * 100)

	assert.Equal(t, false, actuatableThing1.IsOn)
	assert.Equal(t, 6291, actuatableThing1.Index)

	err = action.client.Publish(setTopic1, mqtt_client.ExactlyOnce, false, "1")
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(time.Second)

	assert.Equal(t, true, actuatableThing1.IsOn)
	assert.Equal(t, 6291, actuatableThing1.Index)
	assert.Equal(t, "1", lastPayload1)

	teardownActionTest(action)
}

func TestAction_Teardown_Off(t *testing.T) {
	action := setupActionTest(Off)

	assert.Equal(t, false, actuatableThing1.IsOn)
	assert.Equal(t, 0, actuatableThing1.Index)

	err := action.teardown()
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(time.Millisecond * 100)

	assert.Equal(t, false, actuatableThing1.IsOn)
	assert.Equal(t, 6291, actuatableThing1.Index)
	assert.Equal(t, "0", lastPayload1)

	teardownActionTest(action)
}

func TestAction_Teardown_On(t *testing.T) {
	action := setupActionTest(On)

	assert.Equal(t, false, actuatableThing1.IsOn)
	assert.Equal(t, 0, actuatableThing1.Index)

	err := action.teardown()
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(time.Millisecond * 100)

	assert.Equal(t, true, actuatableThing1.IsOn)
	assert.Equal(t, 6291, actuatableThing1.Index)
	assert.Equal(t, "1", lastPayload1)

	teardownActionTest(action)
}

func setupActionRouterTest() *Router {
	actuatableThing1.Index = 0
	actuatableThing1.IsOn = false

	actuatableThing2.Index = 0
	actuatableThing2.IsOn = false

	// NOTE: requires a local eclipse-mosquitto Docker container to be running
	client := mqtt_client.New("127.0.0.1", "", "")

	err := client.Connect()
	if err != nil {
		log.Fatal(err)
	}

	err = client.Subscribe(getTopic1, mqtt_client.ExactlyOnce, callback1)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Subscribe(getTopic2, mqtt_client.ExactlyOnce, callback2)
	if err != nil {
		log.Fatal(err)
	}

	return New(
		client,
		time.Millisecond*10,
		false,
	)
}

func teardownActionRouterTest(actionRouter *Router) {
	err := actionRouter.client.Disconnect()
	if err != nil {
		log.Fatal(err)
	}
}

func TestActionRouter_Everything(t *testing.T) {
	actionRouter := setupActionRouterTest()

	err := actionRouter.AddAction(
		setTopic1,
		ActuatableThingArguments{Index: 6291},
		actuatableThing1.On,
		actuatableThing1.Off,
		Off,
		getTopic1,
	)
	if err != nil {
		log.Fatal(err)
	}

	err = actionRouter.AddAction(
		setTopic2,
		ActuatableThingArguments{Index: 6291},
		actuatableThing2.On,
		actuatableThing2.Off,
		Off,
		getTopic2,
	)
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, false, actuatableThing1.IsOn)
	assert.Equal(t, 6291, actuatableThing1.Index)

	err = actionRouter.client.Publish(setTopic1, mqtt_client.ExactlyOnce, false, "1")
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(time.Second)

	assert.Equal(t, true, actuatableThing1.IsOn)
	assert.Equal(t, 6291, actuatableThing1.Index)
	assert.Equal(t, "1", lastPayload1)

	assert.Equal(t, false, actuatableThing2.IsOn)
	assert.Equal(t, 6291, actuatableThing2.Index)

	err = actionRouter.client.Publish(setTopic2, mqtt_client.ExactlyOnce, false, "1")
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(time.Second)

	assert.Equal(t, true, actuatableThing2.IsOn)
	assert.Equal(t, 6291, actuatableThing2.Index)
	assert.Equal(t, "1", lastPayload2)

	err = actionRouter.RemoveAllActions()
	if err != nil {
		log.Fatal(err)
	}

	teardownActionRouterTest(actionRouter)
}
