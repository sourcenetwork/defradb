// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package clitest

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcenetwork/defradb/config"
)

// Executing init command creates valid config file.
func TestCLIInitCommand(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	_, stderr := runDefraCommand(t, conf, []string{"init", "--rootdir", conf.rootDir})
	cfgfilePath := filepath.Join(conf.rootDir, config.DefaultConfigFileName)
	assertContainsSubstring(t, stderr, "Created config file at "+cfgfilePath)
	if !assert.FileExists(t, cfgfilePath) {
		t.Fatal("Config file not created")
	}
}

func TestCLIInitCommandTwiceErrors(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	cfgfilePath := filepath.Join(conf.rootDir, config.DefaultConfigFileName)
	_, stderr := runDefraCommand(t, conf, []string{"init", "--rootdir", conf.rootDir})
	assertContainsSubstring(t, stderr, "Created config file at "+cfgfilePath)
	_, stderr = runDefraCommand(t, conf, []string{"init", "--rootdir", conf.rootDir})
	assertContainsSubstring(t, stderr, "Configuration file already exists at "+cfgfilePath)
}

// Executing init command twice, but second time reinitializing.
func TestInitCommandTwiceReinitalize(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	cfgfilePath := filepath.Join(conf.rootDir, config.DefaultConfigFileName)
	_, stderr := runDefraCommand(t, conf, []string{"init", "--rootdir", conf.rootDir})
	assertContainsSubstring(t, stderr, "Created config file at "+cfgfilePath)
	_, stderr = runDefraCommand(t, conf, []string{"init", "--rootdir", conf.rootDir, "--reinitialize"})
	assertContainsSubstring(t, stderr, "Deleted config file at "+cfgfilePath)
	assertContainsSubstring(t, stderr, "Created config file at "+cfgfilePath)
}
