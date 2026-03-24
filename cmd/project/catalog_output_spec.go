package main

import (
	"fmt"

	"github.com/JackDrogon/project/internal/presenters"
)

func defaultCatalogOutputSpec(asTOML, compact bool) presenters.OutputSpec {
	layout := presenters.TextLayoutDefault
	if compact {
		layout = presenters.TextLayoutCompact
	}
	return presenters.OutputSpec{Format: selectedOutputFormat(asTOML), TextLayout: layout}
}

func listTableOutputSpec(asTOML, compact bool) (presenters.OutputSpec, error) {
	if asTOML {
		return presenters.OutputSpec{}, fmt.Errorf("table output is only supported for text format")
	}
	if compact {
		return presenters.OutputSpec{}, fmt.Errorf("--table cannot be combined with --compact")
	}
	return presenters.OutputSpec{Format: outputFormatText, TextLayout: presenters.TextLayoutTable}, nil
}
