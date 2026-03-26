package config

import (
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/JackDrogon/project/internal/adapters/protocoltoml"
)

func TestLoadActiveConfig_UsesUserConfigDirWhenDefaultFileExists(t *testing.T) {
	t.Parallel()

	const userConfigDir = "/home/test/.config"
	const expectedPath = "/home/test/.config/project/config.toml"
	decoded := protocoltoml.Config{Version: protocoltoml.ConfigVersion}

	var gotJoinArgs []string
	svc := NewServiceWithDeps(Dependencies{
		UserConfigDir: func() (string, error) {
			return userConfigDir, nil
		},
		Join: func(elem ...string) string {
			gotJoinArgs = append([]string(nil), elem...)
			return expectedPath
		},
		ReadFile: func(path string) ([]byte, error) {
			if path != expectedPath {
				t.Fatalf("ReadFile() path = %q, want %q", path, expectedPath)
			}
			return []byte("version = 1\n"), nil
		},
		Decode: func(content []byte, path string) (protocoltoml.Config, error) {
			if string(content) != "version = 1\n" {
				t.Fatalf("Decode() content = %q", string(content))
			}
			if path != expectedPath {
				t.Fatalf("Decode() path = %q, want %q", path, expectedPath)
			}
			return decoded, nil
		},
	})

	active, err := svc.LoadActiveConfig(Context{})
	if err != nil {
		t.Fatalf("LoadActiveConfig() error = %v", err)
	}

	if !reflect.DeepEqual(gotJoinArgs, []string{userConfigDir, "project", defaultConfigFileName}) {
		t.Fatalf("Join() args = %v, want %v", gotJoinArgs, []string{userConfigDir, "project", defaultConfigFileName})
	}
	if active.Source != SourceUserConfig {
		t.Fatalf("Source = %q, want %q", active.Source, SourceUserConfig)
	}
	if active.Path != expectedPath {
		t.Fatalf("Path = %q, want %q", active.Path, expectedPath)
	}
	if active.Config == nil {
		t.Fatal("Config = nil, want decoded config")
	}
	if !reflect.DeepEqual(*active.Config, decoded) {
		t.Fatalf("Config = %#v, want %#v", *active.Config, decoded)
	}
}

func TestLoadActiveConfig_ExplicitPathShortCircuitsUserConfig(t *testing.T) {
	t.Parallel()

	const explicitPath = "/tmp/explicit.toml"
	decoded := protocoltoml.Config{Version: protocoltoml.ConfigVersion}

	svc := NewServiceWithDeps(Dependencies{
		UserConfigDir: func() (string, error) {
			t.Fatal("UserConfigDir() should not be called when explicit path is set")
			return "", nil
		},
		Join: func(elem ...string) string {
			t.Fatal("Join() should not be called when explicit path is set")
			return ""
		},
		ReadFile: func(path string) ([]byte, error) {
			if path != explicitPath {
				t.Fatalf("ReadFile() path = %q, want %q", path, explicitPath)
			}
			return []byte("version = 1\n"), nil
		},
		Decode: func(content []byte, path string) (protocoltoml.Config, error) {
			if path != explicitPath {
				t.Fatalf("Decode() path = %q, want %q", path, explicitPath)
			}
			return decoded, nil
		},
	})

	active, err := svc.LoadActiveConfig(Context{ExplicitPath: explicitPath})
	if err != nil {
		t.Fatalf("LoadActiveConfig() error = %v", err)
	}

	if active.Source != SourceExplicit {
		t.Fatalf("Source = %q, want %q", active.Source, SourceExplicit)
	}
	if active.Path != explicitPath {
		t.Fatalf("Path = %q, want %q", active.Path, explicitPath)
	}
	if active.Config == nil {
		t.Fatal("Config = nil, want decoded config")
	}
}

func TestLoadActiveConfig_MissingExplicitPathFailsAndMissingDefaultIsNoop(t *testing.T) {
	t.Parallel()

	t.Run("missing explicit path fails", func(t *testing.T) {
		t.Parallel()

		const explicitPath = "/tmp/missing.toml"
		svc := NewServiceWithDeps(Dependencies{
			ReadFile: func(path string) ([]byte, error) {
				if path != explicitPath {
					t.Fatalf("ReadFile() path = %q, want %q", path, explicitPath)
				}
				return nil, os.ErrNotExist
			},
		})

		_, err := svc.LoadActiveConfig(Context{ExplicitPath: explicitPath})
		if err == nil {
			t.Fatal("LoadActiveConfig() error = nil, want missing explicit path error")
		}
		if !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("LoadActiveConfig() error = %v, want os.ErrNotExist", err)
		}
	})

	t.Run("missing default path is noop", func(t *testing.T) {
		t.Parallel()

		const userConfigDir = "/home/test/.config"
		const expectedPath = "/home/test/.config/project/config.toml"

		svc := NewServiceWithDeps(Dependencies{
			UserConfigDir: func() (string, error) {
				return userConfigDir, nil
			},
			Join: func(elem ...string) string {
				return expectedPath
			},
			ReadFile: func(path string) ([]byte, error) {
				if path != expectedPath {
					t.Fatalf("ReadFile() path = %q, want %q", path, expectedPath)
				}
				return nil, os.ErrNotExist
			},
		})

		active, err := svc.LoadActiveConfig(Context{})
		if err != nil {
			t.Fatalf("LoadActiveConfig() error = %v", err)
		}
		if active.Source != SourceNone {
			t.Fatalf("Source = %q, want %q", active.Source, SourceNone)
		}
		if active.Path != "" {
			t.Fatalf("Path = %q, want empty", active.Path)
		}
		if active.Config != nil {
			t.Fatalf("Config = %#v, want nil", active.Config)
		}
	})
}
