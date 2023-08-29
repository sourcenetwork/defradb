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
	"flag"
	"log"
	"os"

	"github.com/spf13/cobra/doc"

	"github.com/sourcenetwork/defradb/cli"
	"github.com/sourcenetwork/defradb/config"
)

var path string

func init() {
	flag.StringVar(&path, "o", "docs/cmd", "path to write the cmd docs to")
}

func main() {
	flag.Parse()

	defraCmd := cli.NewDefraCommand(config.DefaultConfig())
	defraCmd.DisableAutoGenTag = true

	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		log.Fatal("Creating the filesystem path failed", err)
	}

	if err := doc.GenMarkdownTree(defraCmd, path); err != nil {
		log.Fatal("Generating cmd docs failed", err)
	}
}
