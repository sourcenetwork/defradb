// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package wizard

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/cli/config"
)

func getRootDir() string {
	return os.Getenv("HOME") + "/.defradb"
}

// setConfigValue parses the config file into a *viper.Viper, modifies a target, then saves back to the file
func setConfigValue(target string, value any) error {
	rootdir := getRootDir()
	defaultCmd := &cobra.Command{}
	cfg, err := config.LoadConfig(rootdir, defaultCmd.Flags())
	if err != nil {
		return err
	}
	cfg.Set(target, value)
	return cfg.WriteConfigAs(filepath.Join(rootdir, "config.yaml"))
}

// getConfigValue parses the config file into a *viper.Viper, then gets a value from it
func getConfigValue(target string) any {
	rootdir := getRootDir()
	defaultCmd := &cobra.Command{}
	cfg, err := config.LoadConfig(rootdir, defaultCmd.Flags())
	if err != nil {
		return nil
	}
	return cfg.Get(target)
}
