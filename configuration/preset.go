package configuration

type Preset interface {
	Substitutions(string) string
}

type DefaultPreset struct{}

func (d DefaultPreset) Substitutions(userdata string) string { return userdata }
