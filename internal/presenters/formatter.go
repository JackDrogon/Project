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

type summaryWriter interface {
	WriteSummaries(w io.Writer, summaries []catalog.Summary) error
}

type inspectionWriter interface {
	WriteInspection(w io.Writer, inspection catalog.Inspection) error
}

type textFormatter struct {
	summary    summaryWriter
	inspection inspectionWriter
}

func (f *textFormatter) WriteLangs(w io.Writer, langs []string) error {
	return writeTextLangs(w, langs)
}

func (f *textFormatter) WriteSummaries(w io.Writer, summaries []catalog.Summary) error {
	return f.summary.WriteSummaries(w, summaries)
}

func (f *textFormatter) WriteInspection(w io.Writer, inspection catalog.Inspection) error {
	return f.inspection.WriteInspection(w, inspection)
}

type tomlFormatter struct{}

type defaultSummaryTextWriter struct{}
type compactSummaryTextWriter struct{}
type tableSummaryTextWriter struct{}

type defaultInspectionTextWriter struct{}
type compactInspectionTextWriter struct{}

func (defaultSummaryTextWriter) WriteSummaries(w io.Writer, summaries []catalog.Summary) error {
	return writeTextSummaries(w, summaries)
}

func (compactSummaryTextWriter) WriteSummaries(w io.Writer, summaries []catalog.Summary) error {
	return writeCompactTextSummaries(w, summaries)
}

func (tableSummaryTextWriter) WriteSummaries(w io.Writer, summaries []catalog.Summary) error {
	return writeTableTextSummaries(w, summaries)
}

func (defaultInspectionTextWriter) WriteInspection(w io.Writer, inspection catalog.Inspection) error {
	return writeTextInspection(w, inspection)
}

func (compactInspectionTextWriter) WriteInspection(w io.Writer, inspection catalog.Inspection) error {
	return writeCompactTextInspection(w, inspection)
}

func (f *tomlFormatter) WriteLangs(w io.Writer, langs []string) error {
	return writeTOMLLangs(w, langs)
}

func (f *tomlFormatter) WriteSummaries(w io.Writer, summaries []catalog.Summary) error {
	return writeTOMLSummaries(w, summaries)
}

func (f *tomlFormatter) WriteInspection(w io.Writer, inspection catalog.Inspection) error {
	return writeTOMLInspection(w, inspection)
}
