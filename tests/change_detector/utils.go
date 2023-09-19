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
	// Repository is the url of the repository to run change detector on.
	Repository string
	// SourceBranch is the name of the source branch to run change detector on.
	SourceBranch string
	// TargetBranch is the name of the target branch to run change detector on.
	TargetBranch string
	// rootDatabaseDir is the shared database directory for running tests.
	rootDatabaseDir string
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
	TargetBranch = os.Getenv(targetBranchEnvName)
	rootDatabaseDir = os.Getenv(rootDataDirEnvName)

	if value, ok := os.LookupEnv(repositoryEnvName); ok {
		Repository = value
	} else {
		Repository = defaultRepository
	}

	if value, ok := os.LookupEnv(sourceBranchEnvName); ok {
		SourceBranch = value
	} else {
		SourceBranch = defaultSourceBranch
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
