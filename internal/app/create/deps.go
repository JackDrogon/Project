package create

type Dependencies struct {
	SettingsResolver   ScaffoldSettingsResolver
	NewTargetResolver  NewTargetResolver
	InitTargetResolver InitTargetResolver
}

func DefaultDependencies() Dependencies {
	return Dependencies{
		SettingsResolver:   newScaffoldSettingsResolver(),
		NewTargetResolver:  newNewTargetResolver(),
		InitTargetResolver: newInitTargetResolver(),
	}
}

func (d Dependencies) withDefaults() Dependencies {
	defaults := DefaultDependencies()
	if d.SettingsResolver == nil {
		d.SettingsResolver = defaults.SettingsResolver
	}
	if d.NewTargetResolver == nil {
		d.NewTargetResolver = defaults.NewTargetResolver
	}
	if d.InitTargetResolver == nil {
		d.InitTargetResolver = defaults.InitTargetResolver
	}
	return d
}
