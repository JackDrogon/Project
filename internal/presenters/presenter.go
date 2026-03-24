package presenters

import (
	"fmt"
	"io"

	"github.com/JackDrogon/project/internal/app/catalog"
)

type Presenter struct {
	formatter Formatter
}

func NewPresenter(spec OutputSpec) (*Presenter, error) {
	var formatter Formatter
	switch spec.Format {
	case "text":
		summaryWriter, err := newSummaryTextWriter(spec.Summary)
		if err != nil {
			return nil, err
		}
		inspectionWriter, err := newInspectionTextWriter(spec.Inspection)
		if err != nil {
			return nil, err
		}
		formatter = &textFormatter{summary: summaryWriter, inspection: inspectionWriter}
	case "toml":
		if spec.Summary.TextLayout != "" && spec.Summary.TextLayout != TextLayoutDefault {
			return nil, fmt.Errorf("%s output is only supported for text format", spec.Summary.TextLayout)
		}
		if spec.Inspection.TextLayout != "" && spec.Inspection.TextLayout != TextLayoutDefault {
			return nil, fmt.Errorf("%s output is only supported for text format", spec.Inspection.TextLayout)
		}
		formatter = &tomlFormatter{}
	default:
		return nil, fmt.Errorf("unsupported format: %s", spec.Format)
	}
	return &Presenter{formatter: formatter}, nil
}

func NewTextPresenter() *Presenter {
	presenter, _ := NewPresenter(OutputSpec{Format: "text", Summary: DefaultSummaryViewSpec(), Inspection: DefaultInspectionViewSpec()})
	return presenter
}

func NewTOMLPresenter() *Presenter {
	return &Presenter{formatter: &tomlFormatter{}}
}

func NewCompactTextPresenter() *Presenter {
	presenter, _ := NewPresenter(OutputSpec{Format: "text", Summary: SummaryViewSpec{TextLayout: TextLayoutCompact}, Inspection: InspectionViewSpec{TextLayout: TextLayoutCompact}})
	return presenter
}

func NewTableTextPresenter() *Presenter {
	presenter, _ := NewPresenter(OutputSpec{Format: "text", Summary: SummaryViewSpec{TextLayout: TextLayoutTable}, Inspection: DefaultInspectionViewSpec()})
	return presenter
}

func newSummaryTextWriter(spec SummaryViewSpec) (summaryWriter, error) {
	switch spec.TextLayout {
	case "", TextLayoutDefault:
		return defaultSummaryTextWriter{}, nil
	case TextLayoutCompact:
		return compactSummaryTextWriter{}, nil
	case TextLayoutTable:
		return tableSummaryTextWriter{}, nil
	default:
		return nil, fmt.Errorf("unsupported summary text layout: %s", spec.TextLayout)
	}
}

func newInspectionTextWriter(spec InspectionViewSpec) (inspectionWriter, error) {
	switch spec.TextLayout {
	case "", TextLayoutDefault:
		return defaultInspectionTextWriter{}, nil
	case TextLayoutCompact:
		return compactInspectionTextWriter{}, nil
	case TextLayoutTable:
		return nil, fmt.Errorf("table output is only supported for summary text views")
	default:
		return nil, fmt.Errorf("unsupported inspection text layout: %s", spec.TextLayout)
	}
}

func (p *Presenter) WriteLangs(w io.Writer, langs []string) error {
	return p.formatter.WriteLangs(w, langs)
}

func (p *Presenter) WriteSummaries(w io.Writer, summaries []catalog.Summary) error {
	return p.formatter.WriteSummaries(w, summaries)
}

func (p *Presenter) WriteInspection(w io.Writer, inspection catalog.Inspection) error {
	return p.formatter.WriteInspection(w, inspection)
}
