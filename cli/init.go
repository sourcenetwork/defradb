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
	"fmt"
	"os"

	"github.com/sourcenetwork/defradb/config"
	"github.com/spf13/cobra"
)

var reinitialize bool

/*
The `init` command initializes the configuration file and root directory..

It covers three possible situations:
- root dir doesn't exist
- root dir exists and doesn't contain a config file
- root dir exists and contains a config file
*/
var initCmd = &cobra.Command{
	Use:   "init [rootdir]",
	Short: "Initialize DefraDB's root directory and configuration file",
	Long:  "Initialize a directory for configuration and data at the given path.",
	// Load a default configuration, considering env. variables and CLI flags.
	PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
		err := cfg.LoadWithoutRootDir()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		// parse loglevel overrides.
		// binding the flags / EnvVars to the struct
		return parseAndConfigLog(cmd.Context(), cfg.Log, cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		rootDirPath := ""
		if len(args) == 1 {
			rootDirPath = args[0]
		} else if len(args) > 1 {
			if err := cmd.Usage(); err != nil {
				return err
			}
			return fmt.Errorf("init command requires one rootdir argument, or no argument")
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
						return fmt.Errorf("failed to remove configuration file: %w", err)
					}
					err = cfg.WriteConfigFileToRootDir(rootDir)
					if err != nil {
						return fmt.Errorf("failed to create configuration file: %w", err)
					}
					log.FeedbackInfo(cmd.Context(), fmt.Sprintf("Reinitialized configuration file at %v", configFilePath))
				} else {
					log.FeedbackError(
						cmd.Context(),
						fmt.Sprintf(
							"Configuration file already exists at %v. Consider using --reinitialize",
							configFilePath,
						),
					)
				}
			} else {
				err = cfg.WriteConfigFileToRootDir(rootDir)
				if err != nil {
					return fmt.Errorf("failed to create configuration file: %w", err)
				}
				log.FeedbackInfo(cmd.Context(), fmt.Sprintf("Initialized configuration file at %v", configFilePath))
			}
		} else {
			err = config.CreateRootDirWithDefaultConfig(rootDir)
			if err != nil {
				return fmt.Errorf("failed to create root dir: %w", err)
			}
			log.FeedbackInfo(cmd.Context(), fmt.Sprintf("Created DefraDB root directory at %v", rootDir))
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().BoolVar(
		&reinitialize, "reinitialize", false,
		"Reinitialize the configuration file",
	)
}
