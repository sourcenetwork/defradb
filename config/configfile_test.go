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
	dir := t.TempDir()
	err := cfg.WriteConfigFileToRootDir(dir)
	assert.NoError(t, err)
	path := dir + "/" + DefaultDefraDBConfigFileName
	_, err = os.Stat(path)
	assert.Nil(t, err)
}

func TestWritesConfigFileErroneousPath(t *testing.T) {
	cfg := DefaultConfig()
	dir := t.TempDir()
	err := cfg.WriteConfigFileToRootDir(dir + "////*&^^(*8769876////bar")
	assert.Error(t, err)
}

func TestReadConfigFileForLogger(t *testing.T) {
	dir := t.TempDir()

	cfg := DefaultConfig()
	cfg.Log.Caller = true
	cfg.Log.Format = "json"
	cfg.Log.Level = logLevelDebug
	cfg.Log.NoColor = true
	cfg.Log.Output = dir + "/log.txt"
	cfg.Log.Stacktrace = true

	err := cfg.WriteConfigFileToRootDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	path := dir + "/" + DefaultDefraDBConfigFileName

	_, err = os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}

	cfgFromFile := DefaultConfig()

	err = cfgFromFile.Load(dir)
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

func TestReadConfigFileForDatastore(t *testing.T) {
	dir := t.TempDir()

	cfg := DefaultConfig()
	cfg.Datastore.Store = "badger"
	cfg.Datastore.Badger.Path = "dataPath"
	cfg.Datastore.Badger.ValueLogFileSize = 512 * MiB
	cfg.Datastore.MaxRetries = 3

	err := cfg.WriteConfigFileToRootDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	path := dir + "/" + DefaultDefraDBConfigFileName

	_, err = os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}

	cfgFromFile := DefaultConfig()

	err = cfgFromFile.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, cfg.Datastore.Store, cfgFromFile.Datastore.Store)
	assert.Equal(t, dir+"/"+cfg.Datastore.Badger.Path, cfgFromFile.Datastore.Badger.Path)
	assert.Equal(t, cfg.Datastore.Badger.ValueLogFileSize, cfgFromFile.Datastore.Badger.ValueLogFileSize)
	assert.Equal(t, cfg.Datastore.MaxRetries, cfgFromFile.Datastore.MaxRetries)
}
