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
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
)

const (
	DefaultConfigFileName = "config.yaml"
	configType            = "yaml"
	defaultDirPerm        = 0o700
	defaultConfigFilePerm = 0o644
)

// defaultConfigTemplate must reflect the Config struct in content and configuration.
// All parameters must be represented here, to support Viper's automatic environment variable handling.
// We embed the default config template for clarity and to avoid autoformatters from breaking it.
//
//go:embed configfile_yaml.gotmpl
var defaultConfigTemplate string

func (cfg *Config) ConfigFilePath() string {
	return filepath.Join(cfg.Rootdir, DefaultConfigFileName)
}

func (cfg *Config) WriteConfigFile() error {
	path := cfg.ConfigFilePath()
	buffer, err := cfg.toBytes()
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, buffer, defaultConfigFilePerm); err != nil {
		return NewErrFailedToWriteFile(err, path)
	}
	log.FeedbackInfo(fmt.Sprintf("Created config file at %v", path))
	return nil
}

func (cfg *Config) DeleteConfigFile() error {
	if err := os.Remove(cfg.ConfigFilePath()); err != nil {
		return NewErrFailedToRemoveConfigFile(err)
	}
	log.FeedbackInfo(fmt.Sprintf("Deleted config file at %v", cfg.ConfigFilePath()))
	return nil
}

func (cfg *Config) CreateRootDirAndConfigFile() error {
	if err := os.MkdirAll(cfg.Rootdir, defaultDirPerm); err != nil {
		return err
	}
	log.FeedbackInfo(fmt.Sprintf("Created DefraDB root directory at %v", cfg.Rootdir))
	if err := cfg.WriteConfigFile(); err != nil {
		return err
	}
	return nil
}

func (cfg *Config) ConfigFileExists() bool {
	statInfo, err := os.Stat(cfg.ConfigFilePath())
	existsAsFile := (err == nil && !statInfo.IsDir())
	return existsAsFile
}

func DefaultRootDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.FatalE("error determining user directory", err)
	}
	return filepath.Join(home, ".defradb")
}

func FolderExists(folderPath string) bool {
	statInfo, err := os.Stat(folderPath)
	existsAsFolder := (err == nil && statInfo.IsDir())
	return existsAsFolder
}
