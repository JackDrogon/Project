package main

import (
	"embed"
	"flag"
)

var (
	langFlag      = flag.String("lang", "go", "programming language")
	listLangsFlag = flag.Bool("list", false, "list all support languages")
)

//go:embed templates
var templates embed.FS

func listLangs() {
	langs, err := templates.ReadDir("templates")
	if err != nil {
		panic(err)
	}

	for _, lang := range langs {
		println(lang.Name())
	}
}

func createProject() {
	langTemplateDir, err := templates.ReadDir("templates")
	if err != nil {
		panic(err)
	}
	for _, file := range langTemplateDir {
		println(file.Name())
	}
}

func run() {
	flag.Parse()

	if *listLangsFlag {
		listLangs()
		return
	}
}

func main() {
	run()
}
