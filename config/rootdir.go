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
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

func DefaultRootDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, defaultDefraDBRootDir), nil
}

// Returns rootdir path and whether it exists as directory, considering the env. variable and CLI flag.
func GetRootDir(rootDir string) (string, bool, error) {
	var err error
	var path string
	var exists bool
	err = viper.BindEnv("rootdir", defraEnvPrefix+"_ROOT")
	if err != nil {
		return "", false, err
	}
	rootDirEnv := viper.GetString("rootdir")
	if rootDirEnv == "" && rootDir == "" {
		path, err = DefaultRootDir()
		if err != nil {
			return "", false, err
		}
	} else if rootDirEnv != "" && rootDir == "" {
		path = rootDirEnv
	} else {
		path = rootDir
	}
	path, err = filepath.Abs(path)
	if err != nil {
		return "", false, err
	}
	info, err := os.Stat(path)
	exists = (err == nil && info.IsDir())
	return path, exists, nil
}

func CreateRootDirWithDefaultConfig(rootDir string) error {
	var err error
	err = os.MkdirAll(rootDir, defaultDirPerm)
	if err != nil {
		return err
	}
	cfg := DefaultConfig()
	err = cfg.WriteConfigFileToRootDir(rootDir)
	if err != nil {
		return err
	}
	return nil
}
