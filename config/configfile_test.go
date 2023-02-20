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
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
)

func TestConfigTemplateSerialize(t *testing.T) {
	var buffer bytes.Buffer
	cfg := DefaultConfig()
	tmpl := template.New("configTemplate")
	configTemplate, err := tmpl.Parse(defaultConfigTemplate)
	if err != nil {
		t.Error(err)
	}
	if err := configTemplate.Execute(&buffer, cfg); err != nil {
		t.Error(err)
	}
	if _, err := cfg.ToJSON(); err != nil {
		t.Error(err)
	}
}

func TestConfigTemplateExecutes(t *testing.T) {
	cfg := DefaultConfig()
	var buffer bytes.Buffer
	tmpl := template.New("configTemplate")
	configTemplate, err := tmpl.Parse(defaultConfigTemplate)
	if err != nil {
		t.Error(err)
	}
	if err := configTemplate.Execute(&buffer, cfg); err != nil {
		t.Error(err)
	}
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
	if err != nil {
		t.Fatal(err)
	}

	assert.True(t, FileExists(cfg.ConfigFilePath()))

	cfgFromFile := DefaultConfig()
	cfgFromFile.Rootdir = tmpdir
	err = cfgFromFile.LoadWithRootdir(true)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, cfg.Log.Caller, cfgFromFile.Log.Caller)
	assert.Equal(t, cfg.Log.Format, cfgFromFile.Log.Format)
	assert.Equal(t, cfg.Log.Level, cfgFromFile.Log.Level)
	assert.Equal(t, cfg.Log.NoColor, cfgFromFile.Log.NoColor)
	assert.Equal(t, cfg.Log.Output, cfgFromFile.Log.Output)
	assert.Equal(t, cfg.Log.Stacktrace, cfgFromFile.Log.Stacktrace)
}

func ReadAndPrintFile(t *testing.T, path string) {
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	buf := make([]byte, 1024)
	for {
		n, err := f.Read(buf)
		if err != nil {
			break
		}
		t.Log(string(buf[:n]))
	}
}

func TestReadConfigFileForDatastore(t *testing.T) {
	tmpdir := t.TempDir()

	cfg := DefaultConfig()
	cfg.Rootdir = tmpdir
	cfg.Datastore.Store = "badger"
	cfg.Datastore.Badger.Path = "dataPath"
	cfg.Datastore.Badger.ValueLogFileSize = 512 * MiB

	err := cfg.WriteConfigFile()
	assert.NoError(t, err)

	configPath := filepath.Join(tmpdir, DefaultConfigFileName)
	_, err = os.Stat(configPath)
	assert.NoError(t, err)

	cfgFromFile := DefaultConfig()
	cfgFromFile.Rootdir = tmpdir
	err = cfgFromFile.LoadWithRootdir(true)
	assert.NoError(t, err)

	assert.Equal(t, cfg.Datastore.Store, cfgFromFile.Datastore.Store)
	assert.Equal(t, filepath.Join(tmpdir, cfg.Datastore.Badger.Path), cfgFromFile.Datastore.Badger.Path)
	assert.Equal(t, cfg.Datastore.Badger.ValueLogFileSize, cfgFromFile.Datastore.Badger.ValueLogFileSize)
}

func TestFileExists(t *testing.T) {
	tmpdir := t.TempDir()
	// Verify that a file that doesn't exist returns false.
	assert.False(t, FileExists(filepath.Join(tmpdir, "nonexistentfile")))

	// Verify that a file that does exist returns true.
	fpath := filepath.Join(tmpdir, "file")
	f, err := os.Create(fpath)
	f.Close()
	assert.NoError(t, err)
	assert.True(t, FileExists(fpath))

	// Test that a directory is not considered a file.
	dpath := filepath.Join(tmpdir, "dir")
	err = os.Mkdir(dpath, 0755)
	assert.NoError(t, err)
	assert.False(t, FileExists(dpath))
}

func TestConfigFileExists(t *testing.T) {
	cfg := DefaultConfig()
	tmpdir := t.TempDir()
	cfg.Rootdir = tmpdir
	assert.False(t, FileExists(cfg.ConfigFilePath()))
	err := cfg.WriteConfigFile()
	assert.NoError(t, err)
	assert.True(t, FileExists(cfg.ConfigFilePath()))
}

func TestConfigFileExistsErroneousPath(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Rootdir = filepath.Join(t.TempDir(), "////*&^^(*8769876////bar")
	assert.False(t, FileExists(cfg.ConfigFilePath()))
}

func TestInvalidConfigDatastore(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Datastore.Badger.Path = "[][][]"

	err := cfg.LoadWithRootdir(false)
	assert.Error(t, err)
}

func TestDeleteConfigFile(t *testing.T) {
	cfg := DefaultConfig()
	tmpdir := t.TempDir()
	cfg.Rootdir = tmpdir
	err := cfg.WriteConfigFile()
	assert.NoError(t, err)

	assert.True(t, FileExists(cfg.ConfigFilePath()))

	err = cfg.DeleteConfigFile()
	assert.NoError(t, err)
	assert.False(t, FileExists(cfg.ConfigFilePath()))
}
