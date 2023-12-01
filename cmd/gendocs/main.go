package main

import (
	"os"

	"github.com/sourcenetwork/defradb/cli"
	"github.com/sourcenetwork/defradb/config"
)

func main() {
	conf := config.DefaultConfig()
	rootCmd := cli.NewDefraCommand(conf)
	gendocsCmd := cli.MakeGenDocCommand(conf)
	rootCmd.AddCommand(gendocsCmd)
	if err := rootCmd.Execute(); err != nil {
		// this error is okay to discard because cobra
		// logs any errors encountered during execution
		//
		// exiting with a non-zero status code signals
		// that an error has ocurred during execution
		os.Exit(1)
	}
}
