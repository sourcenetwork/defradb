// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

// Package snapshottest verifies the wire snapshot against a committed golden
// file. It blank-imports every package that registers a wire type so the
// registry is fully populated, then compares wire.Snapshot() to the golden.
//
// A wire-format change (a renamed or retyped field on a wire type) changes the
// snapshot and fails this test. Regenerate the golden with WIRE_SNAPSHOT_UPDATE=1
// once the change is intentional; the diff in that golden file is the record of it.
package snapshottest

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/internal/wire"

	// Blank imports run each package's init so its wire types register.
	_ "github.com/sourcenetwork/defradb/internal/db/p2p"
	_ "github.com/sourcenetwork/defradb/internal/db/p2p/message"
	_ "github.com/sourcenetwork/defradb/internal/db/p2p/protocol"
	_ "github.com/sourcenetwork/defradb/internal/kms"
	_ "github.com/sourcenetwork/defradb/internal/se"
)

const goldenPath = "wire_snapshot.golden"

// Set WIRE_SNAPSHOT_UPDATE=1 to regenerate the golden. An env var avoids a
// test-flag collision with other packages linked into this binary.
func TestWireSnapshot(t *testing.T) {
	got := wire.Snapshot()

	if os.Getenv("WIRE_SNAPSHOT_UPDATE") == "1" {
		require.NoError(t, os.WriteFile(goldenPath, []byte(got), 0o600))
		return
	}

	want, err := os.ReadFile(goldenPath)
	require.NoError(t, err, "read golden; run with -update to create it")
	require.Equal(t, string(want), got,
		"wire snapshot changed; if intentional, regenerate with: WIRE_SNAPSHOT_UPDATE=1 go test ./internal/wire/snapshottest")
}
