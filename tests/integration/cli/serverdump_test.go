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

func TestServerDumpMemoryErrs(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	_, stderr := runDefraCommand(t, conf, []string{"server-dump", "--store", "memory"})
	assertContainsSubstring(t, stderr, "server-side dump is only supported for the Badger datastore")
}

func TestServerDumpInvalidStoreErrs(t *testing.T) {
	conf := NewDefraNodeDefaultConfig(t)
	_, stderr := runDefraCommand(t, conf, []string{"server-dump", "--store", "invalid"})
	// assertContainsSubstring(t, stderr, "invalid datastore type")
	assertContainsSubstring(t, stderr, "server-side dump is only supported for the Badger datastore")
}
