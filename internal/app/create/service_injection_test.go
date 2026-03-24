package create

import "testing"

func TestServiceWithDeps_UsesInjectedResolvers(t *testing.T) {
	deps := Dependencies{
		SettingsResolver: stubSettingsResolver{settings: resolvedScaffoldSettings{
			Lang: "cpp", ModulePath: "example.com/custom", GitMode: "none",
			TemplateInputValues: map[string]string{"author": "custom"},
		}},
		NewTargetResolver:  stubNewTargetResolver{target: targetResolution{ProjectName: "custom-new", TargetDir: "custom-new", ModulePath: "example.com/custom-new", Force: true}},
		InitTargetResolver: stubInitTargetResolver{target: targetResolution{ProjectName: "custom-init", TargetDir: "workspace", ModulePath: "example.com/custom-init", AllowExistingEmptyDir: true}},
	}
	svc := NewServiceWithDeps(deps)

	newSpec, err := svc.BuildNewSpec(NewRequest{Flags: Flags{}, Changed: Changed{}, Arg: "ignored", HasArg: true})
	if err != nil {
		t.Fatalf("BuildNewSpec() error = %v", err)
	}
	if newSpec.Options.ProjectName != "custom-new" || newSpec.Options.Lang != "cpp" || !newSpec.Options.Force {
		t.Fatalf("new spec = %#v, want injected resolver values", newSpec)
	}

	initSpec, err := svc.BuildInitSpec(InitRequest{Flags: Flags{}, Changed: Changed{}, Arg: "ignored", HasArg: true})
	if err != nil {
		t.Fatalf("BuildInitSpec() error = %v", err)
	}
	if initSpec.Options.ProjectName != "custom-init" || initSpec.Options.TargetDir != "workspace" || !initSpec.Options.AllowExistingEmptyDir {
		t.Fatalf("init spec = %#v, want injected resolver values", initSpec)
	}
	if got := initSpec.Options.TemplateInputValues["author"]; got != "custom" {
		t.Fatalf("init spec template inputs = %#v, want author custom", initSpec.Options.TemplateInputValues)
	}
}
