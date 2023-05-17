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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRootCommandEmptyRootDir(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stdout, _ := runDefraCommand(t, conf, []string{})
	assert.Contains(t, stdout, "Usage:")
}

func TestRootCommandRootDirWithDefaultConfig(t *testing.T) {
	conf := DefraNodeConfig{
		logPath: t.TempDir(),
	}
	stdout, _ := runDefraCommand(t, conf, []string{})
	assert.Contains(t, stdout, "Usage:")
}

func TestRootCommandRootDirFromEnv(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stdout, _ := runDefraCommand(t, conf, []string{})
	assert.Contains(t, stdout, "Usage:")
}

func TestRootCommandRootWithNonexistentFlag(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stdout, _ := runDefraCommand(t, conf, []string{"--foo"})
	assert.Contains(t, stdout, "Usage:")
}
