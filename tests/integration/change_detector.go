// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package tests

import (
	"context"
	"fmt"
	"io/fs"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var skip bool

func IsDetectingDbChanges() bool {
	return DetectDbChanges
}

// Returns true if test should pass early
func DetectDbChangesPreTestChecks(
	t *testing.T,
	collectionNames []string,
) bool {
	if skip {
		t.SkipNow()
	}

	if previousTestCaseTestName == t.Name() {
		// The database format changer currently only supports running the first test
		//  case, if a second case is detected we return early
		return true
	}
	previousTestCaseTestName = t.Name()

	if areDatabaseFormatChangesDocumented {
		// If we are checking that database formatting changes have been made and
		//  documented, and changes are documented, then the tests can all pass.
		return true
	}

	if len(collectionNames) == 0 {
		// If the test doesn't specify any collections, then we can't use it to check
		//  the database format, so we skip it
		t.SkipNow()
	}

	if !SetupOnly {
		dbDirectory := path.Join(rootDatabaseDir, t.Name())
		_, err := os.Stat(dbDirectory)
		if os.IsNotExist(err) {
			// This is a new test that does not exist in the target branch, we should
			// skip it.
			t.SkipNow()
		} else {
			require.NoError(t, err)
		}
	}

	return false
}

func detectDbChangesInit(repository string, targetBranch string) {
	badgerFile = true
	badgerInMemory = false

	if SetupOnly {
		// Only the primary test process should perform the setup below
		return
	}

	defraTempDir := path.Join(os.TempDir(), "defradb")
	changeDetectorTempDir := path.Join(defraTempDir, "tests", "changeDetector")

	latestTargetCommitHash := getLatestCommit(repository, targetBranch)
	detectDbChangesCodeDir = path.Join(changeDetectorTempDir, "code", latestTargetCommitHash)
	rand.Seed(time.Now().Unix())
	randNumber := rand.Int()
	dbsDir := path.Join(changeDetectorTempDir, "dbs", fmt.Sprint(randNumber))

	testPackagePath, isIntegrationTest := getTestPackagePath()
	if !isIntegrationTest {
		skip = true
		return
	}
	rootDatabaseDir = path.Join(dbsDir, strings.ReplaceAll(testPackagePath, "/", "_"))

	_, err := os.Stat(detectDbChangesCodeDir)
	// Warning - there is a race condition here, where if running multiple packages in
	//  parallel (as per default) against a new target commit multiple test pacakges will
	//  try and clone the target branch at the same time (and will fail).
	// This could be solved by using a file lock or similar, however running the change
	//  detector in parallel is significantly slower than running it serially due to machine
	//  resource constraints, so I am leaving the race condition in and recommending running
	//  the change detector with the CLI args `-p 1`
	if os.IsNotExist(err) {
		cloneCmd := exec.Command(
			"git",
			"clone",
			"-b",
			targetBranch,
			"--single-branch",
			repository,
			detectDbChangesCodeDir,
		)
		cloneCmd.Stdout = os.Stdout
		cloneCmd.Stderr = os.Stderr
		err := cloneCmd.Run()
		if err != nil {
			panic(err)
		}
	} else if err != nil {
		panic(err)
	} else {
		// Cache must be cleaned, or it might not run the test setup!
		// Note: this also acts as a race condition if multiple build are running against the
		//       same target if this happens some tests might be silently skipped if the
		//       child-setup fails.  Currently I think it is worth it for slightly faster build
		//       times, but feel very free to change this!
		goTestCacheCmd := exec.Command("go", "clean", "-testcache")
		goTestCacheCmd.Dir = detectDbChangesCodeDir
		err = goTestCacheCmd.Run()
		if err != nil {
			panic(err)
		}
	}

	areDatabaseFormatChangesDocumented = checkIfDatabaseFormatChangesAreDocumented()
	if areDatabaseFormatChangesDocumented {
		// Dont bother doing anything if the changes are documented
		return
	}

	targetTestPackage := detectDbChangesCodeDir + "/tests/integration/" + testPackagePath

	_, err = os.Stat(targetTestPackage)
	if os.IsNotExist(err) {
		// This is a new test package, and thus the change detector is not applicable
		// as the tests do not exist in the target branch.
		skip = true
		return
	} else if err != nil {
		panic(err)
	}

	// If we are checking for database changes, and we are not seting up the database,
	// then we must be in the main test process, and need to create a new process
	// setting up the database for this test using the old branch We should not setup
	// the database using the current branch/process
	goTestCmd := exec.Command(
		"go",
		"test",
		"./...",
		"-v",
	)

	goTestCmd.Dir = targetTestPackage
	goTestCmd.Env = os.Environ()
	goTestCmd.Env = append(
		goTestCmd.Env,
		setupOnlyEnvName+"=true",
		rootDBFilePathEnvName+"="+rootDatabaseDir,
	)
	out, err := goTestCmd.Output()
	if err != nil {
		log.ErrorE(context.TODO(), string(out), err)
		panic(err)
	}
}

// getTestPackagePath returns the path to the package currently under test, relative
// to `./tests/integration/`. Will return an empty string and false if the tests
// are not within that directory.
func getTestPackagePath() (string, bool) {
	currentTestPackage, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	splitPath := strings.Split(
		currentTestPackage,
		"/tests/integration/",
	)

	if len(splitPath) != 2 {
		return "", false
	}
	return splitPath[1], true
}

func checkIfDatabaseFormatChangesAreDocumented() bool {
	previousDbChangeFiles, targetDirFound := getDatabaseFormatDocumentation(
		detectDbChangesCodeDir,
		false,
	)
	if !targetDirFound {
		panic("Documentation directory not found")
	}

	previousDbChanges := make(map[string]struct{}, len(previousDbChangeFiles))
	for _, f := range previousDbChangeFiles {
		// Note: we assume flat directory for now - sub directories are not expanded
		previousDbChanges[f.Name()] = struct{}{}
	}

	_, thisFilePath, _, _ := runtime.Caller(0)
	currentDbChanges, currentDirFound := getDatabaseFormatDocumentation(thisFilePath, true)
	if !currentDirFound {
		panic("Documentation directory not found")
	}

	for _, f := range currentDbChanges {
		if _, isChangeOld := previousDbChanges[f.Name()]; !isChangeOld {
			// If there is a new file in the directory then the change
			// has been documented and the test should pass
			return true
		}
	}

	return false
}

func getDatabaseFormatDocumentation(startPath string, allowDescend bool) ([]fs.DirEntry, bool) {
	startInfo, err := os.Stat(startPath)
	if err != nil {
		panic(err)
	}

	var currentDirectory string
	if startInfo.IsDir() {
		currentDirectory = startPath
	} else {
		currentDirectory = path.Dir(startPath)
	}

	for {
		directoryContents, err := os.ReadDir(currentDirectory)
		if err != nil {
			panic(err)
		}

		for _, directoryItem := range directoryContents {
			directoryItemPath := path.Join(currentDirectory, directoryItem.Name())
			if directoryItem.Name() == documentationDirectoryName {
				probableFormatChangeDirectoryContents, err := os.ReadDir(directoryItemPath)
				if err != nil {
					panic(err)
				}
				for _, possibleDocumentationItem := range probableFormatChangeDirectoryContents {
					if path.Ext(possibleDocumentationItem.Name()) == ".md" {
						// If the directory's name matches the expected, and contains .md files
						// we assume it is the documentation directory
						return probableFormatChangeDirectoryContents, true
					}
				}
			} else {
				if directoryItem.IsDir() {
					childContents, directoryFound := getDatabaseFormatDocumentation(directoryItemPath, false)
					if directoryFound {
						return childContents, true
					}
				}
			}
		}

		if allowDescend {
			// If not found in this directory, continue down the path
			currentDirectory = path.Dir(currentDirectory)

			if currentDirectory == "." || currentDirectory == "/" {
				panic("Database documentation directory not found")
			}
		} else {
			return []fs.DirEntry{}, false
		}
	}
}

func getLatestCommit(repoName string, branchName string) string {
	cmd := exec.Command("git", "ls-remote", repoName, "refs/heads/"+branchName)
	result, err := cmd.Output()
	if err != nil {
		panic(err)
	}

	// This is a tab, not a space!
	seperator := "\t"
	return strings.Split(string(result), seperator)[0]
}
