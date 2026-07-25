package config

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

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

	active, err := svc.LoadActiveConfig(LoadOptions{})
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

	active, err := svc.LoadActiveConfig(LoadOptions{ExplicitPath: explicitPath})
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

		_, err := svc.LoadActiveConfig(LoadOptions{ExplicitPath: explicitPath})
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

		active, err := svc.LoadActiveConfig(LoadOptions{})
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

func TestResolvePath_UsesExplicitOrDefaultPath(t *testing.T) {
	t.Parallel()

	t.Run("explicit path wins", func(t *testing.T) {
		t.Parallel()

		svc := NewServiceWithDeps(Dependencies{
			UserConfigDir: func() (string, error) {
				t.Fatal("UserConfigDir() should not be called when explicit path is set")
				return "", nil
			},
		})

		path, err := svc.ResolvePath(LoadOptions{ExplicitPath: "/tmp/custom.toml"})
		if err != nil {
			t.Fatalf("ResolvePath() error = %v", err)
		}
		if path != "/tmp/custom.toml" {
			t.Fatalf("ResolvePath() = %q, want %q", path, "/tmp/custom.toml")
		}
	})

	t.Run("default path uses user config dir", func(t *testing.T) {
		t.Parallel()

		svc := NewServiceWithDeps(Dependencies{
			UserConfigDir: func() (string, error) { return "/home/test/.config", nil },
			Join: func(elem ...string) string {
				return "/home/test/.config/project/config.toml"
			},
		})

		path, err := svc.ResolvePath(LoadOptions{})
		if err != nil {
			t.Fatalf("ResolvePath() error = %v", err)
		}
		if path != "/home/test/.config/project/config.toml" {
			t.Fatalf("ResolvePath() = %q, want %q", path, "/home/test/.config/project/config.toml")
		}
	})
}

func TestInitConfig_CreatesSeedFileAndRejectsExistingTarget(t *testing.T) {
	t.Parallel()

	t.Run("creates default config file", func(t *testing.T) {
		t.Parallel()

		var mkdirPath string
		var writePath string
		var writeContent []byte

		svc := NewServiceWithDeps(Dependencies{
			UserConfigDir: func() (string, error) { return "/home/test/.config", nil },
			Join:          func(elem ...string) string { return "/home/test/.config/project/config.toml" },
			MkdirAll: func(path string, _ os.FileMode) error {
				mkdirPath = path
				return nil
			},
			Stat: func(name string) (os.FileInfo, error) {
				if name != "/home/test/.config/project/config.toml" {
					t.Fatalf("Stat() name = %q", name)
				}
				return nil, os.ErrNotExist
			},
			WriteFile: func(name string, data []byte, _ os.FileMode) error {
				writePath = name
				writeContent = append([]byte(nil), data...)
				return nil
			},
		})

		path, err := svc.InitConfig(LoadOptions{})
		if err != nil {
			t.Fatalf("InitConfig() error = %v", err)
		}
		if path != "/home/test/.config/project/config.toml" {
			t.Fatalf("InitConfig() path = %q", path)
		}
		if mkdirPath != "/home/test/.config/project" {
			t.Fatalf("MkdirAll() path = %q, want %q", mkdirPath, "/home/test/.config/project")
		}
		if writePath != path {
			t.Fatalf("WriteFile() path = %q, want %q", writePath, path)
		}
		if string(writeContent) != "version = 1\n" {
			t.Fatalf("WriteFile() content = %q, want %q", string(writeContent), "version = 1\n")
		}
	})

	t.Run("fails when config file already exists", func(t *testing.T) {
		t.Parallel()

		svc := NewServiceWithDeps(Dependencies{
			MkdirAll: func(path string, _ os.FileMode) error { return nil },
			Stat:     func(name string) (os.FileInfo, error) { return fakeFileInfo{name: filepath.Base(name)}, nil },
		})

		_, err := svc.InitConfig(LoadOptions{ExplicitPath: "/tmp/config.toml"})
		if err == nil {
			t.Fatal("InitConfig() error = nil, want existing file error")
		}
		if !strings.Contains(err.Error(), "already exists") {
			t.Fatalf("InitConfig() error = %v, want already exists", err)
		}
	})
}

type fakeFileInfo struct{ name string }

func (f fakeFileInfo) Name() string     { return f.name }
func (fakeFileInfo) Size() int64        { return 0 }
func (fakeFileInfo) Mode() os.FileMode  { return 0 }
func (fakeFileInfo) ModTime() time.Time { return time.Time{} }
func (fakeFileInfo) IsDir() bool        { return false }
func (fakeFileInfo) Sys() any           { return nil }
