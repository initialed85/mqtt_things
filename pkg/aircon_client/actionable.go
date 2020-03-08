package aircon_client

type Arguments struct {
}

type Actionable struct {
	Client Client
}

func (a *Actionable) On(arguments interface{}) error {
	return a.Client.TurnOn()
}

func (a *Actionable) Off(arguments interface{}) error {
	return a.Client.TurnOff()
}
