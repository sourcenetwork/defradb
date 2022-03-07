package main

import (
	"log"

	"github.com/sourcenetwork/defradb/cli/defradb/cmd"

	"github.com/spf13/cobra/doc"
)

func main() {
	root := cmd.RootCmd
	err := doc.GenMarkdownTree(root, "./")
	if err != nil {
		log.Fatal(err)
	}
}
