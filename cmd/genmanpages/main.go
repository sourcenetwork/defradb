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
genmanpages handles the generation of man pages. It is meant to be used as part of packaging scripts, as man pages
installation is packaging and system dependent.
*/
package main

import (
	"flag"
	"log"
	"os"

	"github.com/spf13/cobra/doc"

	"github.com/sourcenetwork/defradb/cli"
)

const defaultPerm os.FileMode = 0o777

var dir string

var header = &doc.GenManHeader{
	Title:   "defradb - Peer-to-Peer Edge Database",
	Section: "1",
}

func init() {
	flag.StringVar(&dir, "o", "build/man", "Directory in which to generate DefraDB man pages")
}

func main() {
	flag.Parse()

	defraCmd := cli.NewDefraCommand()

	if err := os.MkdirAll(dir, defaultPerm); err != nil {
		log.Fatal("Failed to create directory", err)
	}

	if err := doc.GenManTree(defraCmd, header, dir); err != nil {
		log.Fatal("Failed generation of man pages", err)
	}
}
