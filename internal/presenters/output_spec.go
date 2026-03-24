package presenters

type TextLayout string

const (
	TextLayoutDefault TextLayout = "default"
	TextLayoutCompact TextLayout = "compact"
	TextLayoutTable   TextLayout = "table"
)

type OutputSpec struct {
	Format     string
	Summary    SummaryViewSpec
	Inspection InspectionViewSpec
}

type SummaryViewSpec struct {
	TextLayout TextLayout
}

type InspectionViewSpec struct {
	TextLayout TextLayout
}

func DefaultSummaryViewSpec() SummaryViewSpec {
	return SummaryViewSpec{TextLayout: TextLayoutDefault}
}

func DefaultInspectionViewSpec() InspectionViewSpec {
	return InspectionViewSpec{TextLayout: TextLayoutDefault}
}

func (s OutputSpec) withDefaults() OutputSpec {
	if s.Format == "" {
		s.Format = "text"
	}
	if s.Summary.TextLayout == "" {
		s.Summary = DefaultSummaryViewSpec()
	}
	if s.Inspection.TextLayout == "" {
		s.Inspection = DefaultInspectionViewSpec()
	}
	return s
}
