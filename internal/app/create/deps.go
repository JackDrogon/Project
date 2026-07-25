package create

// dependencies carries the collaborators a Service resolves through. It stays
// unexported because every injection site is inside this package: production
// builds go through NewService, and the tests build a value and override the
// fields they care about.
type dependencies struct {
	SettingsResolver   scaffoldSettingsResolver
	NewTargetResolver  newTargetResolver
	InitTargetResolver initTargetResolver
}

func defaultDependencies() dependencies {
	return dependencies{
		SettingsResolver:   newScaffoldSettingsResolver(),
		NewTargetResolver:  newNewTargetResolver(),
		InitTargetResolver: newInitTargetResolver(),
	}
}

func (d dependencies) withDefaults() dependencies {
	defaults := defaultDependencies()
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
