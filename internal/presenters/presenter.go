package presenters

import (
	"io"

	"github.com/JackDrogon/project/internal/app/catalog"
)

type Presenter struct {
	formatter Formatter
}

func NewPresenter(spec OutputSpec) (*Presenter, error) {
	return NewPresenterWithFactory(spec, newDefaultFormatterFactory())
}

func mustNewPresenter(spec OutputSpec) *Presenter {
	presenter, err := NewPresenter(spec)
	if err != nil {
		panic(err)
	}
	return presenter
}

func NewPresenterWithFactory(spec OutputSpec, factory FormatterFactory) (*Presenter, error) {
	formatter, err := factory.Build(spec)
	if err != nil {
		return nil, err
	}
	return &Presenter{formatter: formatter}, nil
}

func NewTextPresenter() *Presenter {
	return mustNewPresenter(OutputSpec{Format: "text", Summary: DefaultSummaryViewSpec(), Inspection: DefaultInspectionViewSpec()})
}

func NewTOMLPresenter() *Presenter {
	return &Presenter{formatter: &tomlFormatter{}}
}

func NewCompactTextPresenter() *Presenter {
	return mustNewPresenter(OutputSpec{Format: "text", Summary: SummaryViewSpec{TextLayout: TextLayoutCompact}, Inspection: InspectionViewSpec{TextLayout: TextLayoutCompact}})
}

func NewTableTextPresenter() *Presenter {
	return mustNewPresenter(OutputSpec{Format: "text", Summary: SummaryViewSpec{TextLayout: TextLayoutTable}, Inspection: DefaultInspectionViewSpec()})
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
