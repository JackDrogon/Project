package config

import "github.com/JackDrogon/project/internal/adapters/protocoltoml"

type Source string

const (
	SourceNone       Source = "none"
	SourceUserConfig Source = "user-config"
	SourceExplicit   Source = "explicit-config"
)

type ActiveConfig struct {
	Source Source
	Path   string
	Config *protocoltoml.Config
}
