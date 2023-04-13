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
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"testing"

	"github.com/sourcenetwork/defradb/client"
)

func IsDetectingDbChanges() bool {
	return DetectDbChanges
}

// Returns true if test should pass early
func DetectDbChangesPreTestChecks(
	t *testing.T,
	collectionNames []string,
) bool {
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

	return false
}

func detectDbChangesInit(repository string, targetBranch string) {
	badgerFile = true
	badgerInMemory = false

	if SetupOnly {
		// Only the primary test process should perform the setup below
		return
	}

	tempDir := os.TempDir()

	latestTargetCommitHash := getLatestCommit(repository, targetBranch)
	detectDbChangesCodeDir = path.Join(tempDir, "defra", latestTargetCommitHash, "code")

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
}

func SetupDatabaseUsingTargetBranch(
	ctx context.Context,
	t *testing.T,
	collectionNames []string,
) client.DB {
	currentTestPackage, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	targetTestPackage := detectDbChangesCodeDir + "/tests/integration/" + strings.Split(
		currentTestPackage,
		"/tests/integration/",
	)[1]

	// If we are checking for database changes, and we are not seting up the database,
	// then we must be in the main test process, and need to create a new process
	// setting up the database for this test using the old branch We should not setup
	// the database using the current branch/process
	goTestCmd := exec.Command(
		"go",
		"test",
		"./...",
		"--run",
		fmt.Sprintf("^%s$", t.Name()),
		"-v",
	)

	path := t.TempDir()

	goTestCmd.Dir = targetTestPackage
	goTestCmd.Env = os.Environ()
	goTestCmd.Env = append(
		goTestCmd.Env,
		setupOnlyEnvName+"=true",
		fileBadgerPathEnvName+"="+path,
	)
	out, err := goTestCmd.Output()

	if err != nil {
		// If file is not found - this must be a new test and
		// doesn't exist in the target branch, so we pass it
		// because the child process tries to run the test, but
		// if it doesnt find it, the parent test should pass (not panic).
		if strings.Contains(err.Error(), ": no such file or directory") {
			t.SkipNow()
		} else {
			// Only log the output if there is an error different from above,
			// logging child test runs confuses the go test runner making it
			// think there are no tests in the parent run (it will still
			// run everything though)!
			log.ErrorE(ctx, string(out), err)
			panic(err)
		}
	}

	refreshedDb, err := newBadgerFileDB(ctx, t, path)
	if err != nil {
		panic(err)
	}

	_, err = refreshedDb.GetCollectionByName(ctx, collectionNames[0])
	if err != nil {
		if err.Error() == "datastore: key not found" {
			// If collection is not found - this must be a new test and
			// doesn't exist in the target branch, so we pass it
			t.SkipNow()
		} else {
			panic(err)
		}
	}
	return refreshedDb
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
