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
	"fmt"
	"math/rand"
	"net"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"github.com/sourcenetwork/defradb/config"
	"github.com/sourcenetwork/defradb/errors"
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

// findFreePortInRange returns a free port in the range [minPort, maxPort].
// The range of ports that are unfrequently used is [49152, 65535].
func findFreePortInRange(minPort, maxPort int) (int, error) {
	if minPort < 1 || maxPort > 65535 || minPort > maxPort {
		return 0, errors.New("invalid port range")
	}

	const maxAttempts = 100
	for i := 0; i < maxAttempts; i++ {
		port := rand.Intn(maxPort-minPort+1) + minPort
		addr := fmt.Sprintf("127.0.0.1:%d", port)
		listener, err := net.Listen("tcp", addr)
		if err == nil {
			_ = listener.Close()
			return port, nil
		}
	}

	return 0, errors.New("unable to find a free port")
}
