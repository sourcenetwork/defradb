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

import "testing"

func TestClientBlocksEmpty(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stdout, _ := runDefraCommand(t, conf, []string{"client", "blocks"})
	assertContainsSubstring(t, stdout, "Usage:")
}

func TestClientBlocksGetEmpty(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stdout, _ := runDefraCommand(t, conf, []string{"client", "blocks", "get"})
	assertContainsSubstring(t, stdout, "Usage:")
}

func TestClientBlocksGetInvalidCID(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stopDefra := runDefraNode(t, conf)
	stdout, _ := runDefraCommand(t, conf, []string{"client", "blocks", "get", "invalid-cid"})
	_ = stopDefra()
	assertContainsSubstring(t, stdout, "\"errors\"")
}

func TestClientBlocksGetNonExistentCID(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stopDefra := runDefraNode(t, conf)
	stdout, _ := runDefraCommand(t, conf, []string{"client", "blocks", "get", "bafybeieelb43ol5e5jiick2p7k4p577ph72ecwcuowlhbops4hpz24zhz4"})
	_ = stopDefra()
	assertContainsSubstring(t, stdout, "could not find")
}
