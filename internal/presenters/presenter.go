package presenters

import (
	"fmt"
	"io"

	"github.com/JackDrogon/project/internal/app/catalog"
)

type Presenter struct {
	formatter Formatter
}

func NewPresenter(format string, compact bool) (*Presenter, error) {
	var formatter Formatter
	switch format {
	case "text":
		formatter = &textFormatter{compact: compact}
	case "toml":
		if compact {
			return nil, fmt.Errorf("compact output is only supported for text format")
		}
		formatter = &tomlFormatter{}
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
	return &Presenter{formatter: formatter}, nil
}

func NewTextPresenter() *Presenter {
	return &Presenter{formatter: &textFormatter{}}
}

func NewTOMLPresenter() *Presenter {
	return &Presenter{formatter: &tomlFormatter{}}
}

func NewCompactTextPresenter() *Presenter {
	return &Presenter{formatter: &textFormatter{compact: true}}
}

func NewTableTextPresenter() *Presenter {
	return &Presenter{formatter: &textFormatter{table: true}}
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
