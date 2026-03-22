package cli

import (
	"io"

	"github.com/JackDrogon/project/internal/app/catalog"
)

type Formatter interface {
	WriteLangs(w io.Writer, langs []string) error
	WriteSummaries(w io.Writer, summaries []catalog.Summary) error
	WriteInspection(w io.Writer, inspection catalog.Inspection) error
}

type textFormatter struct{}

func (f *textFormatter) WriteLangs(w io.Writer, langs []string) error {
	return writeTextLangs(w, langs)
}

func (f *textFormatter) WriteSummaries(w io.Writer, summaries []catalog.Summary) error {
	return writeTextSummaries(w, summaries)
}

func (f *textFormatter) WriteInspection(w io.Writer, inspection catalog.Inspection) error {
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
