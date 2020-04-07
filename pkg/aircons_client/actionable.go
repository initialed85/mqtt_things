package aircons_client

type Arguments struct {
	Name string
}

type Actionable struct {
	Client Client
}

func (a *Actionable) On(arguments interface{}) error {
	aircon, err := a.Client.GetAircon(arguments.(Arguments).Name)
	if err != nil {
		return err
	}

	return aircon.On()
}

func (a *Actionable) Off(arguments interface{}) error {
	aircon, err := a.Client.GetAircon(arguments.(Arguments).Name)
	if err != nil {
		return err
	}

	return aircon.Off()
}
