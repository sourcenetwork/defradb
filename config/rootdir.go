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
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sourcenetwork/defradb/logging"
	"github.com/spf13/viper"
)

func DefaultRootDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(context.Background(), fmt.Sprintf("could not get home directory: %v", home), logging.NewKV("Error", err))
	}
	return filepath.Join(home, defaultDefraDBRootDir)
}

// Returns rootdir path and whether it exists as directory, considering the env. variable and CLI flag.
func GetRootDir(rootDir string) (string, bool) {
	var err error
	var path string
	var exists bool
	err = viper.BindEnv("rootdir", defraEnvPrefix+"_ROOT")
	if err != nil {
		log.Fatal(context.Background(), "could not bind env variable", logging.NewKV("Error", err))
	}
	rootDirEnv := viper.GetString("rootdir")
	if rootDirEnv == "" && rootDir == "" {
		path = DefaultRootDir()
	} else if rootDirEnv != "" && rootDir == "" {
		path = rootDirEnv
	} else {
		path = rootDir
	}
	path, err = filepath.Abs(path)
	if err != nil {
		log.Fatal(context.Background(), fmt.Sprintf("could not get absolute path for %q", path), logging.NewKV("error", err))
	}
	info, err := os.Stat(path)
	exists = (err == nil && info.IsDir())
	return path, exists
}

func CreateRootDirWithDefaultConfig(rootDir string) {
	var err error
	err = os.MkdirAll(rootDir, defaultDirPerm)
	if err != nil {
		log.Fatal(context.Background(), fmt.Sprintf("could not create directory %q", rootDir), logging.NewKV("error", err))
	}
	cfg := DefaultConfig()
	err = cfg.WriteConfigFileToRootDir(rootDir)
	if err != nil {
		log.Fatal(
			context.Background(),
			fmt.Sprintf("could not write config file to directory %q", rootDir),
			logging.NewKV("error", err))
	}
}
