package presenters

import (
	"io"

	"github.com/JackDrogon/project/internal/app/catalog"
)

type Formatter interface {
	WriteLangs(w io.Writer, langs []string) error
	WriteSummaries(w io.Writer, summaries []catalog.Summary) error
	WriteInspection(w io.Writer, inspection catalog.Inspection) error
}

type textFormatter struct {
	layout TextLayout
}

func (f *textFormatter) WriteLangs(w io.Writer, langs []string) error {
	return writeTextLangs(w, langs)
}

func (f *textFormatter) WriteSummaries(w io.Writer, summaries []catalog.Summary) error {
	if f.layout == TextLayoutTable {
		return writeTableTextSummaries(w, summaries)
	}
	if f.layout == TextLayoutCompact {
		return writeCompactTextSummaries(w, summaries)
	}
	return writeTextSummaries(w, summaries)
}

func (f *textFormatter) WriteInspection(w io.Writer, inspection catalog.Inspection) error {
	if f.layout == TextLayoutCompact {
		return writeCompactTextInspection(w, inspection)
	}
	return writeTextInspection(w, inspection)
}

type tomlFormatter struct{}

func (f *tomlFormatter) WriteLangs(w io.Writer, langs []string) error {
	return writeTOMLLangs(w, langs)
}

func (f *tomlFormatter) WriteSummaries(w io.Writer, summaries []catalog.Summary) error {
	return writeTOMLSummaries(w, summaries)
}

func (f *tomlFormatter) WriteInspection(w io.Writer, inspection catalog.Inspection) error {
	return writeTOMLInspection(w, inspection)
}
