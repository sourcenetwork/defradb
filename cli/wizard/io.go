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

// getRootDir is a helper function that gets the root directory of the defradb installation.
func getRootDir() string {
	return os.Getenv("HOME") + "/.defradb"
}

// getEnvFilename is a helper function that gets the .env filename from the config.yaml file.
func getEnvFilename(ctx *WizardContext) (string, error) {
	// First look at the config file, if it exists, for the .env filename
	envFilename, ok := getConfigValue(ctx, "secretfile").(string)
	if !ok {
		return "", errors.New(errFailedToGetEnvFilename)
	}
	// If there was a problem getting the filename, use the default
	if envFilename == "" {
		envFilename = ".env"
	}
	return envFilename, nil
}

// setConfigValue parses the config file into a *viper.Viper, modifies a target, then saves back to the file
func setConfigValue(ctx *WizardContext, target string, value any) error {
	defaultCmd := &cobra.Command{}
	cfg, err := config.LoadConfig(ctx.RootDir, defaultCmd.Flags())
	if err != nil {
		return err
	}
	cfg.Set(target, value)
	return cfg.WriteConfigAs(filepath.Join(ctx.RootDir, "config.yaml"))
}

// getConfigValue parses the config file into a *viper.Viper, then gets a value from it
func getConfigValue(ctx *WizardContext, target string) any {
	defaultCmd := &cobra.Command{}
	cfg, err := config.LoadConfig(ctx.RootDir, defaultCmd.Flags())
	if err != nil {
		return nil
	}
	return cfg.Get(target)
}

// loadEnvVariables loads environment variables from the .env file if it exists
func loadEnvVariablesFromFile(ctx *WizardContext) error {
	envFilename, err := getEnvFilename(ctx)
	if err != nil {
		return err
	}
	err = godotenv.Load(envFilename)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

// ensureEnvValue ensures a key=value pairexists in an .env file. If the .env does not
// exist, it will be created. If the key exists, its value will be replaced. Otherwise, it will be added.
func ensureEnvValue(ctx *WizardContext, key, value string) error {
	envFilename, err := getEnvFilename(ctx)
	if err != nil {
		return err
	}

	line := key + "=" + value

	// If file does not exist, create it
	if _, err := os.Stat(envFilename); os.IsNotExist(err) {
		return os.WriteFile(envFilename, []byte(line+"\n"), 0o600)
	}

	// Read the file
	file, err := os.Open(envFilename)
	if err != nil {
		return err
	}

	// Defer closing the file,
	defer func() {
		_ = file.Close()
	}()

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
	return os.WriteFile(envFilename, []byte(output), 0o600)
}
