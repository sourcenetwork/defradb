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
Package cli provides the command-line interface.
*/
package cli

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"os"
	"strings"

	"github.com/sourcenetwork/defradb/config"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/logging"
)

const badgerDatastoreName = "badger"

// Errors with how the command is invoked by user
var usageErrors = []string{
	// cobra errors - subject to change with new versions of cobra
	"flag needs an argument",
	"invalid syntax",
	"unknown flag",
	"unknown shorthand flag",
	"unknown command",
	// custom defradb errors
	errMissingArg,
	errMissingArgs,
	errTooManyArgs,
}

var log = logging.MustNewLogger("defra.cli")

var cfg = config.DefaultConfig()
var RootCmd = rootCmd

func Execute() {
	ctx := context.Background()
	// Silence cobra's default output to control usage and error display.
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true
	rootCmd.SetOut(os.Stdout)
	cmd, err := rootCmd.ExecuteContextC(ctx)
	if err != nil {
		for _, cobraError := range usageErrors {
			if strings.HasPrefix(err.Error(), cobraError) {
				log.FeedbackErrorE(ctx, "Usage error", err)
				if usageErr := cmd.Usage(); usageErr != nil {
					log.FeedbackFatalE(ctx, "error displaying usage help", usageErr)
				}
				os.Exit(1)
			}
		}
		log.FeedbackFatalE(ctx, "Execution error", err)
	}
}

func isFileInfoPipe(fi os.FileInfo) bool {
	return fi.Mode()&os.ModeNamedPipe != 0
}

func readStdin() (string, error) {
	var s strings.Builder
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		s.Write(scanner.Bytes())
	}
	if err := scanner.Err(); err != nil {
		return "", errors.Wrap("reading standard input", err)
	}
	return s.String(), nil
}

func indentJSON(b []byte) (string, error) {
	var indentedJSON bytes.Buffer
	err := json.Indent(&indentedJSON, b, "", "  ")
	return indentedJSON.String(), err
}

type graphqlErrors struct {
	Errors any `json:"errors"`
}

func hasGraphQLErrors(buf []byte) (bool, error) {
	errs := graphqlErrors{}
	err := json.Unmarshal(buf, &errs)
	if err != nil {
		return false, errors.Wrap("couldn't parse GraphQL response %w", err)
	}
	if errs.Errors != nil {
		return true, nil
	} else {
		return false, nil
	}
}
