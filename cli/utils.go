// Copyright 2023 Democratized Data Foundation
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
	"encoding/json"

	"github.com/sourcenetwork/defradb/config"
	"github.com/spf13/cobra"
)

// loadConfig loads the rootDir containing the configuration file,
// otherwise warn about it and load a default configuration.
func loadConfig(cfg *config.Config) error {
	if err := cfg.LoadRootDirFromFlagOrDefault(); err != nil {
		return err
	}
	return cfg.LoadWithRootdir(cfg.ConfigFileExists())
}

// createConfig creates the config directories and writes
// the current config to a file.
func createConfig(cfg *config.Config) error {
	if config.FolderExists(cfg.Rootdir) {
		return cfg.WriteConfigFile()
	}
	return cfg.CreateRootDirAndConfigFile()
}

func writeJSON(cmd *cobra.Command, out any) error {
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
