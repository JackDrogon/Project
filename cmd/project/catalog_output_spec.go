package main

import (
	"fmt"

	"github.com/JackDrogon/project/internal/presenters"
)

const (
	outputFormatText = "text"
	outputFormatTOML = "toml"
)

func defaultCatalogOutputSpec(asTOML, compact bool) presenters.OutputSpec {
	layout := presenters.TextLayoutDefault
	if compact {
		layout = presenters.TextLayoutCompact
	}
	return presenters.OutputSpec{
		Format:     selectedOutputFormat(asTOML),
		Summary:    presenters.SummaryViewSpec{TextLayout: layout},
		Inspection: presenters.InspectionViewSpec{TextLayout: layout},
	}
}

func listTableOutputSpec(asTOML, compact bool) (presenters.OutputSpec, error) {
	if asTOML {
		return presenters.OutputSpec{}, fmt.Errorf("table output is only supported for text format")
	}
	if compact {
		return presenters.OutputSpec{}, fmt.Errorf("--table cannot be combined with --compact")
	}
	return presenters.OutputSpec{
		Format:     outputFormatText,
		Summary:    presenters.SummaryViewSpec{TextLayout: presenters.TextLayoutTable},
		Inspection: presenters.DefaultInspectionViewSpec(),
	}, nil
}

func inspectOutputSpec(asTOML, compact bool) presenters.OutputSpec {
	layout := presenters.TextLayoutDefault
	if compact {
		layout = presenters.TextLayoutCompact
	}
	return presenters.OutputSpec{
		Format:     selectedOutputFormat(asTOML),
		Summary:    presenters.DefaultSummaryViewSpec(),
		Inspection: presenters.InspectionViewSpec{TextLayout: layout},
	}
}

func selectedOutputFormat(asTOML bool) string {
	if asTOML {
		return outputFormatTOML
	}

	return outputFormatText
}
