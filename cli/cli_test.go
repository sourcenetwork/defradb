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
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"github.com/sourcenetwork/defradb/config"
)

// Verify that the top-level commands are registered, and if particular ones have subcommands.
func TestNewDefraCommand(t *testing.T) {
	expectedCommandNames := []string{
		"client",
		"init",
		"server-dump",
		"start",
		"version",
	}
	actualCommandNames := []string{}
	r := NewDefraCommand(config.DefaultConfig())
	for _, c := range r.RootCmd.Commands() {
		actualCommandNames = append(actualCommandNames, c.Name())
	}
	for _, expectedCommandName := range expectedCommandNames {
		assert.Contains(t, actualCommandNames, expectedCommandName)
	}
	for _, c := range r.RootCmd.Commands() {
		if c.Name() == "client" {
			assert.NotEmpty(t, c.Commands())
		}
	}
}

func TestAllHaveUsage(t *testing.T) {
	cfg := config.DefaultConfig()
	defra := NewDefraCommand(cfg)
	walkCommandTree(t, defra.RootCmd, func(c *cobra.Command) {
		assert.NotEmpty(t, c.Use)
	})
}

func walkCommandTree(t *testing.T, cmd *cobra.Command, f func(*cobra.Command)) {
	f(cmd)
	for _, c := range cmd.Commands() {
		walkCommandTree(t, c, f)
	}
}
