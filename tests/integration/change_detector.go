package tests

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func DetectDbChangesPreTestChecks(
	t *testing.T,
	collectionNames []string,
) {
	if previousTestCaseTestName == t.Name() {
		// The database format changer currently only supports running the first test
		//  case, if a second case is detected we return early
		t.Skip()
	}
	previousTestCaseTestName = t.Name()

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
}
