// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cli

import (
	"bytes"
	"encoding/hex"
	"os"
	"regexp"
	"testing"

	"github.com/sourcenetwork/defradb/crypto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeyringList(t *testing.T) {
	rootdir := t.TempDir()

	err := os.Setenv("DEFRA_KEYRING_SECRET", "password")
	require.NoError(t, err)

	keyNames := []string{"keyname1", "keyname2", "keyname3"}

	// Insert the keys into the keyring
	for _, keyName := range keyNames {
		keyBytes, err := crypto.GenerateAES256()
		require.NoError(t, err)
		keyHex := hex.EncodeToString(keyBytes)
		cmd := NewDefraCommand()
		cmd.SetArgs([]string{"keyring", "import", "--rootdir", rootdir, keyName, keyHex})
		err = cmd.Execute()
		require.NoError(t, err)
	}

	// Run the 'keyring list' command, and require no error on the output
	var output bytes.Buffer
	cmd := NewDefraCommand()
	cmd.SetOut(&output)
	cmd.SetArgs([]string{"keyring", "list", "--rootdir", rootdir})
	err = cmd.Execute()
	require.NoError(t, err)

	outputString := output.String()

	// Use regex to extract the keys, and compare with the expected values
	// We know what the format the output should be, which is:
	// "Keys in the keyring:\n- keyname1\n- keyname2\n- keyname3\n"
	re := regexp.MustCompile(`-\s([^\n]+)`)
	matches := re.FindAllStringSubmatch(outputString, -1)
	var extractedKeys []string
	for _, match := range matches {
		extractedKeys = append(extractedKeys, match[1])
	}

	assert.ElementsMatch(t, keyNames, extractedKeys, "The listed keys do not match the expected keys.")
}
