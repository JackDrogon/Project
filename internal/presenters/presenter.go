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
		formatter = &textFormatter{layout: spec.TextLayout}
	case "toml":
		if spec.TextLayout != "" && spec.TextLayout != TextLayoutDefault {
			return nil, fmt.Errorf("%s output is only supported for text format", spec.TextLayout)
		}
		formatter = &tomlFormatter{}
	default:
		return nil, fmt.Errorf("unsupported format: %s", spec.Format)
	}
	return &Presenter{formatter: formatter}, nil
}

func NewTextPresenter() *Presenter {
	return &Presenter{formatter: &textFormatter{layout: TextLayoutDefault}}
}

func NewTOMLPresenter() *Presenter {
	return &Presenter{formatter: &tomlFormatter{}}
}

func NewCompactTextPresenter() *Presenter {
	return &Presenter{formatter: &textFormatter{layout: TextLayoutCompact}}
}

func NewTableTextPresenter() *Presenter {
	return &Presenter{formatter: &textFormatter{layout: TextLayoutTable}}
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
