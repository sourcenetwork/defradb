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

func TestBackupImport_WithValidFile_ShouldImport(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stopDefra := runDefraNode(t, conf)

	createUserCollection(t, conf)

	filepath := t.TempDir() + "/test.json"

	err := os.WriteFile(
		filepath,
		[]byte(`{"User":[{"_key":"bae-decf6467-4c7c-50d7-b09d-0a7097ef6bad","_newKey":"bae-decf6467-4c7c-50d7-b09d-0a7097ef6bad","name":"John"}]}`),
		0644,
	)
	require.NoError(t, err)

	stdout, _ := runDefraCommand(t, conf, []string{
		"client", "backup", "import", filepath,
	})
	stopDefra()

	assertContainsSubstring(t, stdout, "success")
}

func TestBackupImport_WithExistingDoc_ShouldFail(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stopDefra := runDefraNode(t, conf)

	createUserCollection(t, conf)

	createUser(t, conf)

	filepath := t.TempDir() + "/test.json"

	err := os.WriteFile(
		filepath,
		[]byte(`{"User":[{"_key":"bae-decf6467-4c7c-50d7-b09d-0a7097ef6bad","_newKey":"bae-decf6467-4c7c-50d7-b09d-0a7097ef6bad","name":"John"}]}`),
		0644,
	)
	require.NoError(t, err)

	stdout, _ := runDefraCommand(t, conf, []string{
		"client", "backup", "import", filepath,
	})
	stopDefra()

	assertContainsSubstring(t, stdout, "a document with the given dockey already exists")
}

func TestBackupImport_ForInvalidCollection_ShouldFail(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stopDefra := runDefraNode(t, conf)

	createUserCollection(t, conf)

	createUser(t, conf)

	filepath := t.TempDir() + "/test.json"

	err := os.WriteFile(
		filepath,
		[]byte(`{"Invalid":[{"_key":"bae-decf6467-4c7c-50d7-b09d-0a7097ef6bad","_newKey":"bae-decf6467-4c7c-50d7-b09d-0a7097ef6bad","name":"John"}]}`),
		0644,
	)
	require.NoError(t, err)

	stdout, _ := runDefraCommand(t, conf, []string{
		"client", "backup", "import", filepath,
	})
	stopDefra()

	assertContainsSubstring(t, stdout, "failed to get collection: datastore: key not found. Name: Invalid")
}

func TestBackupImport_InvalidFilePath_ShouldFail(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stopDefra := runDefraNode(t, conf)

	createUserCollection(t, conf)

	createUser(t, conf)

	filepath := t.TempDir() + "/some/test.json"

	stdout, _ := runDefraCommand(t, conf, []string{
		"client", "backup", "import", filepath,
	})
	stopDefra()

	assertContainsSubstring(t, stdout, "invalid file path")
}
