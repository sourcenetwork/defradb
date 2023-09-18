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
	"io/fs"
	"os"
	"path"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	// Enabled is true when the change detector is running.
	Enabled bool
	// SetupOnly is true when the change detector is running in setup mode.
	SetupOnly bool
	// rootDatabaseDir is the shared database directory for running tests.
	rootDatabaseDir string
	// repository is the url of the repository to run change detector on.
	repository string
	// sourceBranch is the name of the source branch to run change detector on.
	sourceBranch string
	// targetBranch is the name of the target branch to run change detector on.
	targetBranch string
	// previousTestCaseTestName is the name of the previous test.
	previousTestCaseTestName string
)

const (
	repositoryEnvName   = "DEFRA_CHANGE_DETECTOR_REPOSITORY"
	sourceBranchEnvName = "DEFRA_CHANGE_DETECTOR_SOURCE_BRANCH"
	targetBranchEnvName = "DEFRA_CHANGE_DETECTOR_TARGET_BRANCH"
	setupOnlyEnvName    = "DEFRA_CHANGE_DETECTOR_SETUP_ONLY"
	rootDataDirEnvName  = "DEFRA_CHANGE_DETECTOR_ROOT_DATA_DIR"
	enableEnvName       = "DEFRA_CHANGE_DETECTOR_ENABLE"
)

const (
	defaultRepository          = "https://github.com/sourcenetwork/defradb.git"
	defaultSourceBranch        = "develop"
	documentationDirectoryName = "data_format_changes"
)

func init() {
	Enabled, _ = strconv.ParseBool(os.Getenv(enableEnvName))
	SetupOnly, _ = strconv.ParseBool(os.Getenv(setupOnlyEnvName))
	rootDatabaseDir = os.Getenv(rootDataDirEnvName)
	targetBranch = os.Getenv(targetBranchEnvName)

	if value, ok := os.LookupEnv(repositoryEnvName); ok {
		repository = value
	} else {
		repository = defaultRepository
	}

	if value, ok := os.LookupEnv(sourceBranchEnvName); ok {
		sourceBranch = value
	} else {
		sourceBranch = defaultSourceBranch
	}
}

// DatabaseDir returns the database directory for change detector test.
func DatabaseDir(t testing.TB) string {
	return path.Join(rootDatabaseDir, t.Name())
}

// PreTestChecks skips any test that can't be run by the change detector.
func PreTestChecks(t *testing.T, collectionNames []string) {
	if !Enabled {
		return
	}

	if previousTestCaseTestName == t.Name() {
		t.Skip("skipping duplicate test")
	}
	previousTestCaseTestName = t.Name()

	if len(collectionNames) == 0 {
		t.Skip("skipping test with no collections")
	}

	if SetupOnly {
		return
	}

	_, err := os.Stat(DatabaseDir(t))
	if os.IsNotExist(err) {
		t.Skip("skipping new test package")
	}
	require.NoError(t, err)
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
