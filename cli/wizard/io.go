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
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
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

// loadEnvVariables loads environment variables from the .env file if it exists
func loadEnvVariablesFromFile() error {
	err := godotenv.Load(".env")
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

// ensureEnvValue ensures a key=value pairexists in a .env file. If the .env does not
// exist, it will be created. If the key exists, its value will be replaced. Otherwise, it will be added.
func ensureEnvValue(filename, key, value string) error {
	line := key + "=" + value

	// If file does not exist, create it
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return os.WriteFile(filename, []byte(line+"\n"), 0o600)
	}

	// Read the file
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	var lines []string
	found := false

	// Scan through the file line by
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()

		// If we find the key...
		if strings.HasPrefix(text, key+"=") {
			lines = append(lines, line)
			found = true
			continue
		}

		// ...replace it with our new key=value pair
		lines = append(lines, text)
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	// They key was not found, so add it to the file
	if !found {
		lines = append(lines, line)
	}

	// Write the file back
	output := strings.Join(lines, "\n") + "\n"
	return os.WriteFile(filename, []byte(output), 0o600)
}
