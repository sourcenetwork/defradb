// Copyright 2023 Democratized Data Foundation
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
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func createUser(t *testing.T, conf DefraNodeConfig) {
	_, _ = runDefraCommand(t, conf, []string{
		"client", "query", `mutation { create_User(data: "{\"name\": \"John\"}") { _key } }`,
	})
}

func TestBackup_IfNoArgs_ShowUsage(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stdout, _ := runDefraCommand(t, conf, []string{"client", "backup"})
	assertContainsSubstring(t, stdout, "Usage:")
}

func TestBackupExport_ForAllCollections_ShouldExport(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stopDefra := runDefraNode(t, conf)

	createUserCollection(t, conf)

	createUser(t, conf)

	filepath := t.TempDir() + "/test.json"

	stdout, _ := runDefraCommand(t, conf, []string{
		"client", "backup", "export", filepath,
	})
	stopDefra()

	assertContainsSubstring(t, stdout, "success")

	b, err := os.ReadFile(filepath)
	require.NoError(t, err)
	require.Equal(
		t,
		`{"User":[{"_key":"bae-decf6467-4c7c-50d7-b09d-0a7097ef6bad","_newKey":"bae-decf6467-4c7c-50d7-b09d-0a7097ef6bad","name":"John"}]}`,
		string(b),
	)
}

func TestBackupExport_ForUserCollection_ShouldExport(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stopDefra := runDefraNode(t, conf)

	createUserCollection(t, conf)

	createUser(t, conf)

	filepath := t.TempDir() + "/test.json"

	stdout, _ := runDefraCommand(t, conf, []string{
		"client", "backup", "export", filepath, "--collections", "User",
	})
	stopDefra()

	assertContainsSubstring(t, stdout, "success")

	b, err := os.ReadFile(filepath)
	require.NoError(t, err)
	require.Equal(
		t,
		`{"User":[{"_key":"bae-decf6467-4c7c-50d7-b09d-0a7097ef6bad","_newKey":"bae-decf6467-4c7c-50d7-b09d-0a7097ef6bad","name":"John"}]}`,
		string(b),
	)
}

func TestBackupExport_ForInvalidCollection_ShouldFail(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stopDefra := runDefraNode(t, conf)

	createUserCollection(t, conf)

	createUser(t, conf)

	filepath := t.TempDir() + "/test.json"

	stdout, _ := runDefraCommand(t, conf, []string{
		"client", "backup", "export", filepath, "--collections", "Invalid",
	})
	stopDefra()

	assertContainsSubstring(t, stdout, "collection does not exist")
}

func TestBackupExport_InvalidFilePath_ShouldFail(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stopDefra := runDefraNode(t, conf)

	createUserCollection(t, conf)

	createUser(t, conf)

	filepath := t.TempDir() + "/some/test.json"

	stdout, _ := runDefraCommand(t, conf, []string{
		"client", "backup", "export", filepath, "--collections", "Invalid",
	})
	stopDefra()

	assertContainsSubstring(t, stdout, "invalid file path")
}
