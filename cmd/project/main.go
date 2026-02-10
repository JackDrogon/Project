package main

import (
	"embed"
)

//go:embed templates
var templates embed.FS

func main() {
	Execute()
}
