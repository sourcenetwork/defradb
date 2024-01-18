// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cli

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

// The version information comes from the build process which is not [easily] accessible from unit tests.
// Therefore we test that the command outputs the expected formats *without the version info*.

// case: no args, meaning `--format text`
func TestVersionNoArg(t *testing.T) {
	cmd := MakeVersionCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	err := cmd.Execute()
	assert.NoError(t, err)
	t.Log(buf.String())
	assert.Contains(t, buf.String(), "defradb")
	assert.Contains(t, buf.String(), "built with Go")
}

// case: `--full`, meaning `--format text --full`
func TestVersionFull(t *testing.T) {
	cmd := MakeVersionCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--full"})
	err := cmd.Execute()
	assert.NoError(t, err)
	t.Log(buf.String())
	assert.Contains(t, buf.String(), "* HTTP API")
	assert.Contains(t, buf.String(), "* DocID versions")
	assert.Contains(t, buf.String(), "* P2P multicodec")
}

// case: `--format json`
func TestVersionJSON(t *testing.T) {
	cmd := MakeVersionCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--format", "json"})
	err := cmd.Execute()
	assert.NoError(t, err)
	t.Log(buf.String())
	assert.JSONEq(t, buf.String(), `
	{
		"release": "",
		"commit": "",
		"commitDate": "",
		"go": "",
		"httpAPI": "v0",
		"docIDVersions": "1",
		"netProtocol": "/defra/0.0.1"
	}`)
}

// case: `--format json --full` (is equivalent to previous one)
func TestVersionJSONFull(t *testing.T) {
	cmd := MakeVersionCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--format", "json", "--full"})
	err := cmd.Execute()
	assert.NoError(t, err)
	t.Log(buf.String())
	assert.JSONEq(t, buf.String(), `
	{
		"release": "",
		"commit": "",
		"commitDate": "",
		"go": "",
		"httpAPI": "v0",
		"docIDVersions": "1",
		"netProtocol": "/defra/0.0.1"
	}`)
}
