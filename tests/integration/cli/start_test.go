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
	"testing"
)

func TestStartCommandBasic(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	_, stderr := runDefraCommand(t, conf, []string{
		"start",
		"--url", conf.APIURL,
		"--tcpaddr", conf.GRPCAddr,
	})
	assertContainsSubstring(t, stderr, "Starting DefraDB service...")
	assertNotContainsSubstring(t, stderr, "Error")
}

func TestStartCommandWithTLSIncomplete(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	_, stderr := runDefraCommand(t, conf, []string{
		"start",
		"--tls",
		"--url", conf.APIURL,
		"--tcpaddr", conf.GRPCAddr,
	})
	assertContainsSubstring(t, stderr, "Starting DefraDB service...")
	assertContainsSubstring(t, stderr, "Error")
}

func TestStartCommandWithStoreMemory(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	_, stderr := runDefraCommand(t, conf, []string{
		"start", "--store", "memory",
		"--url", conf.APIURL,
		"--tcpaddr", conf.GRPCAddr,
	})
	assertContainsSubstring(t, stderr, "Starting DefraDB service...")
	assertContainsSubstring(t, stderr, "Building new memory store")
	assertNotContainsSubstring(t, stderr, "Error")
}

func TestStartCommandWithP2PAddr(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	p2pport, err := findFreePortInRange(t, 49152, 65535)
	if err != nil {
		t.Fatal(err)
	}
	addr := fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", p2pport)
	_, stderr := runDefraCommand(t, conf, []string{
		"start",
		"--p2paddr", addr,
		"--url", conf.APIURL,
		"--tcpaddr", conf.GRPCAddr,
	})
	assertContainsSubstring(t, stderr, "Starting DefraDB service...")
	logstring := fmt.Sprintf("Starting P2P node, {\"P2P address\": \"%s\"}", addr)
	assertContainsSubstring(t, stderr, logstring)
	assertNotContainsSubstring(t, stderr, "Error")
}

func TestStartCommandWithNoP2P(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	_, stderr := runDefraCommand(t, conf, []string{
		"start",
		"--no-p2p",
	})
	assertContainsSubstring(t, stderr, "Starting DefraDB service...")
	assertNotContainsSubstring(t, stderr, "Starting P2P node")
	assertNotContainsSubstring(t, stderr, "Error")
}

func TestStartCommandWithInvalidStoreType(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	_, stderr := runDefraCommand(t, conf, []string{
		"start",
		"--store", "invalid",
	})
	assertContainsSubstring(t, stderr, "failed to load config: failed to validate config: invalid store type")
}
