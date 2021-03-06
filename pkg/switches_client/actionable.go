package switches_client

type Arguments struct {
	Name string
}

type Actionable struct {
	Client Client
}

func (a *Actionable) On(arguments interface{}) error {
	s, err := a.Client.GetSwitch(arguments.(Arguments).Name)
	if err != nil {
		return err
	}

	return s.On()
}

func (a *Actionable) Off(arguments interface{}) error {
	s, err := a.Client.GetSwitch(arguments.(Arguments).Name)
	if err != nil {
		return err
	}

	return s.Off()
}
