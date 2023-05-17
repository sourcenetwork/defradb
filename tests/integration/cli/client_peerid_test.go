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
)

func TestPeerID(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	stopDefra := runDefraNode(t, conf)

	stdout, _ := runDefraCommand(t, conf, []string{"client", "peerid"})

	defraLogLines := stopDefra()

	assertNotContainsSubstring(t, defraLogLines, "ERROR")

	assertContainsSubstring(t, stdout, "peerID")
}

func TestPeerIDWithNoHost(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	_, stderr := runDefraCommand(t, conf, []string{"client", "peerid"})
	assertContainsSubstring(t, stderr, "failed to request PeerID")
}
