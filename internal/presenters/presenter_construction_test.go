package presenters

import (
	"strings"
	"testing"
)

func TestNewPresenterConstruction(t *testing.T) {
	tests := []struct {
		name    string
		format  string
		summary SummaryViewSpec
		inspect InspectionViewSpec
		wantErr bool
	}{
		{"text format", "text", DefaultSummaryViewSpec(), DefaultInspectionViewSpec(), false},
		{"compact text format", "text", SummaryViewSpec{TextLayout: TextLayoutCompact}, InspectionViewSpec{TextLayout: TextLayoutCompact}, false},
		{"table text format", "text", SummaryViewSpec{TextLayout: TextLayoutTable}, DefaultInspectionViewSpec(), false},
		{"toml format", "toml", DefaultSummaryViewSpec(), DefaultInspectionViewSpec(), false},
		{"compact toml rejected", "toml", SummaryViewSpec{TextLayout: TextLayoutCompact}, DefaultInspectionViewSpec(), true},
		{"compact toml inspection rejected", "toml", DefaultSummaryViewSpec(), InspectionViewSpec{TextLayout: TextLayoutCompact}, true},
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
