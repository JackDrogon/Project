package main

import (
	"fmt"
	"strconv"
)

const (
	outputFormatText = "text"
	outputFormatJSON = "json"
	outputFormatYAML = "yaml"
)

func selectedOutputFormat(asJSON, asYAML bool) (string, error) {
	if asJSON && asYAML {
		return "", fmt.Errorf("--json and --yaml cannot be used together")
	}

	if asJSON {
		return outputFormatJSON, nil
	}
	if asYAML {
		return outputFormatYAML, nil
	}

	return outputFormatText, nil
}

func yamlQuote(s string) string {
	return strconv.Quote(s)
}
