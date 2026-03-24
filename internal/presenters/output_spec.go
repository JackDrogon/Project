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
