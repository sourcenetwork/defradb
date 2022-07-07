// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/sourcenetwork/defradb/config"
	"github.com/sourcenetwork/defradb/logging"
	"github.com/spf13/cobra"
)

var reinitialize bool

/*
initCmd provides the `init` command.

It covers three possible situations:
- root dir doesn't exist
- root dir exists and doesn't contain a config file
- root dir exists and contains a config file
*/
var initCmd = &cobra.Command{
	Use:   "init [rootdir]",
	Short: "Initialize DefraDB's root directory and config file",
	Long: `
Initialize a directory for DefraDB's configuration and data at the given path.
The --reinitialize flag replaces a config file with a default one.`,
	// Load a default configuration, considering env. variables and CLI flags.
	PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
		err := cfg.LoadWithoutRootDir()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}
		loggingConfig, err := cfg.GetLoggingConfig()
		if err != nil {
			return fmt.Errorf("failed to load logging configuration: %w", err)
		}
		logging.SetConfig(loggingConfig)
		return nil
	},
	Args: cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		rootDirPath := ""
		if len(args) == 1 {
			rootDirPath = args[0]
		}
		rootDir, rootDirExists, err := config.GetRootDir(rootDirPath)
		if err != nil {
			return fmt.Errorf("failed to get root dir: %w", err)
		}
		if rootDirExists {
			// we assume the config file is using its default path in the rootdir
			configFilePath := fmt.Sprintf("%v/%v", rootDir, config.DefaultDefraDBConfigFileName)
			info, err := os.Stat(configFilePath)
			configFileExists := (err == nil && !info.IsDir())
			if configFileExists {
				if reinitialize {
					err = os.Remove(configFilePath)
					if err != nil {
						return fmt.Errorf("failed to remove config file: %w", err)
					}
					err = cfg.WriteConfigFileToRootDir(rootDir)
					if err != nil {
						return fmt.Errorf("failed to create config file: %w", err)
					}
					log.FeedbackInfo(ctx, fmt.Sprintf("Reinitialized config file at %v", configFilePath))
				} else {
					log.FeedbackInfo(
						ctx,
						fmt.Sprintf("Config file already exists at %v. Consider using --reinitialize", configFilePath),
					)
				}
			} else {
				err = cfg.WriteConfigFileToRootDir(rootDir)
				if err != nil {
					return fmt.Errorf("failed to create config file: %w", err)
				}
				log.FeedbackInfo(ctx, fmt.Sprintf("Initialized config file at %v", configFilePath))
			}
		} else {
			err = config.CreateRootDirWithDefaultConfig(rootDir)
			if err != nil {
				return fmt.Errorf("failed to create root dir: %w", err)
			}
			log.FeedbackInfo(ctx, fmt.Sprintf("Created DefraDB root directory at %v", rootDir))
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().BoolVar(
		&reinitialize, "reinitialize", false,
		"reinitialize the config file",
	)
}
