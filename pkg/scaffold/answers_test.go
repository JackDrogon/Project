package scaffold

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestAnswersFile_ReadWriteRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "answers.json")
	want := AnswersFile{
		SchemaVersion: answersFileSchemaVersion,
		Command:       AnswersCommandNew,
		Lang:          "go",
		Create: AnswersFileCreate{
			ProjectName: "demo",
			TargetDir:   "demo",
			GitMode:     GitModeInitCommit,
			Signoff:     true,
			Force:       true,
		},
		TemplateInputs: map[string]string{
			"author":      "alice",
			"module_path": "example.com/demo",
		},
	}

	if err := WriteAnswersFile(path, want); err != nil {
		t.Fatalf("WriteAnswersFile() error = %v", err)
	}

	got, err := ReadAnswersFile(path)
	if err != nil {
		t.Fatalf("ReadAnswersFile() error = %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ReadAnswersFile() = %#v, want %#v", got, want)
	}
}

func TestAnswersFile_RejectsUnknownFieldsAndSchemaVersion(t *testing.T) {
	t.Run("rejects unknown top level fields", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "answers.json")
		content := []byte(`{"schema_version":1,"command":"new","lang":"go","create":{"project_name":"demo","target_dir":"demo","git_mode":"init+commit","signoff":false,"force":false},"template_inputs":{},"unknown":true}`)
		if err := os.WriteFile(path, content, 0o644); err != nil {
			t.Fatalf("os.WriteFile() error = %v", err)
		}

		_, err := ReadAnswersFile(path)
		if err == nil {
			t.Fatal("ReadAnswersFile() expected unknown field error, got nil")
		}
		if !strings.Contains(err.Error(), `unknown field "unknown"`) {
			t.Fatalf("ReadAnswersFile() error = %v, want unknown field error", err)
		}
	})

	t.Run("rejects unknown nested create fields", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "answers.json")
		content := []byte(`{"schema_version":1,"command":"new","lang":"go","create":{"project_name":"demo","target_dir":"demo","git_mode":"init+commit","signoff":false,"force":false,"module_path":"example.com/demo"},"template_inputs":{}}`)
		if err := os.WriteFile(path, content, 0o644); err != nil {
			t.Fatalf("os.WriteFile() error = %v", err)
		}

		_, err := ReadAnswersFile(path)
		if err == nil {
			t.Fatal("ReadAnswersFile() expected unknown field error, got nil")
		}
		if !strings.Contains(err.Error(), `unknown field "module_path"`) {
			t.Fatalf("ReadAnswersFile() error = %v, want unknown nested field error", err)
		}
	})

	t.Run("rejects malformed json", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "answers.json")
		content := []byte(`{"schema_version":1,"command":"new"`)
		if err := os.WriteFile(path, content, 0o644); err != nil {
			t.Fatalf("os.WriteFile() error = %v", err)
		}

		_, err := ReadAnswersFile(path)
		if err == nil {
			t.Fatal("ReadAnswersFile() expected malformed JSON error, got nil")
		}
		if !strings.Contains(err.Error(), "failed to decode answers file") {
			t.Fatalf("ReadAnswersFile() error = %v, want decode error", err)
		}
	})

	t.Run("rejects unsupported schema version", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "answers.json")
		content := []byte(`{"schema_version":2,"command":"new","lang":"go","create":{"project_name":"demo","target_dir":"demo","git_mode":"init+commit","signoff":false,"force":false},"template_inputs":{}}`)
		if err := os.WriteFile(path, content, 0o644); err != nil {
			t.Fatalf("os.WriteFile() error = %v", err)
		}

		_, err := ReadAnswersFile(path)
		if err == nil {
			t.Fatal("ReadAnswersFile() expected schema version error, got nil")
		}
		if !strings.Contains(err.Error(), "unsupported schema_version 2") {
			t.Fatalf("ReadAnswersFile() error = %v, want schema version error", err)
		}
	})

	t.Run("rejects unsupported command", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "answers.json")
		content := []byte(`{"schema_version":1,"command":"list","lang":"go","create":{"project_name":"demo","target_dir":"demo","git_mode":"init+commit","signoff":false,"force":false},"template_inputs":{}}`)
		if err := os.WriteFile(path, content, 0o644); err != nil {
			t.Fatalf("os.WriteFile() error = %v", err)
		}

		_, err := ReadAnswersFile(path)
		if err == nil {
			t.Fatal("ReadAnswersFile() expected command error, got nil")
		}
		if !strings.Contains(err.Error(), `unsupported command "list"`) {
			t.Fatalf("ReadAnswersFile() error = %v, want command error", err)
		}
	})

	t.Run("rejects empty lang", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "answers.json")
		content := []byte(`{"schema_version":1,"command":"new","lang":"","create":{"project_name":"demo","target_dir":"demo","git_mode":"init+commit","signoff":false,"force":false},"template_inputs":{}}`)
		if err := os.WriteFile(path, content, 0o644); err != nil {
			t.Fatalf("os.WriteFile() error = %v", err)
		}

		_, err := ReadAnswersFile(path)
		if err == nil {
			t.Fatal("ReadAnswersFile() expected lang validation error, got nil")
		}
		if !strings.Contains(err.Error(), "lang must not be empty") {
			t.Fatalf("ReadAnswersFile() error = %v, want lang validation error", err)
		}
	})
}

