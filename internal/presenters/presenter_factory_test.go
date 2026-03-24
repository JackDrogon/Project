package presenters

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/JackDrogon/project/internal/app/catalog"
)

type stubSummaryWriter struct{ text string }

func (w stubSummaryWriter) WriteSummaries(out io.Writer, summaries []catalog.Summary) error {
	_, err := io.WriteString(out, w.text)
	return err
}

type stubInspectionWriter struct{ text string }

func (w stubInspectionWriter) WriteInspection(out io.Writer, inspection catalog.Inspection) error {
	_, err := io.WriteString(out, w.text)
	return err
}

type stubTextFormatterRegistry struct {
	summary    summaryWriter
	inspection inspectionWriter
}

func (r stubTextFormatterRegistry) SummaryWriter(SummaryViewSpec) (summaryWriter, error) {
	return r.summary, nil
}

func (r stubTextFormatterRegistry) InspectionWriter(InspectionViewSpec) (inspectionWriter, error) {
	return r.inspection, nil
}

type stubFormatterFactory struct{ formatter Formatter }

func (f stubFormatterFactory) Build(OutputSpec) (Formatter, error) {
	if f.formatter == nil {
		return nil, fmt.Errorf("missing formatter")
	}
	return f.formatter, nil
}

func TestNewPresenterFactoryConstruction(t *testing.T) {
	tests := []struct {
		name    string
		format  string
		summary SummaryViewSpec
		inspect InspectionViewSpec
		wantErr bool
	}{
		{"text format", "text", DefaultSummaryViewSpec(), DefaultInspectionViewSpec(), false},
		{"compact text format", "text", SummaryViewSpec{TextLayout: TextLayoutCompact}, InspectionViewSpec{TextLayout: TextLayoutCompact}, false},
		{"toml format", "toml", DefaultSummaryViewSpec(), DefaultInspectionViewSpec(), false},
		{"compact toml rejected", "toml", SummaryViewSpec{TextLayout: TextLayoutCompact}, DefaultInspectionViewSpec(), true},
		{"invalid format", "json", DefaultSummaryViewSpec(), DefaultInspectionViewSpec(), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			presenter, err := NewPresenter(OutputSpec{Format: tt.format, Summary: tt.summary, Inspection: tt.inspect})
			if tt.wantErr {
				if err == nil {
					t.Fatal("NewPresenter() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("NewPresenter() error = %v", err)
			}
			if presenter == nil {
				t.Fatal("NewPresenter() = nil")
			}
		})
	}
}

func TestNewPresenter_RejectsInspectionTableLayout(t *testing.T) {
	_, err := NewPresenter(OutputSpec{
		Format:     "text",
		Summary:    DefaultSummaryViewSpec(),
		Inspection: InspectionViewSpec{TextLayout: TextLayoutTable},
	})
	if err == nil {
		t.Fatal("NewPresenter() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "table output is only supported for summary text views") {
		t.Fatalf("NewPresenter() error = %v, want inspection table error", err)
	}
}

func TestNewPresenterWithFactory_UsesInjectedFactory(t *testing.T) {
	presenter, err := NewPresenterWithFactory(OutputSpec{Format: "text"}, stubFormatterFactory{formatter: &tomlFormatter{}})
	if err != nil {
		t.Fatalf("NewPresenterWithFactory() error = %v", err)
	}
	if presenter == nil {
		t.Fatal("NewPresenterWithFactory() = nil")
	}
}

func TestDefaultFormatterFactory_UsesInjectedTextRegistry(t *testing.T) {
	factory := defaultFormatterFactory{
		textRegistry: stubTextFormatterRegistry{
			summary:    stubSummaryWriter{text: "summary-from-registry"},
			inspection: stubInspectionWriter{text: "inspection-from-registry"},
		},
	}

	formatter, err := factory.Build(OutputSpec{Format: "text", Summary: DefaultSummaryViewSpec(), Inspection: DefaultInspectionViewSpec()})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	var summaryBuf bytes.Buffer
	if err := formatter.WriteSummaries(&summaryBuf, nil); err != nil {
		t.Fatalf("WriteSummaries() error = %v", err)
	}
	if got := summaryBuf.String(); got != "summary-from-registry" {
		t.Fatalf("WriteSummaries() = %q, want summary-from-registry", got)
	}

	var inspectBuf bytes.Buffer
	if err := formatter.WriteInspection(&inspectBuf, catalog.Inspection{}); err != nil {
		t.Fatalf("WriteInspection() error = %v", err)
	}
	if got := inspectBuf.String(); got != "inspection-from-registry" {
		t.Fatalf("WriteInspection() = %q, want inspection-from-registry", got)
	}
}

func TestPresenterConstructors(t *testing.T) {
	if NewTextPresenter() == nil {
		t.Fatal("NewTextPresenter() = nil")
	}
	if NewCompactTextPresenter() == nil {
		t.Fatal("NewCompactTextPresenter() = nil")
	}
	if NewTableTextPresenter() == nil {
		t.Fatal("NewTableTextPresenter() = nil")
	}
	if NewTOMLPresenter() == nil {
		t.Fatal("NewTOMLPresenter() = nil")
	}
}
