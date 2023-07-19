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
	"context"
	"flag"
	"os"

	"github.com/spf13/cobra/doc"

	"github.com/sourcenetwork/defradb/cli"
	"github.com/sourcenetwork/defradb/config"
	"github.com/sourcenetwork/defradb/logging"
)

const defaultPerm os.FileMode = 0o777

var log = logging.MustNewLogger("genmanpages")

func main() {
	dirFlag := flag.String("o", "build/man", "Directory in which to generate DefraDB man pages")
	flag.Parse()
	genRootManPages(*dirFlag)
}

func genRootManPages(dir string) {
	ctx := context.Background()
	header := &doc.GenManHeader{
		Title:   "defradb - Peer-to-Peer Edge Database",
		Section: "1",
	}
	err := os.MkdirAll(dir, defaultPerm)
	if err != nil {
		log.FatalE(ctx, "Failed to create directory", err, logging.NewKV("dir", dir))
	}
	defraCmd := cli.NewDefraCommand(config.DefaultConfig())
	err = doc.GenManTree(defraCmd.RootCmd, header, dir)
	if err != nil {
		log.FatalE(ctx, "Failed generation of man pages", err)
	}
}
