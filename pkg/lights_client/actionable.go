package lights_client

type Arguments struct {
	Name string
}

type Actionable struct {
	Client Client
}

func (a *Actionable) On(arguments interface{}) error {
	light, err := a.Client.GetLight(arguments.(Arguments).Name)
	if err != nil {
		return err
	}

	return light.On()
}

func (a *Actionable) Off(arguments interface{}) error {
	light, err := a.Client.GetLight(arguments.(Arguments).Name)
	if err != nil {
		return err
	}

	return light.Off()
}
