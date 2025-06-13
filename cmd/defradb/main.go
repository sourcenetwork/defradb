// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

//go:build !js

// defradb is a decentralized peer-to-peer, user-centric, privacy-focused document database.
package main

import (
	"os"

	"github.com/sourcenetwork/defradb/cli"
)

// Execute adds all child commands to the root command and sets flags appropriately.
func main() {
	defraCmd := cli.NewDefraCommand()
	if err := defraCmd.Execute(); err != nil {
		// this error is okay to discard because cobra
		// logs any errors encountered during execution
		//
		// exiting with a non-zero status code signals
		// that an error has ocurred during execution
		os.Exit(1)
	}
}
