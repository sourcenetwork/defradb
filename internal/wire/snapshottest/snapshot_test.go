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
	_ "github.com/sourcenetwork/defradb/internal/core/block"
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

// TestWireSnapshot keeps the committed golden honest: it must match the current
// wire types (else a shape change went unrecorded). Set WIRE_SNAPSHOT_UPDATE=1 to
// regenerate it.
func TestWireSnapshot(t *testing.T) {
	got := wire.Snapshot()

	if os.Getenv("WIRE_SNAPSHOT_UPDATE") == "1" {
		require.NoError(t, os.WriteFile(goldenPath, []byte(got), 0o600))
		return
	}

	want, err := os.ReadFile(goldenPath)
	require.NoError(t, err, "read golden; create it with WIRE_SNAPSHOT_UPDATE=1")
	require.Equal(t, string(want), got,
		"wire types changed but the golden was not regenerated; run WIRE_SNAPSHOT_UPDATE=1 go test ./internal/wire/snapshottest")
}

// TestWireFormatChangeIsDocumented requires a note whenever the golden differs
// from the base branch: a changed golden means a wire type changed shape, and
// that change must be documented, the same bar as a data format change. Adding a
// note under docs/wire_format_changes is the acknowledgement. Skips when the base
// branch is not available (e.g. a local run with no fetched base).
func TestWireFormatChangeIsDocumented(t *testing.T) {
	base := baseRef(t)
	if base == "" {
		t.Skipf("base branch not found locally; skipping note check")
	}
	if !goldenChangedFrom(t, base) {
		return
	}
	require.True(t, noteAddedSince(t, base),
		"the wire snapshot changed from %s but no note was added under docs/%s; "+
			"add a short markdown file describing the change and whether old and new nodes can still talk",
		base, changesDir)
}

// goldenChangedFrom reports whether the golden file differs between base and HEAD.
func goldenChangedFrom(t *testing.T, base string) bool {
	out, err := runGit(t, "diff", "--name-only", base+"...HEAD", "--", relGoldenPath())
	require.NoError(t, err)
	return strings.TrimSpace(out) != ""
}

// noteAddedSince reports whether a note (not the README) was added under the wire
// format changes directory between base and HEAD.
func noteAddedSince(t *testing.T, base string) bool {
	dir := changesDocPackage + "/" + changesDir
	out, err := runGit(t, "diff", "--name-only", "--diff-filter=A", base+"...HEAD", "--", dir)
	require.NoError(t, err)
	for f := range strings.SplitSeq(strings.TrimSpace(out), "\n") {
		if strings.HasSuffix(f, ".md") && !strings.HasSuffix(f, "README.md") {
			return true
		}
	}
	return false
}

// relGoldenPath is the golden's path relative to the repo root, for git.
func relGoldenPath() string {
	return "internal/wire/snapshottest/" + goldenPath
}

// sourceBranchEnv is the change detector's base-branch variable, reused here so
// both checks share one notion of the base branch. Both read a bare branch name
// (e.g. develop). The change detector clones it; here it is resolved as a local
// remote ref (origin/<name>).
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

// runGit runs a git command from the repo root and returns its stdout. Running
// at the root makes repo-relative pathspecs resolve regardless of the test's cwd.
func runGit(t *testing.T, args ...string) (string, error) {
	t.Helper()
	full := append([]string{"-C", repoRoot(t)}, args...)
	out, err := exec.Command("git", full...).Output()
	return string(out), err
}

// repoRoot returns the repository root, found by walking up for a .git entry.
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	require.NoError(t, err)
	for {
		if _, err := os.Stat(dir + "/.git"); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("repo root (.git) not found")
		}
		dir = parent
	}
}
