package scaffold

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

const answersFileSchemaVersion = 1

type AnswersCommand string

const (
	AnswersCommandNew  AnswersCommand = "new"
	AnswersCommandInit AnswersCommand = "init"
)

type AnswersFile struct {
	SchemaVersion  int               `json:"schema_version"`
	Command        AnswersCommand    `json:"command"`
	Lang           string            `json:"lang"`
	Create         AnswersFileCreate `json:"create"`
	TemplateInputs map[string]string `json:"template_inputs"`
}

type AnswersFileCreate struct {
	ProjectName string  `json:"project_name"`
	TargetDir   string  `json:"target_dir"`
	GitMode     GitMode `json:"git_mode"`
	Signoff     bool    `json:"signoff"`
	Force       bool    `json:"force"`
}

func (create *AnswersFileCreate) UnmarshalJSON(data []byte) error {
	type answersFileCreateJSON struct {
		ProjectName string  `json:"project_name"`
		TargetDir   string  `json:"target_dir"`
		GitMode     GitMode `json:"git_mode"`
		Signoff     *bool   `json:"signoff"`
		Force       *bool   `json:"force"`
	}

	var decoded answersFileCreateJSON
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&decoded); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return fmt.Errorf("expected a single JSON object")
	}

	if decoded.Signoff == nil {
		return fmt.Errorf("missing required field %q", "signoff")
	}
	if decoded.Force == nil {
		return fmt.Errorf("missing required field %q", "force")
	}

	create.ProjectName = decoded.ProjectName
	create.TargetDir = decoded.TargetDir
	create.GitMode = decoded.GitMode
	create.Signoff = *decoded.Signoff
	create.Force = *decoded.Force

	return nil
}

func ReadAnswersFile(path string) (AnswersFile, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return AnswersFile{}, fmt.Errorf("failed to read answers file %s: %w", path, err)
	}

	answers, err := decodeAnswersFile(content, path)
	if err != nil {
		return AnswersFile{}, err
	}

	return answers, nil
}

func WriteAnswersFile(path string, answers AnswersFile) error {
	content, err := marshalAnswersFile(path, answers)
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, content, 0o644); err != nil {
		return fmt.Errorf("failed to write answers file %s: %w", path, err)
	}

	return nil
}

func decodeAnswersFile(content []byte, path string) (AnswersFile, error) {
	var answers AnswersFile
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&answers); err != nil {
		return AnswersFile{}, fmt.Errorf("failed to decode answers file %s: %w", path, err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return AnswersFile{}, fmt.Errorf("failed to decode answers file %s: expected a single JSON object", path)
	}

	if err := validateAnswersFile(answers, path); err != nil {
		return AnswersFile{}, err
	}

	return answers, nil
}

func marshalAnswersFile(path string, answers AnswersFile) ([]byte, error) {
	normalized := normalizeAnswersFileForWrite(answers)
	if err := validateAnswersFile(normalized, path); err != nil {
		return nil, err
	}

	content, err := json.MarshalIndent(normalized, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to encode answers file %s: %w", path, err)
	}

	return append(content, '\n'), nil
}

func normalizeAnswersFileForWrite(answers AnswersFile) AnswersFile {
	if answers.SchemaVersion == 0 {
		answers.SchemaVersion = answersFileSchemaVersion
	}
	if answers.TemplateInputs == nil {
		answers.TemplateInputs = map[string]string{}
	}

	return answers
}

func validateAnswersFile(answers AnswersFile, path string) error {
	if answers.SchemaVersion != answersFileSchemaVersion {
		return fmt.Errorf("answers file %s has unsupported schema_version %d", path, answers.SchemaVersion)
	}

	switch answers.Command {
	case AnswersCommandNew, AnswersCommandInit:
	default:
		return fmt.Errorf("answers file %s has unsupported command %q", path, answers.Command)
	}

	if answers.Lang == "" {
		return fmt.Errorf("answers file %s lang must not be empty", path)
	}
	if answers.Create.ProjectName == "" {
		return fmt.Errorf("answers file %s create.project_name must not be empty", path)
	}
	if answers.Create.TargetDir == "" {
		return fmt.Errorf("answers file %s create.target_dir must not be empty", path)
	}

	switch answers.Create.GitMode {
	case GitModeNone, GitModeInitOnly, GitModeInitCommit:
	default:
		return fmt.Errorf(
			"answers file %s create.git_mode %q is invalid: must be one of %s, %s, %s",
			path,
			answers.Create.GitMode,
			GitModeNone,
			GitModeInitOnly,
			GitModeInitCommit,
		)
	}

	if answers.TemplateInputs == nil {
		return fmt.Errorf("answers file %s template_inputs must be an object", path)
	}

	return nil
}
