package scaffold

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

const replayFileSchemaVersion = 1

type ReplayCommand string

const (
	ReplayCommandNew  ReplayCommand = "new"
	ReplayCommandInit ReplayCommand = "init"
)

type ReplayFile struct {
	SchemaVersion  int               `json:"schema_version"`
	Command        ReplayCommand     `json:"command"`
	Lang           string            `json:"lang"`
	Create         ReplayFileCreate  `json:"create"`
	TemplateInputs map[string]string `json:"template_inputs"`
}

type ReplayFileCreate struct {
	ProjectName string  `json:"project_name"`
	TargetDir   string  `json:"target_dir"`
	GitMode     GitMode `json:"git_mode"`
	Signoff     bool    `json:"signoff"`
	Force       bool    `json:"force"`
}

func (create *ReplayFileCreate) UnmarshalJSON(data []byte) error {
	type replayFileCreateJSON struct {
		ProjectName string  `json:"project_name"`
		TargetDir   string  `json:"target_dir"`
		GitMode     GitMode `json:"git_mode"`
		Signoff     *bool   `json:"signoff"`
		Force       *bool   `json:"force"`
	}

	var decoded replayFileCreateJSON
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

func ReadReplayFile(path string) (ReplayFile, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return ReplayFile{}, fmt.Errorf("failed to read replay file %s: %w", path, err)
	}

	replay, err := decodeReplayFile(content, path)
	if err != nil {
		return ReplayFile{}, err
	}

	return replay, nil
}

func WriteReplayFile(path string, replay ReplayFile) error {
	content, err := marshalReplayFile(path, replay)
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, content, 0o644); err != nil {
		return fmt.Errorf("failed to write replay file %s: %w", path, err)
	}

	return nil
}

func decodeReplayFile(content []byte, path string) (ReplayFile, error) {
	var replay ReplayFile
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&replay); err != nil {
		return ReplayFile{}, fmt.Errorf("failed to decode replay file %s: %w", path, err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return ReplayFile{}, fmt.Errorf("failed to decode replay file %s: expected a single JSON object", path)
	}

	if err := validateReplayFile(replay, path); err != nil {
		return ReplayFile{}, err
	}

	return replay, nil
}

func marshalReplayFile(path string, replay ReplayFile) ([]byte, error) {
	normalized := normalizeReplayFileForWrite(replay)
	if err := validateReplayFile(normalized, path); err != nil {
		return nil, err
	}

	content, err := json.MarshalIndent(normalized, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to encode replay file %s: %w", path, err)
	}

	return append(content, '\n'), nil
}

func normalizeReplayFileForWrite(replay ReplayFile) ReplayFile {
	if replay.SchemaVersion == 0 {
		replay.SchemaVersion = replayFileSchemaVersion
	}
	if replay.TemplateInputs == nil {
		replay.TemplateInputs = map[string]string{}
	}

	return replay
}

func validateReplayFile(replay ReplayFile, path string) error {
	if replay.SchemaVersion != replayFileSchemaVersion {
		return fmt.Errorf("replay file %s has unsupported schema_version %d", path, replay.SchemaVersion)
	}

	switch replay.Command {
	case ReplayCommandNew, ReplayCommandInit:
	default:
		return fmt.Errorf("replay file %s has unsupported command %q", path, replay.Command)
	}

	if replay.Lang == "" {
		return fmt.Errorf("replay file %s lang must not be empty", path)
	}
	if replay.Create.ProjectName == "" {
		return fmt.Errorf("replay file %s create.project_name must not be empty", path)
	}
	if replay.Create.TargetDir == "" {
		return fmt.Errorf("replay file %s create.target_dir must not be empty", path)
	}

	switch replay.Create.GitMode {
	case GitModeNone, GitModeInitOnly, GitModeInitCommit:
	default:
		return fmt.Errorf(
			"replay file %s create.git_mode %q is invalid: must be one of %s, %s, %s",
			path,
			replay.Create.GitMode,
			GitModeNone,
			GitModeInitOnly,
			GitModeInitCommit,
		)
	}

	if replay.TemplateInputs == nil {
		return fmt.Errorf("replay file %s template_inputs must be an object", path)
	}

	return nil
}
