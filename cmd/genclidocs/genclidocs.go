// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

/*
genclidocs is a tool to generate the command line interface documentation.
*/
package main

import (
	"context"
	"flag"
	"os"

	"github.com/spf13/cobra/doc"

	"github.com/sourcenetwork/defradb/cli"
	"github.com/sourcenetwork/defradb/config"
	"github.com/sourcenetwork/defradb/logging"
)

var log = logging.MustNewLogger("genclidocs")

func main() {
	path := flag.String("o", "docs/cmd", "path to write the cmd docs to")
	flag.Parse()
	err := os.MkdirAll(*path, os.ModePerm)
	if err != nil {
		log.FatalE(context.Background(), "Creating the filesystem path failed", err)
	}
	defraCmd := cli.NewDefraCommand(config.DefaultConfig())
	defraCmd.RootCmd.DisableAutoGenTag = true
	err = doc.GenMarkdownTree(defraCmd.RootCmd, *path)
	if err != nil {
		log.FatalE(context.Background(), "Generating cmd docs failed", err)
	}
}
