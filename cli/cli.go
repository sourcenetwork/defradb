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
	"fmt"
	"os"
	"strings"

	"github.com/sourcenetwork/defradb/config"
	"github.com/sourcenetwork/defradb/logging"
)

const badgerDatastoreName = "badger"

var log = logging.MustNewLogger("defra.cli")

var cfg = config.DefaultConfig()
var RootCmd = rootCmd

func Execute() {
	ctx := context.Background()
	// Silence cobra's default output to control usage and error display.
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true
	err := rootCmd.ExecuteContext(ctx)
	if err != nil {
		log.FeedbackError(ctx, fmt.Sprintf("%s", err))
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
		return "", fmt.Errorf("reading standard input: %w", err)
	}
	return s.String(), nil
}

func indentJSON(b []byte) (string, error) {
	var indentedJSON bytes.Buffer
	err := json.Indent(&indentedJSON, b, "", "  ")
	return indentedJSON.String(), err
}

type graphqlErrors struct {
	Errors interface{} `json:"errors"`
}

func hasGraphQLErrors(buf []byte) (bool, error) {
	errs := graphqlErrors{}
	err := json.Unmarshal(buf, &errs)
	if err != nil {
		return false, fmt.Errorf("couldn't parse GraphQL response %w", err)
	}
	if errs.Errors != nil {
		return true, nil
	} else {
		return false, nil
	}
}
