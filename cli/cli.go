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

	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/config"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/logging"
)

var log = logging.MustNewLogger("cli")

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

type DefraCommand struct {
	RootCmd *cobra.Command
	Cfg     *config.Config
}

// NewDefraCommand returns the root command instanciated with its tree of subcommands.
func NewDefraCommand(cfg *config.Config) DefraCommand {
	rootCmd := MakeRootCommand(cfg)
	rpcCmd := MakeRPCCommand(cfg)
	blocksCmd := MakeBlocksCommand()
	schemaCmd := MakeSchemaCommand()
	schemaMigrationCmd := MakeSchemaMigrationCommand()
	indexCmd := MakeIndexCommand()
	clientCmd := MakeClientCommand()
	rpcReplicatorCmd := MakeReplicatorCommand()
	p2pCollectionCmd := MakeP2PCollectionCommand()
	p2pCollectionCmd.AddCommand(
		MakeP2PCollectionAddCommand(cfg),
		MakeP2PCollectionRemoveCommand(cfg),
		MakeP2PCollectionGetallCommand(cfg),
	)
	rpcReplicatorCmd.AddCommand(
		MakeReplicatorGetallCommand(cfg),
		MakeReplicatorSetCommand(cfg),
		MakeReplicatorDeleteCommand(cfg),
	)
	rpcCmd.AddCommand(
		rpcReplicatorCmd,
		p2pCollectionCmd,
	)
	blocksCmd.AddCommand(
		MakeBlocksGetCommand(cfg),
	)
	schemaMigrationCmd.AddCommand(
		MakeSchemaMigrationSetCommand(cfg),
		MakeSchemaMigrationGetCommand(cfg),
	)
	schemaCmd.AddCommand(
		MakeSchemaAddCommand(cfg),
		MakeSchemaListCommand(cfg),
		MakeSchemaPatchCommand(cfg),
		schemaMigrationCmd,
	)
	indexCmd.AddCommand(
		MakeIndexCreateCommand(cfg),
		MakeIndexDropCommand(cfg),
		MakeIndexListCommand(cfg),
	)
	clientCmd.AddCommand(
		MakeDumpCommand(cfg),
		MakePingCommand(cfg),
		MakeRequestCommand(cfg),
		MakePeerIDCommand(cfg),
		MakeDBExportCommand(cfg),
		MakeDBImportCommand(cfg),
		schemaCmd,
		indexCmd,
		rpcCmd,
		blocksCmd,
	)
	rootCmd.AddCommand(
		clientCmd,
		MakeStartCommand(cfg),
		MakeServerDumpCmd(cfg),
		MakeVersionCommand(),
		MakeInitCommand(cfg),
	)

	return DefraCommand{rootCmd, cfg}
}

func (defraCmd *DefraCommand) Execute(ctx context.Context) error {
	// Silence cobra's default output to control usage and error display.
	defraCmd.RootCmd.SilenceUsage = true
	defraCmd.RootCmd.SilenceErrors = true
	defraCmd.RootCmd.SetOut(os.Stdout)
	cmd, err := defraCmd.RootCmd.ExecuteContextC(ctx)
	if err != nil {
		// Intentional cancellation.
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil
		}
		// User error.
		for _, cobraError := range usageErrors {
			if strings.HasPrefix(err.Error(), cobraError) {
				log.FeedbackErrorE(ctx, "Usage error", err)
				if usageErr := cmd.Usage(); usageErr != nil {
					log.FeedbackFatalE(ctx, "error displaying usage help", usageErr)
				}
				return err
			}
		}
		// Internal error.
		log.FeedbackErrorE(ctx, "Execution error", err)
		return err
	}
	return nil
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
