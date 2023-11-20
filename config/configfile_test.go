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
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
)

func TestConfigTemplateSerialize(t *testing.T) {
	var buffer bytes.Buffer
	cfg := DefaultConfig()
	tmpl := template.New("configTemplate")
	configTemplate, err := tmpl.Parse(defaultConfigTemplate)
	assert.NoError(t, err)
	err = configTemplate.Execute(&buffer, cfg)
	assert.NoError(t, err)
	_, err = cfg.ToJSON()
	assert.NoError(t, err)
}

func TestConfigTemplateExecutes(t *testing.T) {
	cfg := DefaultConfig()
	var buffer bytes.Buffer
	tmpl := template.New("configTemplate")
	configTemplate, err := tmpl.Parse(defaultConfigTemplate)
	assert.NoError(t, err)
	err = configTemplate.Execute(&buffer, cfg)
	assert.NoError(t, err)
}

func TestWritesConfigFile(t *testing.T) {
	cfg := DefaultConfig()
	tmpdir := t.TempDir()
	cfg.Rootdir = tmpdir
	err := cfg.WriteConfigFile()
	assert.NoError(t, err)
	path := filepath.Join(tmpdir, DefaultConfigFileName)
	_, err = os.Stat(path)
	assert.Nil(t, err)
}

func TestWritesConfigFileErroneousPath(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Rootdir = filepath.Join(t.TempDir(), "////*&^^(*8769876////bar")
	err := cfg.WriteConfigFile()
	assert.Error(t, err)
}

func TestReadConfigFileForLogger(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skipf("Test is not supported on windows as it leaks resources, see https://github.com/sourcenetwork/defradb/issues/2057")
	}

	cfg := DefaultConfig()
	tmpdir := t.TempDir()
	cfg.Rootdir = tmpdir
	cfg.Log.Caller = true
	cfg.Log.Format = "json"
	cfg.Log.Level = logLevelDebug
	cfg.Log.NoColor = true
	cfg.Log.Output = filepath.Join(tmpdir, "log.txt")
	cfg.Log.Stacktrace = true

	err := cfg.WriteConfigFile()
	assert.NoError(t, err)

	assert.True(t, cfg.ConfigFileExists())

	cfgFromFile := DefaultConfig()
	err = cfgFromFile.setRootdir(tmpdir)
	assert.NoError(t, err)
	err = cfgFromFile.LoadWithRootdir(true)
	assert.NoError(t, err)

	assert.Equal(t, cfg.Log.Caller, cfgFromFile.Log.Caller)
	assert.Equal(t, cfg.Log.Format, cfgFromFile.Log.Format)
	assert.Equal(t, cfg.Log.Level, cfgFromFile.Log.Level)
	assert.Equal(t, cfg.Log.NoColor, cfgFromFile.Log.NoColor)
	assert.Equal(t, cfg.Log.Output, cfgFromFile.Log.Output)
	assert.Equal(t, cfg.Log.Stacktrace, cfgFromFile.Log.Stacktrace)
}

func TestReadConfigFileForDatastore(t *testing.T) {
	tmpdir := t.TempDir()

	cfg := DefaultConfig()
	err := cfg.setRootdir(tmpdir)
	assert.NoError(t, err)
	cfg.Datastore.Store = "badger"
	cfg.Datastore.Badger.Path = "dataPath"
	cfg.Datastore.Badger.ValueLogFileSize = 512 * MiB

	err = cfg.WriteConfigFile()
	assert.NoError(t, err)

	configPath := filepath.Join(tmpdir, DefaultConfigFileName)
	_, err = os.Stat(configPath)
	assert.NoError(t, err)

	cfgFromFile := DefaultConfig()
	err = cfgFromFile.setRootdir(tmpdir)
	assert.NoError(t, err)
	err = cfgFromFile.LoadWithRootdir(true)
	assert.NoError(t, err)

	assert.Equal(t, cfg.Datastore.Store, cfgFromFile.Datastore.Store)
	assert.Equal(t, filepath.Join(tmpdir, cfg.Datastore.Badger.Path), cfgFromFile.Datastore.Badger.Path)
	assert.Equal(t, cfg.Datastore.Badger.ValueLogFileSize, cfgFromFile.Datastore.Badger.ValueLogFileSize)
}

func TestConfigFileExists(t *testing.T) {
	cfg := DefaultConfig()
	err := cfg.setRootdir(t.TempDir())
	assert.NoError(t, err)
	assert.False(t, cfg.ConfigFileExists())

	err = cfg.WriteConfigFile()
	assert.NoError(t, err)
	assert.True(t, cfg.ConfigFileExists())
}

func TestConfigFileExistsErroneousPath(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Rootdir = filepath.Join(t.TempDir(), "////*&^^(*8769876////bar")
	assert.False(t, cfg.ConfigFileExists())
}

func TestDeleteConfigFile(t *testing.T) {
	cfg := DefaultConfig()
	tmpdir := t.TempDir()
	cfg.Rootdir = tmpdir
	err := cfg.WriteConfigFile()
	assert.NoError(t, err)

	assert.True(t, cfg.ConfigFileExists())

	err = cfg.DeleteConfigFile()
	assert.NoError(t, err)
	assert.False(t, cfg.ConfigFileExists())
}
