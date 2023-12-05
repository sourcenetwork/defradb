// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

/*
gendocs is a tool to generate the collections' documents automatically.
*/
package main

import (
	"os"

	"github.com/sourcenetwork/defradb/cli"
	"github.com/sourcenetwork/defradb/config"
	genCLI "github.com/sourcenetwork/defradb/tests/gen/cli"
)

func main() {
	conf := config.DefaultConfig()
	rootCmd := cli.NewDefraCommand(conf)
	gendocsCmd := genCLI.MakeGenDocCommand(conf)
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
