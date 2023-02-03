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
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/config"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/logging"
)

const badgerDatastoreName = "badger"

// List of cobra errors indicating an error occurred in the way the command was invoked.
// They are subject to change with new versions of cobra.
var usageErrors = []string{
	"flag needs an argument",
	"invalid syntax",
	"unknown flag",
	"unknown shorthand flag",
	"missing argument", // custom to defradb
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
	err := rootCmd.ExecuteContext(ctx)
	if err != nil {
		for _, cobraError := range usageErrors {
			if strings.HasPrefix(err.Error(), cobraError) {
				log.FeedbackErrorE(ctx, "Usage error", err)
				if usageErr := rootCmd.Usage(); usageErr != nil {
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

// parseAndConfigLog parses and then configures the given config.Config logging subconfig.
// we use log.Fatal instead of returning an error because we can't gurantee
// atomic updates, its either everything is properly set, or we Fatal()
func parseAndConfigLog(ctx context.Context, cfg *config.LoggingConfig, cmd *cobra.Command) error {
	// handle --loglevels <default>,<name>=<value>,...
	err := parseAndConfigLogStringParam(ctx, cfg, cfg.Level, func(l *config.LoggingConfig, v string) {
		l.Level = v
	})
	if err != nil {
		return err
	}

	// handle --logger <name>,<field>=<value>,... --logger <name2>,<field>=<value>,..
	loggerKVs, err := cmd.Flags().GetStringArray("logger")
	if err != nil {
		return errors.Wrap("can't get logger flag", err)
	}

	for _, kvs := range loggerKVs {
		if err := parseAndConfigLogAllParams(ctx, cfg, kvs); err != nil {
			return err
		}
	}

	loggingConfig, err := cfg.ToLoggerConfig()
	if err != nil {
		return errors.Wrap("could not get logging config", err)
	}
	logging.SetConfig(loggingConfig)

	return nil
}

func parseAndConfigLogAllParams(ctx context.Context, cfg *config.LoggingConfig, kvs string) error {
	if kvs == "" {
		return nil
	}

	parsed := strings.Split(kvs, ",")
	if len(parsed) <= 1 {
		return errors.New(fmt.Sprintf("logger was not provided as comma-separated pairs of <name>=<value>: %s", kvs))
	}
	name := parsed[0]

	// verify KV format (<default>,<field>=<value>,...)
	// skip the first as that will be set above
	for _, kv := range parsed[1:] {
		parsedKV := strings.Split(kv, "=")
		if len(parsedKV) != 2 {
			return errors.New(fmt.Sprintf("level was not provided as <key>=<value> pair: %s", kv))
		}

		logcfg, err := cfg.GetOrCreateNamedLogger(name)
		if err != nil {
			return errors.Wrap("could not get named logger config", err)
		}

		// handle field
		switch param := strings.ToLower(parsedKV[0]); param {
		case "level": // string
			logcfg.Level = parsedKV[1]
		case "format": // string
			logcfg.Format = parsedKV[1]
		case "output": // string
			logcfg.Output = parsedKV[1]
		case "stacktrace": // bool
			boolValue, err := strconv.ParseBool(parsedKV[1])
			if err != nil {
				return errors.Wrap("couldn't parse kv bool", err)
			}
			logcfg.Stacktrace = boolValue
		case "nocolor": // bool
			boolValue, err := strconv.ParseBool(parsedKV[1])
			if err != nil {
				return errors.Wrap("couldn't parse kv bool", err)
			}
			logcfg.NoColor = boolValue
		case "caller": // bool
			boolValue, err := strconv.ParseBool(parsedKV[1])
			if err != nil {
				return errors.Wrap("couldn't parse kv bool", err)
			}
			logcfg.Caller = boolValue
		default:
			return errors.New(fmt.Sprintf("unknown parameter for logger: %s", param))
		}
	}
	return nil
}

func parseAndConfigLogStringParam(
	ctx context.Context,
	cfg *config.LoggingConfig,
	kvs string,
	paramSetterFn logParamSetterStringFn) error {
	if kvs == "" {
		return nil //nothing todo
	}

	// check if a CSV is provided
	// if its not a CSV, then just do the regular binding to the config
	parsed := strings.Split(kvs, ",")
	paramSetterFn(cfg, parsed[0])
	if len(parsed) == 1 {
		return nil //nothing more todo
	}

	// verify KV format (<default>,<name>=<value>,...)
	// skip the first as that will be set above
	for _, kv := range parsed[1:] {
		parsedKV := strings.Split(kv, "=")
		if len(parsedKV) != 2 {
			return errors.New(fmt.Sprintf("level was not provided as <key>=<value> pair: %s", kv))
		}

		logcfg, err := cfg.GetOrCreateNamedLogger(parsedKV[0])
		if err != nil {
			return errors.Wrap("could not get named logger config", err)
		}

		paramSetterFn(&logcfg.LoggingConfig, parsedKV[1])
	}
	return nil
}

type logParamSetterStringFn func(*config.LoggingConfig, string)

//
// LEAVE FOR NOW - IMPLEMENTING SOON - PLEASE IGNORE FOR NOW
//
// func parseAndConfigLogBoolParam(
//	 	ctx context.Context, cfg *config.LoggingConfig, kvs string, paramFn logParamSetterBoolFn) {
// 	if kvs == "" {
// 		return //nothing todo
// 	}

// 	// check if a CSV is provided
// 	// if its not a CSV, then just do the regular binding to the config
// 	parsed := strings.Split(kvs, ",")
// 	boolValue, err := strconv.ParseBool(parsed[0])
// 	if err != nil {
// 		log.FatalE(ctx, "couldn't parse kv bool", err)
// 	}
// 	paramFn(cfg, boolValue)
// 	if len(parsed) == 1 {
// 		return //nothing more todo
// 	}

// 	// verify KV format (<default>,<name>=<level>,...)
// 	// skip the first as that will be set above
// 	for _, kv := range parsed[1:] {
// 		parsedKV := strings.Split(kv, "=")
// 		if len(parsedKV) != 2 {
// 			log.Fatal(ctx, "field was not provided as <key>=<value> pair", logging.NewKV("pair", kv))
// 		}

// 		logcfg, err := cfg.GetOrCreateNamedLogger(parsedKV[0])
// 		if err != nil {
// 			log.FatalE(ctx, "could not get named logger config", err)
// 		}

// 		boolValue, err := strconv.ParseBool(parsedKV[1])
// 		if err != nil {
// 			log.FatalE(ctx, "couldn't parse kv bool", err)
// 		}
// 		paramFn(&logcfg.LoggingConfig, boolValue)
// 	}
// }

// type logParamSetterBoolFn func(*config.LoggingConfig, bool)
