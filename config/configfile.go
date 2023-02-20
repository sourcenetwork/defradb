// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package config

import (
	"fmt"
	"os"
)

const (
	DefaultDefraDBConfigFileName = "config.yaml"
	configType                   = "yaml"
	defaultDirPerm               = 0o700
	defaultConfigFilePerm        = 0o644
)

// defaultConfigTemplate must reflect the Config struct in content and configuration.
// All parameters must be represented here, to support Viper's automatic environment variable handling.
// We embed the default config template for clarity and to avoid autoformatters from breaking it.
//
//go:embed configfile_yaml.gotmpl
var defaultConfigTemplate string

	buffer, err := cfg.toBytes()
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, buffer, defaultConfigFilePerm); err != nil {
		return NewErrFailedToWriteFile(err, path)
	}
	return nil
}

// WriteConfigFile writes a config file in a given root directory.
func (cfg *Config) WriteConfigFileToRootDir(rootDir string) error {
	path := fmt.Sprintf("%v/%v", rootDir, DefaultDefraDBConfigFileName)
	return cfg.writeConfigFile(path)
}

func (cfg *Config) CreateRootDirAndConfigFile() error {
	if err := os.MkdirAll(cfg.Rootdir, defaultDirPerm); err != nil {
		return err
	}
	log.FeedbackInfo(context.Background(), fmt.Sprintf("Created DefraDB root directory at %v", cfg.Rootdir))
	if err := cfg.WriteConfigFile(); err != nil {
		return err
	}
	return nil
}

func DefaultRootDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.FatalE(context.Background(), "error determining user directory", err)
	}
	return filepath.Join(home, ".defradb")
}

func FileExists(filePath string) bool {
	statInfo, err := os.Stat(filePath)
	existsAsFile := (err == nil && !statInfo.IsDir())
	return existsAsFile
}

func FolderExists(folderPath string) bool {
	statInfo, err := os.Stat(folderPath)
	existsAsFolder := (err == nil && statInfo.IsDir())
	return existsAsFolder
}
