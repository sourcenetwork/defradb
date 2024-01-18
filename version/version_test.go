// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package version

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDefraVersion(t *testing.T) {
	dv, err := NewDefraVersion()
	assert.NoError(t, err)

	assert.NotEmpty(t, dv.VersionHTTPAPI)
	assert.NotEmpty(t, dv.NetProtocol)
	assert.NotEmpty(t, dv.DocIDVersions)

	// These variables are set in the Makefile via BUILD_FLAGS when building defradb.
	// This test assumes the test suite is not using these BUILD_FLAGS.
	// Therefore, we expect them to be empty in this unit test.
	assert.Empty(t, dv.GoInfo)
	assert.Empty(t, dv.Release)
	assert.Empty(t, dv.Commit)
	assert.Empty(t, dv.CommitDate)
}

func TestDefraVersionString(t *testing.T) {
	dv := defraVersion{
		Release:    "test-release",
		Commit:     "abc123def456",
		CommitDate: "2022-01-01T12:00:00Z",
		GoInfo:     "1.17.5",
	}
	assert.Equal(t, dv.String(), "defradb test-release (abc123de 2022-01-01T12:00:00Z) built with Go 1.17.5")
}

func TestDefraVersionStringFull(t *testing.T) {
	dv := defraVersion{
		Release:        "test-release",
		Commit:         "abc123def456",
		CommitDate:     "2022-01-01T12:00:00Z",
		GoInfo:         "1.17.5",
		VersionHTTPAPI: "v0",
		DocIDVersions:  "1",
		NetProtocol:    "/defra/0.0.1",
	}

	expected := `defradb test-release (abc123de 2022-01-01T12:00:00Z)
* HTTP API: v0
* P2P multicodec: /defra/0.0.1
* DocID versions: 1
* Go: 1.17.5`

	assert.Equal(t, expected, dv.StringFull())
}

func TestDefraVersion_JSON(t *testing.T) {
	dv1 := defraVersion{
		Release:        "test-release",
		Commit:         "abc123def456",
		CommitDate:     "2022-01-01T12:00:00Z",
		GoInfo:         "go1.17.5",
		VersionHTTPAPI: "1.2.3",
		DocIDVersions:  "0123456789abcdef",
		NetProtocol:    "test-protocol",
	}

	_, err := json.Marshal(dv1)
	assert.NoError(t, err)
}
