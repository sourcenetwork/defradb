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

	"github.com/stretchr/testify/assert"
)

func TestCreateRootDirWithDefaultConfig(t *testing.T) {
	tempdir := t.TempDir()
	rootdir := filepath.Join(tempdir, "defra_rootdir")

	err := CreateRootDirWithDefaultConfig(rootdir)

	_, errStat := os.Stat(rootdir)
	notExists := os.IsNotExist(errStat)
	assert.Equal(t, false, notExists)
	assert.NoError(t, errStat)
	assert.NoError(t, err)
}

func TestGetRootDirDefault(t *testing.T) {
	rootdir := ""
	obtainedRootDir, _, err := GetRootDir(rootdir)
	defaultDir, errDefaultRootDir := DefaultRootDir()

	assert.NoError(t, err)
	assert.Equal(t, defaultDir, obtainedRootDir)
	assert.NoError(t, errDefaultRootDir)
}
