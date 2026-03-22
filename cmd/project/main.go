package main

import (
	"os"
)

func main() {
	creator := newCommandCreator(os.Stdout)
	Execute(creator)
}
