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

	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/config"
	"github.com/sourcenetwork/defradb/errors"
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
	Use:   "init",
	Short: "Initialize DefraDB's root directory and configuration file",
	Long:  "Initialize a directory for configuration and data at the given path.",
	// Load a default configuration, considering env. variables and CLI flags.
	PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
		if err := cfg.LoadWithRootdir(false); err != nil {
			return errors.Wrap("failed to load configuration", err)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if config.FolderExists(cfg.Rootdir) {
			if cfg.ConfigFileExists() {
				if reinitialize {
					if err := cfg.DeleteConfigFile(); err != nil {
						return err
					}
					if err := cfg.WriteConfigFile(); err != nil {
						return err
					}
				} else {
					log.FeedbackError(
						cmd.Context(),
						fmt.Sprintf(
							"Configuration file already exists at %v. Consider using --reinitialize",
							cfg.ConfigFilePath(),
						),
					)
				}
			} else {
				if err := cfg.WriteConfigFile(); err != nil {
					return errors.Wrap("failed to create configuration file", err)
				}
			}
		} else {
			if err := cfg.CreateRootDirAndConfigFile(); err != nil {
				return err
			}
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

	initCmd.Flags().StringVar(
		&cfg.Rootdir, "rootdir", config.DefaultRootDir(),
		"Directory for data and configuration to use",
	)
}
