// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

//go:build change_detector

package change_detector

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChanges(t *testing.T) {
	sourceRepoDir := t.TempDir()
	execClone(t, sourceRepoDir, Repository, SourceBranch)
	execMakeDeps(t, sourceRepoDir)

	var targetRepoDir string
	if TargetBranch == "" {
		// default to the local branch
		out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
		require.NoError(t, err, string(out))
		targetRepoDir = strings.TrimSpace(string(out))
	} else {
		// check out the target branch
		targetRepoDir = t.TempDir()
		execClone(t, targetRepoDir, Repository, TargetBranch)
		execMakeDeps(t, targetRepoDir)
	}

	if checkIfDatabaseFormatChangesAreDocumented(t, sourceRepoDir, targetRepoDir) {
		t.Skip("skipping test with documented database format changes")
	}

	targetRepoTestDir := filepath.Join(targetRepoDir, "tests", "integration")
	targetRepoPkgList := execList(t, targetRepoTestDir)

	sourceRepoTestDir := filepath.Join(sourceRepoDir, "tests", "integration")
	sourceRepoPkgList := execList(t, sourceRepoTestDir)

	sourceRepoPkgMap := make(map[string]bool)
	for _, pkg := range sourceRepoPkgList {
		sourceRepoPkgMap[pkg] = true
	}

	for _, pkg := range targetRepoPkgList {
		pkgName := strings.TrimPrefix(pkg, "github.com/sourcenetwork/defradb/")
		t.Run(pkgName, func(t *testing.T) {
			if pkg == "" || !sourceRepoPkgMap[pkg] {
				t.Skip("skipping unknown or new test package")
			}

			t.Parallel()
			dataDir := t.TempDir()

			sourceTestPkg := filepath.Join(sourceRepoDir, pkgName)
			execTest(t, sourceTestPkg, dataDir, true)

			targetTestPkg := filepath.Join(targetRepoDir, pkgName)
			execTest(t, targetTestPkg, dataDir, false)
		})
	}
}

// execList returns a list of all packages in the given directory.
func execList(t *testing.T, dir string) []string {
	cmd := exec.Command("go", "list", "./...")
	cmd.Dir = dir

	out, err := cmd.Output()
	require.NoError(t, err, string(out))

	return strings.Split(string(out), "\n")
}

// execTest runs the tests in the given directory and sets the data
// directory and setup only environment variables.
func execTest(t *testing.T, dir, dataDir string, setupOnly bool) {
	cmd := exec.Command("go", "test", ".", "-count", "1", "-v")
	cmd.Dir = dir
	cmd.Env = append(
		os.Environ(),
		fmt.Sprintf("%s=%s", enableEnvName, "true"),
		fmt.Sprintf("%s=%s", rootDataDirEnvName, dataDir),
	)

	if setupOnly {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", setupOnlyEnvName, "true"))
	}

	out, err := cmd.Output()
	require.NoError(t, err, string(out))
}

// execClone clones the repo from the given url and branch into the directory.
func execClone(t *testing.T, dir, url, branch string) {
	cmd := exec.Command(
		"git",
		"clone",
		"--single-branch",
		"--branch", branch,
		"--depth", "1",
		url,
		dir,
	)

	out, err := cmd.Output()
	require.NoError(t, err, string(out))
}

// execMakeDeps runs make:deps in the given directory.
func execMakeDeps(t *testing.T, dir string) {
	cmd := exec.Command("make", "deps:lens")
	cmd.Dir = dir

	out, err := cmd.Output()
	require.NoError(t, err, string(out))
}

func checkIfDatabaseFormatChangesAreDocumented(t *testing.T, sourceDir, targetDir string) bool {
	sourceChanges, ok := getDatabaseFormatDocumentation(t, sourceDir, false)
	require.True(t, ok, "Documentation directory not found")

	changes := make(map[string]struct{}, len(sourceChanges))
	for _, f := range sourceChanges {
		// Note: we assume flat directory for now - sub directories are not expanded
		changes[f.Name()] = struct{}{}
	}

	targetChanges, ok := getDatabaseFormatDocumentation(t, targetDir, true)
	require.True(t, ok, "Documentation directory not found")

	for _, f := range targetChanges {
		if _, isChangeOld := changes[f.Name()]; !isChangeOld {
			// If there is a new file in the directory then the change
			// has been documented and the test should pass
			return true
		}
	}

	return false
}

func getDatabaseFormatDocumentation(t *testing.T, startPath string, allowDescend bool) ([]fs.DirEntry, bool) {
	startInfo, err := os.Stat(startPath)
	require.NoError(t, err)

	var currentDirectory string
	if startInfo.IsDir() {
		currentDirectory = startPath
	} else {
		currentDirectory = path.Dir(startPath)
	}

	for {
		directoryContents, err := os.ReadDir(currentDirectory)
		require.NoError(t, err)

		for _, directoryItem := range directoryContents {
			directoryItemPath := path.Join(currentDirectory, directoryItem.Name())
			if directoryItem.Name() == documentationDirectoryName {
				probableFormatChangeDirectoryContents, err := os.ReadDir(directoryItemPath)
				require.NoError(t, err)

				for _, possibleDocumentationItem := range probableFormatChangeDirectoryContents {
					if path.Ext(possibleDocumentationItem.Name()) == ".md" {
						// If the directory's name matches the expected, and contains .md files
						// we assume it is the documentation directory
						return probableFormatChangeDirectoryContents, true
					}
				}
			} else {
				if directoryItem.IsDir() {
					childContents, directoryFound := getDatabaseFormatDocumentation(t, directoryItemPath, false)
					if directoryFound {
						return childContents, true
					}
				}
			}
		}

		if allowDescend {
			// If not found in this directory, continue down the path
			currentDirectory = path.Dir(currentDirectory)
			require.True(t, currentDirectory != "." && currentDirectory != "/")
		} else {
			return []fs.DirEntry{}, false
		}
	}
}
