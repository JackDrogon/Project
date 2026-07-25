package presenters

import (
	"errors"
	"fmt"
	"io"

	"github.com/JackDrogon/project/internal/app/catalog"
)

type Presenter struct {
	langs      func(io.Writer, []string) error
	summaries  func(io.Writer, []catalog.Summary) error
	inspection func(io.Writer, catalog.Inspection) error
}

func NewPresenter(spec OutputSpec) (*Presenter, error) {
	spec = spec.withDefaults()
	switch spec.Format {
	case "text":
		summaries, err := textSummariesWriter(spec.Summary)
		if err != nil {
			return nil, err
		}
		inspection, err := textInspectionWriter(spec.Inspection)
		if err != nil {
			return nil, err
		}
		return &Presenter{langs: writeTextLangs, summaries: summaries, inspection: inspection}, nil
	case "toml":
		if spec.Summary.TextLayout != TextLayoutDefault {
			return nil, fmt.Errorf("%s output is only supported for text format", spec.Summary.TextLayout)
		}
		if spec.Inspection.TextLayout != TextLayoutDefault {
			return nil, fmt.Errorf("%s output is only supported for text format", spec.Inspection.TextLayout)
		}
		return &Presenter{langs: writeTOMLLangs, summaries: writeTOMLSummaries, inspection: writeTOMLInspection}, nil
	default:
		return nil, fmt.Errorf("unsupported format: %s", spec.Format)
	}
}

func textSummariesWriter(spec SummaryViewSpec) (func(io.Writer, []catalog.Summary) error, error) {
	switch spec.TextLayout {
	case TextLayoutDefault:
		return writeTextSummaries, nil
	case TextLayoutCompact:
		return writeCompactTextSummaries, nil
	case TextLayoutTable:
		return writeTableTextSummaries, nil
	default:
		return nil, fmt.Errorf("unsupported summary text layout: %s", spec.TextLayout)
	}
}

func textInspectionWriter(spec InspectionViewSpec) (func(io.Writer, catalog.Inspection) error, error) {
	switch spec.TextLayout {
	case TextLayoutDefault:
		return writeTextInspection, nil
	case TextLayoutCompact:
		return writeCompactTextInspection, nil
	case TextLayoutTable:
		return nil, errors.New("table output is only supported for summary text views")
	default:
		return nil, fmt.Errorf("unsupported inspection text layout: %s", spec.TextLayout)
	}
}

func (p *Presenter) WriteLangs(w io.Writer, langs []string) error {
	return p.langs(w, langs)
}

func (p *Presenter) WriteSummaries(w io.Writer, summaries []catalog.Summary) error {
	return p.summaries(w, summaries)
}

func (p *Presenter) WriteInspection(w io.Writer, inspection catalog.Inspection) error {
	return p.inspection(w, inspection)
}
