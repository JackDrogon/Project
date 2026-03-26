package create

import (
	"fmt"
	"maps"

	"github.com/JackDrogon/project/internal/adapters/protocoltoml"
	domain "github.com/JackDrogon/project/internal/scaffold"
)

func mergeReplayInputs(replay protocoltoml.Replay, templateInputValues map[string]string) map[string]string {
	mergedInputs := make(map[string]string, len(replay.Inputs)+len(templateInputValues))
	for key, value := range replay.Inputs {
		if key == "module_path" {
			continue
		}
		mergedInputs[key] = value
	}
	maps.Copy(mergedInputs, templateInputValues)
	return mergedInputs
}

func buildReplay(command Command, opts Options) (protocoltoml.Replay, error) {
	resolvedGitMode, err := domain.ResolveGitMode(domain.CreateRequest{NoGit: opts.NoGit, GitMode: domain.GitMode(opts.GitMode)})
	if err != nil {
		return protocoltoml.Replay{}, fmt.Errorf("failed to resolve replay after project creation: %w", err)
	}

	inputs := replayInputsFromOptions(opts)

	return protocoltoml.Replay{
		Version:  protocoltoml.ReplayVersion,
		Mode:     string(command),
		Template: protocoltoml.ReplayTemplate{Lang: opts.Lang},
		Project: protocoltoml.ReplayProject{
			Name:       opts.ProjectName,
			TargetDir:  opts.DestinationDir(),
			ModulePath: opts.ModulePath,
		},
		Git:     protocoltoml.ReplayGit{Mode: domain.GitMode(resolvedGitMode), Signoff: opts.Signoff},
		Options: protocoltoml.ReplayOptions{Force: opts.Force},
		Inputs:  inputs,
	}, nil
}

func replayInputsFromOptions(opts Options) map[string]string {
	inputs := maps.Clone(opts.TemplateInputValues)
	if inputs == nil {
		inputs = map[string]string{}
	}
	if opts.ModulePath != "" {
		inputs["module_path"] = opts.ModulePath
	}
	return inputs
}
