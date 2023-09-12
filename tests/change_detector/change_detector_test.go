// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package change_detector

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	repositoryEnvName   = "DEFRA_CODE_REPOSITORY"
	sourceBranchEnvName = "DEFRA_SOURCE_BRANCH"
	targetBranchEnvName = "DEFRA_TARGET_BRANCH"
)

func TestChanges(t *testing.T) {
	var repository string
	if value, ok := os.LookupEnv(repositoryEnvName); ok {
		repository = value
	} else {
		repository = "https://github.com/nasdf/defradb.git"
	}

	var sourceBranch string
	if value, ok := os.LookupEnv(sourceBranchEnvName); ok {
		sourceBranch = value
	} else {
		sourceBranch = "nasdf/test/parallel-change-detector"
	}

	var targetBranch string
	if value, ok := os.LookupEnv(targetBranchEnvName); ok {
		targetBranch = value
	} else {
		targetBranch = "nasdf/test/parallel-change-detector"
	}

	sourceRepoDir := t.TempDir()
	execClone(t, sourceRepoDir, repository, sourceBranch)

	targetRepoDir := t.TempDir()
	execClone(t, targetRepoDir, repository, targetBranch)

	execMakeDeps(t, sourceRepoDir)
	execMakeDeps(t, targetRepoDir)

	targetRepoTestDir := filepath.Join(targetRepoDir, "tests", "integration")
	targetRepoPkgList := execList(t, targetRepoTestDir)

	sourceRepoTestDir := filepath.Join(sourceRepoDir, "tests", "integration")
	sourceRepoPkgList := execList(t, sourceRepoTestDir)

	sourceRepoPkgMap := make(map[string]bool)
	for _, pkg := range sourceRepoPkgList {
		sourceRepoPkgMap[pkg] = true
	}

	for _, pkg := range targetRepoPkgList {
		if pkg == "" || !sourceRepoPkgMap[pkg] {
			continue
		}
		pkgName := strings.TrimPrefix(pkg, "github.com/sourcenetwork/defradb/")

		t.Run(pkgName, func(t *testing.T) {
			t.Parallel()
			dataDir := t.TempDir()

			fromTestPkg := filepath.Join(sourceRepoDir, pkgName)
			execTest(t, dataDir, fromTestPkg, true)

			toTestPkg := filepath.Join(targetRepoDir, pkgName)
			execTest(t, dataDir, toTestPkg, false)
		})
	}
}

func execList(t *testing.T, dir string) []string {
	cmd := exec.Command(
		"go",
		"list",
		"./...",
	)
	cmd.Dir = dir

	out, err := cmd.Output()
	require.NoError(t, err, string(out))

	return strings.Split(string(out), "\n")
}

func execTest(t *testing.T, dir, pkg string, setupOnly bool) {
	cmd := exec.Command(
		"go",
		"test",
		".",
		"-count", "1",
		"-v",
	)
	cmd.Dir = pkg
	cmd.Env = append(os.Environ(), "DEFRA_BADGER_FILE_PATH="+dir)
	cmd.Env = append(cmd.Env, "DEFRA_DETECT_DATABASE_CHANGES=true")

	if setupOnly {
		cmd.Env = append(cmd.Env, "DEFRA_SETUP_ONLY=true")
	}

	out, err := cmd.Output()
	require.NoError(t, err, string(out))
}

func execClone(t *testing.T, dir, url, branch string) {
	cmd := exec.Command(
		"git",
		"clone",
		"--branch", branch,
		"--single-branch",
		url,
		dir,
	)

	out, err := cmd.Output()
	require.NoError(t, err, string(out))
}

func execMakeDeps(t *testing.T, dir string) {
	cmd := exec.Command(
		"make",
		"deps:lens",
	)
	cmd.Dir = dir

	out, err := cmd.Output()
	require.NoError(t, err, string(out))
}
