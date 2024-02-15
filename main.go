package main

import (
	"embed"
	"flag"
)

var (
	lang string
)

//go:embed templates
var templates embed.FS

func initFlags() {
	flag.StringVar(&lang, "lang", "go", "programming language")
}

func main() {
	initFlags()

	langTemplateDir, err := templates.ReadDir("templates/" + lang)
	if err != nil {
		panic(err)
	}
	for _, file := range langTemplateDir {
		println(file.Name())
	}
}
