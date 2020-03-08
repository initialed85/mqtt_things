package relays_client

type Arguments struct {
	Relay int64
}

type Actionable struct {
	Client Client
}

func (a *Actionable) On(arguments interface{}) error {
	return a.Client.On(arguments.(Arguments).Relay)
}

func (a *Actionable) Off(arguments interface{}) error {
	return a.Client.Off(arguments.(Arguments).Relay)
}
