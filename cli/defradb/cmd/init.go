// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cmd

import (
	"context"
	"os"

	"github.com/sourcenetwork/defradb/config"
	"github.com/sourcenetwork/defradb/logging"
	"github.com/spf13/cobra"
)

var reinitialize bool

var initCmd = &cobra.Command{
	Use:   "init [rootdir]",
	Short: "Initialize DefraDB's root directory",
	Long:  `Initialize a directory for DefraDB's configuration and data at the given path.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		err := cfg.LoadWithoutRootDir()
		if err != nil {
			log.FatalE(context.Background(), "Failed to load config", err)
		}
		loggingConfig, err := cfg.GetLoggingConfig()
		if err != nil {
			log.FatalE(context.Background(), "Failed to get logging config", err)
		}
		logging.SetConfig(loggingConfig)
	},
	Args: cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		if len(args) > 0 {
			rootDir = args[0]
		}
		rootDir, exists, err := config.GetRootDir(rootDir)
		if err != nil {
			log.FatalE(ctx, "Failed to get root dir", err)
		}
		if exists {
			if reinitialize {
				log.Info(ctx, "Reinitializing root directory", logging.NewKV("path", rootDir))
				err := os.RemoveAll(rootDir)
				if err != nil {
					log.FatalE(ctx, "Failed to delete root directory", err)
				}
				err = config.CreateRootDirWithDefaultConfig(rootDir)
				if err != nil {
					log.FatalE(ctx, "Failed to create root directory", err)
				}
			} else {
				log.Warn(ctx, "Root directory already exists. Consider using --reinitialize", logging.NewKV("path", rootDir))
			}
		} else {
			err = config.CreateRootDirWithDefaultConfig(rootDir)
			if err != nil {
				log.FatalE(ctx, "Failed to create root directory", err)
			}
			log.Info(ctx, "Created DefraDB root directory", logging.NewKV("path", rootDir))
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().BoolVar(
		&reinitialize,
		"reinitialize",
		false,
		"Reinitialize the root directory",
	)
}
