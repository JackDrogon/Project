package presenters

import "fmt"

type Factory interface {
	Build(OutputSpec) (Formatter, error)
}

type TextWriterRegistry interface {
	SummaryWriter(SummaryViewSpec) (summaryWriter, error)
	InspectionWriter(InspectionViewSpec) (inspectionWriter, error)
}

type textWriterRegistry struct{}

type formatterFactory struct {
	textRegistry TextWriterRegistry
}

func newDefaultFactory() Factory {
	return formatterFactory{textRegistry: textWriterRegistry{}}
}

func (f formatterFactory) Build(spec OutputSpec) (Formatter, error) {
	formatted := spec.withDefaults()
	switch formatted.Format {
	case "text":
		summary, err := f.textRegistry.SummaryWriter(formatted.Summary)
		if err != nil {
			return nil, err
		}
		inspection, err := f.textRegistry.InspectionWriter(formatted.Inspection)
		if err != nil {
			return nil, err
		}
		return &textFormatter{summary: summary, inspection: inspection}, nil
	case "toml":
		if formatted.Summary.TextLayout != TextLayoutDefault {
			return nil, fmt.Errorf("%s output is only supported for text format", formatted.Summary.TextLayout)
		}
		if formatted.Inspection.TextLayout != TextLayoutDefault {
			return nil, fmt.Errorf("%s output is only supported for text format", formatted.Inspection.TextLayout)
		}
		return &tomlFormatter{}, nil
	default:
		return nil, fmt.Errorf("unsupported format: %s", formatted.Format)
	}
}

func (textWriterRegistry) SummaryWriter(spec SummaryViewSpec) (summaryWriter, error) {
	switch spec.TextLayout {
	case TextLayoutDefault:
		return defaultSummaryTextWriter{}, nil
	case TextLayoutCompact:
		return compactSummaryTextWriter{}, nil
	case TextLayoutTable:
		return tableSummaryTextWriter{}, nil
	default:
		return nil, fmt.Errorf("unsupported summary text layout: %s", spec.TextLayout)
	}
}

func (textWriterRegistry) InspectionWriter(spec InspectionViewSpec) (inspectionWriter, error) {
	switch spec.TextLayout {
	case TextLayoutDefault:
		return defaultInspectionTextWriter{}, nil
	case TextLayoutCompact:
		return compactInspectionTextWriter{}, nil
	case TextLayoutTable:
		return nil, fmt.Errorf("table output is only supported for summary text views")
	default:
		return nil, fmt.Errorf("unsupported inspection text layout: %s", spec.TextLayout)
	}
}
