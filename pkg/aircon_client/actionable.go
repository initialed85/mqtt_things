package aircon_client

type Arguments struct {
}

type Actionable struct {
	Client Client
}

func (a *Actionable) On(arguments interface{}) error {
	return a.Client.On()
}

func (a *Actionable) Off(arguments interface{}) error {
	return a.Client.Off()
}
