package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list all support languages",
	Run: func(cmd *cobra.Command, args []string) {
		listLangs()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func listLangs() {
	langs, err := templates.ReadDir("templates")
	if err != nil {
		panic(err)
	}

	for _, lang := range langs {
		fmt.Println(lang.Name())
	}
}
