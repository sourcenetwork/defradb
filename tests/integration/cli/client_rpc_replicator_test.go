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

func TestReplicatorGetAllEmpty(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	portTCP, err := findFreePortInRange(t, 49152, 65535)
	if err != nil {
		t.Fatal(err)
	}
	conf.GRPCAddr = fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", portTCP)
	if err != nil {
		t.Fatal(err)
	}

	stopDefra := runDefraNode(t, conf)
	defer stopDefra()

	tcpAddr := fmt.Sprintf("localhost:%d", portTCP)
	_, stderr := runDefraCommand(t, conf, []string{"client", "--addr", tcpAddr, "rpc", "replicator", "getall"})
	assertContainsSubstring(t, stderr, "No replicator found")
}