func TestAnswersFile_RejectsMissingRequiredCreateBooleans(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr string
	}{
		{
			name:    "missing signoff",
			content: `{"schema_version":1,"command":"new","lang":"go","create":{"project_name":"demo","target_dir":"demo","git_mode":"init+commit","force":false},"template_inputs":{}}`,
			wantErr: `missing required field "signoff"`,
		},
		{
			name:    "missing force",
			content: `{"schema_version":1,"command":"new","lang":"go","create":{"project_name":"demo","target_dir":"demo","git_mode":"init+commit","signoff":false},"template_inputs":{}}`,
			wantErr: `missing required field "force"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "answers.json")
			if err := os.WriteFile(path, []byte(tt.content), 0o644); err != nil {
				t.Fatalf("os.WriteFile() error = %v", err)
			}

			_, err := ReadAnswersFile(path)
			if err == nil {
				t.Fatal("ReadAnswersFile() expected missing bool field error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("ReadAnswersFile() error = %v, want contains %q", err, tt.wantErr)
			}
		})
	}
}

func TestAnswersFile_JSONOutputIsDeterministic(t *testing.T) {
	dir := t.TempDir()
	firstPath := filepath.Join(dir, "answers-first.json")
	secondPath := filepath.Join(dir, "answers-second.json")

	first := AnswersFile{
		SchemaVersion: answersFileSchemaVersion,
		Command:       AnswersCommandInit,
		Lang:          "go",
		Create: AnswersFileCreate{
			ProjectName: "demo",
			TargetDir:   ".",
			GitMode:     GitModeInitOnly,
			Signoff:     false,
			Force:       false,
		},
		TemplateInputs: map[string]string{
			"module_path": "example.com/demo",
			"author":      "alice",
		},
	}

	second := AnswersFile{
		SchemaVersion: answersFileSchemaVersion,
		Command:       AnswersCommandInit,
		Lang:          "go",
		Create: AnswersFileCreate{
			ProjectName: "demo",
			TargetDir:   ".",
			GitMode:     GitModeInitOnly,
			Signoff:     false,
			Force:       false,
		},
		TemplateInputs: map[string]string{},
	}
	second.TemplateInputs["author"] = "alice"
	second.TemplateInputs["module_path"] = "example.com/demo"

	if err := WriteAnswersFile(firstPath, first); err != nil {
		t.Fatalf("WriteAnswersFile(first) error = %v", err)
	}
	if err := WriteAnswersFile(secondPath, second); err != nil {
		t.Fatalf("WriteAnswersFile(second) error = %v", err)
	}

	firstContent, err := os.ReadFile(firstPath)
	if err != nil {
		t.Fatalf("os.ReadFile(firstPath) error = %v", err)
	}
	secondContent, err := os.ReadFile(secondPath)
	if err != nil {
		t.Fatalf("os.ReadFile(secondPath) error = %v", err)
	}

	if !bytes.Equal(firstContent, secondContent) {
		t.Fatalf("written JSON differs for same logical payload\nfirst:\n%s\nsecond:\n%s", firstContent, secondContent)
	}

	want := []byte("{\n  \"schema_version\": 1,\n  \"command\": \"init\",\n  \"lang\": \"go\",\n  \"create\": {\n    \"project_name\": \"demo\",\n    \"target_dir\": \".\",\n    \"git_mode\": \"init-only\",\n    \"signoff\": false,\n    \"force\": false\n  },\n  \"template_inputs\": {\n    \"author\": \"alice\",\n    \"module_path\": \"example.com/demo\"\n  }\n}\n")
	if !bytes.Equal(firstContent, want) {
		t.Fatalf("written JSON = %q, want %q", firstContent, want)
	}
}
