package main

import (
	"os"

	"github.com/JackDrogon/project/pkg/scaffold"
	"github.com/JackDrogon/project/pkg/templates"
)

func main() {
	creator := scaffold.NewCreator(templates.FS, os.Stdout)
	Execute(creator)
}
