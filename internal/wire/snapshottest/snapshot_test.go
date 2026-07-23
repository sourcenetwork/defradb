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
// snapshot. Regenerate the golden with WIRE_SNAPSHOT_UPDATE=1 and, as with a
// data format change, add a note under docs/wire_format_changes explaining the
// cross-version impact. The note is what marks the change intentional: like the
// change detector, the test passes once a new note is present.
package snapshottest

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

const (
	goldenPath        = "wire_snapshot.golden"
	changesDir        = "wire_format_changes"
	changesDocPackage = "docs"
)

// TestWireSnapshot fails when a wire type changed shape unless the change is
// documented. Set WIRE_SNAPSHOT_UPDATE=1 to regenerate the golden.
func TestWireSnapshot(t *testing.T) {
	got := wire.Snapshot()

	if os.Getenv("WIRE_SNAPSHOT_UPDATE") == "1" {
		require.NoError(t, os.WriteFile(goldenPath, []byte(got), 0o600))
		return
	}

	want, err := os.ReadFile(goldenPath)
	require.NoError(t, err, "read golden; create it with WIRE_SNAPSHOT_UPDATE=1")

	if string(want) == got {
		return
	}

	if wireFormatChangeIsDocumented(t) {
		t.Skip("wire snapshot changed, but the change is documented")
	}

	t.Fatal("wire snapshot changed. If intentional, regenerate the golden with " +
		"WIRE_SNAPSHOT_UPDATE=1 and add a note under docs/wire_format_changes " +
		"describing the change and whether old and new nodes can still talk.")
}

// wireFormatChangeIsDocumented reports whether this branch adds a note under
// docs/wire_format_changes that is not on the base branch, mirroring how the
// change detector treats a new data_format_changes file as the acknowledgement.
func wireFormatChangeIsDocumented(t *testing.T) bool {
	base := baseRef(t)
	if base == "" {
		return false
	}
	dir := filepath.Join(repoRoot(t), changesDocPackage, changesDir)
	out, err := runGit(t, "diff", "--name-only", "--diff-filter=A", base+"...HEAD", "--", dir)
	if err != nil {
		return false
	}
	for f := range strings.SplitSeq(strings.TrimSpace(out), "\n") {
		if strings.HasSuffix(f, ".md") && !strings.HasSuffix(f, "README.md") {
			return true
		}
	}
	return false
}

// sourceBranchEnv is the change detector's base-branch variable, reused here so
// both checks share one notion of the base branch.
const sourceBranchEnv = "DEFRA_CHANGE_DETECTOR_SOURCE_BRANCH"

// baseRef returns the remote base branch to diff the note against: the branch
// named by sourceBranchEnv (default develop), falling back to origin/main if
// that ref does not exist locally.
func baseRef(t *testing.T) string {
	branch := os.Getenv(sourceBranchEnv)
	if branch == "" {
		branch = "develop"
	}
	for _, ref := range []string{"origin/" + branch, "origin/main"} {
		if _, err := runGit(t, "rev-parse", "--verify", "--quiet", ref); err == nil {
			return ref
		}
	}
	return ""
}

// repoRoot returns the repository root.
func repoRoot(t *testing.T) string {
	out, err := runGit(t, "rev-parse", "--show-toplevel")
	require.NoError(t, err)
	return strings.TrimSpace(out)
}

// runGit runs a git command and returns its stdout.
func runGit(t *testing.T, args ...string) (string, error) {
	t.Helper()
	out, err := exec.Command("git", args...).Output()
	return string(out), err
}
