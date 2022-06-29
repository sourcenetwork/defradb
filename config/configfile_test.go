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
	path := dir + "/" + defaultDefraDBConfigFileName
	_, err = os.Stat(path)
	assert.Nil(t, err)
}

func TestWritesConfigFileErroneousPath(t *testing.T) {
	cfg := DefaultConfig()
	dir := t.TempDir()
	err := cfg.WriteConfigFileToRootDir(dir + "////*&^^(*8769876////bar")
	assert.Error(t, err)
}
