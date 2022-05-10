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
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestCreateRootDirWithDefaultConfig(t *testing.T) {
	tempdir := t.TempDir()
	rootdir := filepath.Join(tempdir, "defra_rootdir")
	CreateRootDirWithDefaultConfig(rootdir)
	if _, err := os.Stat(rootdir); os.IsNotExist(err) {
		t.Errorf("rootdir %q wasn't created properly, it does not exist", rootdir)
	}
	viper.SetConfigName(defaultDefraDBConfigFileName)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(rootdir)
	if err := viper.ReadInConfig(); err != nil {
		t.Errorf("could not read config file from rootdir %q", rootdir)
	}
}

func TestGetRootDirDefault(t *testing.T) {
	rootdir := ""
	d, _ := GetRootDir(rootdir)
	assert.Equal(t, DefaultRootDir(), d)
}
