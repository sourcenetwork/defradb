// Copyright 2024 Democratized Data Foundation
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
	"crypto/rand"
	"encoding/hex"
	"os"

	"github.com/spf13/cobra"
)

func MakeInitCommand() *cobra.Command {
	var encrypted bool
	var cmd = &cobra.Command{
		Use:   "init",
		Short: "Create root directory and configuration file",
		Long: `Create root directory and configuration file

Example: with encryption at rest
  defradb init -e
		`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return setContextRootDir(cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			rootdir := mustGetContextRootDir(cmd)
			if err := os.MkdirAll(rootdir, 0755); err != nil {
				return err
			}

			cfg := defaultConfig()
			cfg.AddConfigPath(rootdir)

			err := bindConfigFlags(cfg, cmd.Root().PersistentFlags())
			if err != nil {
				return err
			}

			if encrypted {
				// generate a random encryption key
				key := make([]byte, 32)
				if _, err := rand.Read(key); err != nil {
					return err
				}
				cfg.Set("datastore.badger.encryptionkey", hex.EncodeToString(key))
				// set index cache size to improve read performance.
				cfg.Set("datastore.badger.indexcachesize", 10<<20)
			}

			return cfg.SafeWriteConfig()
		},
	}
	cmd.Flags().BoolVarP(&encrypted, "encrypted", "e", false, "Enable data encryption at rest")
	return cmd
}
