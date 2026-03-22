package main

const (
	outputFormatText = "text"
	outputFormatTOML = "toml"
)

func selectedOutputFormat(asTOML bool) string {
	if asTOML {
		return outputFormatTOML
	}

	return outputFormatText
}
