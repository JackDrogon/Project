package create

import (
	"path/filepath"
	"testing"

	"github.com/JackDrogon/project/internal/adapters/protocoltoml"
	appconfig "github.com/JackDrogon/project/internal/app/config"
)

type stubSettingsResolver struct{ settings resolvedScaffoldSettings }

func (r stubSettingsResolver) Resolve(Flags, Changed, Runtime) (resolvedScaffoldSettings, error) {
	return r.settings, nil
}

type stubNewTargetResolver struct{ target targetResolution }

func (r stubNewTargetResolver) Resolve(NewRequest, Runtime, resolvedScaffoldSettings) (targetResolution, error) {
	return r.target, nil
}

type stubInitTargetResolver struct{ target targetResolution }

func (r stubInitTargetResolver) Resolve(InitRequest, Runtime, resolvedScaffoldSettings) (targetResolution, error) {
	return r.target, nil
}

func writeReplayForCreateServiceTest(t *testing.T, replay protocoltoml.Replay) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "replay.toml")
	if err := protocoltoml.WriteReplay(path, replay); err != nil {
		t.Fatalf("WriteReplay(%q) error = %v", path, err)
	}

	return path
}

func stringPtr(v string) *string {
	return &v
}

func boolPtr(v bool) *bool {
	return &v
}

func activeConfigForCreateServiceTest(cfg protocoltoml.Config) appconfig.ActiveConfig {
	return appconfig.ActiveConfig{
		Source: appconfig.SourceExplicit,
		Path:   "config.toml",
		Config: &cfg,
	}
}
