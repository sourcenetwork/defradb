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

	"github.com/sourcenetwork/defradb/logging"
	"github.com/spf13/cobra"
)

const badgerDatastoreName = "badger"

var log = logging.MustNewLogger("defra.cli")

var RootCmd *cobra.Command

func init() {
	RootCmd = MakeRootCommand()
}

func Execute() {
	ctx := context.Background()
	assembleCommandTree(RootCmd)
	// Silence cobra's default output to control usage and error display.
	RootCmd.SilenceErrors = true
	RootCmd.SilenceUsage = true
	err := RootCmd.ExecuteContext(ctx)
	if err != nil {
		log.FeedbackError(ctx, fmt.Sprintf("%s", err))
	}
}

// assembleCommandTree assembles the command tree for the CLI.
// It leverages MakeCommand funcs that exit early in case of error, for simplicity.
func assembleCommandTree(cmd *cobra.Command) *cobra.Command {
	clientCmd := MakeClientCommand()
	rpcCmd := MakeRPCCommand()
	blocksCmd := MakeBlocksCommand()
	schemaCmd := MakeSchemaCommand()
	blocksCmd.AddCommand(
		MakeBlocksGetCommand(),
	)
	schemaCmd.AddCommand(
		MakeSchemaCommand(),
	)
	rpcCmd.AddCommand(
		MakeAddReplicatorCommand(),
	)
	clientCmd.AddCommand(
		MakeDumpCommand(),
		MakePingCommand(),
		MakeSchemaCommand(),
		MakeBlocksCommand(),
		MakeQueryCommand(),
		schemaCmd,
		rpcCmd,
	)
	cmd.AddCommand(
		clientCmd,
		MakeStartCommand(),
		MakeServerDumpCmd(),
		MakeVersionCommand(),
		MakeInitCommand(),
	)
	return cmd
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
