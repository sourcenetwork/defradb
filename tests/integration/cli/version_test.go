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
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// note: this assumes the version information *without* build-time info integrated.
func TestExecVersion(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stdout, stderr := runDefraCommand(t, conf, []string{"version"})
	for _, line := range stderr {
		assert.NotContains(t, line, "ERROR")
	}
	output := strings.Join(stdout, " ")
	assert.Contains(t, output, "defradb")
	assert.Contains(t, output, "built with Go")
}

func TestExecVersionJSON(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stdout, stderr := runDefraCommand(t, conf, []string{"version", "--format", "json"})
	for _, line := range stderr {
		assert.NotContains(t, line, "ERROR")
	}
	output := strings.Join(stdout, " ")
	assert.Contains(t, output, "go\":")
	assert.Contains(t, output, "commit\":")
	assert.Contains(t, output, "commitdate\":")
	var data map[string]any
	err := json.Unmarshal([]byte(output), &data)
	assert.NoError(t, err)
}
