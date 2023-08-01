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
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcenetwork/defradb/config"
)

func TestPingSimple(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stopDefra := runDefraNode(t, conf)

	stdout, _ := runDefraCommand(t, conf, []string{"client", "ping"})

	nodeLog := stopDefra()

	assert.Contains(t, stdout, `{"data":{"response":"pong"}}`)
	for _, line := range nodeLog {
		assert.NotContains(t, line, "ERROR")
	}
}

func TestPingCommandToInvalidHost(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stopDefra := runDefraNode(t, conf)
	_, stderr := runDefraCommand(t, conf, []string{"client", "ping", "--url", "'1!2:3!4'"})

	nodeLog := stopDefra()

	for _, line := range nodeLog {
		assert.NotContains(t, line, "ERROR")
	}
	// for some line in stderr to contain the error message
	for _, line := range stderr {
		if strings.Contains(line, config.ErrFailedToValidateConfig.Error()) {
			return
		}
	}
	t.Error("expected error message not found in stderr")
}

func TestPingCommandNoHost(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	p, err := findFreePortInRange(t, 49152, 65535)
	assert.NoError(t, err)
	addr := fmt.Sprintf("localhost:%d", p)
	_, stderr := runDefraCommand(t, conf, []string{"client", "ping", "--url", addr})
	assertContainsSubstring(t, stderr, "failed to send ping")
}
