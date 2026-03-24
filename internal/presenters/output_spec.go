package presenters

type TextLayout string

const (
	TextLayoutDefault TextLayout = "default"
	TextLayoutCompact TextLayout = "compact"
	TextLayoutTable   TextLayout = "table"
)

type OutputSpec struct {
	Format     string
	TextLayout TextLayout
}
