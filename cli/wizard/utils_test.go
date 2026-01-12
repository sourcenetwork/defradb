// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package wizard

import (
	"bytes"
	"encoding/hex"
	"os"
	"testing"

	"github.com/sourcenetwork/defradb/keyring"
)

// unsetEnvForTest is a helper function that unsets an environment variable for the
// duration of a test, but which will restore the original value after the test.
func unsetEnvForTest(t *testing.T, key string) {
	t.Helper()
	originalValue, existed := os.LookupEnv(key)
	t.Cleanup(func() {
		if existed {
			_ = os.Setenv(key, originalValue)
		} else {
			_ = os.Unsetenv(key)
		}
	})
	_ = os.Unsetenv(key)
}

// setConfigValueForTest is a helper that unsets a value from the wizard's config.yaml file,
// but which will restore the original value after the test.
func setConfigValueForTest(t *testing.T, ctx *WizardContext, key string, value any) {
	originalValue, ok := getConfigValue(ctx, key).(string)
	if !ok {
		t.Fatal("failed to get original value")
	}
	t.Cleanup(func() {
		_ = setConfigValue(ctx, key, originalValue)
	})
	_ = setConfigValue(ctx, key, value)
}

// setupWorkingDirectoryForTest is a helper that temporarily changes the working directory to a
// temporary one for use by the test. Current working directory will be restored after the test.
// It will return the temporary directory that was created.
func setupWorkingDirectoryForTest(t *testing.T) string {
	tmpDir := t.TempDir()
	origWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(origWD)
	})
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	return tmpDir
}

// requireKeyInKeyring is a helper that will check that a key exists in the keyring with a given prefix and length.
// It will fail the test if the key does not exist, or is in the incorrect format. If the expected key
// value is provided. It will also check that the key value is correct.
func requireKeyInKeyring(
	t *testing.T,
	kr keyring.Keyring,
	keyName string,
	expectedKeyType string,
	expectedLength int,
	expectedKeyValue string,
) {
	t.Helper()
	val, err := kr.Get(keyName)
	if err != nil {
		t.Fatalf("expected key %q to exist: %v", keyName, err)
	}
	prefix := ""
	if expectedKeyType != "" {
		prefix = expectedKeyType + ":"
	}
	if !bytes.HasPrefix(val, []byte(prefix)) {
		t.Fatalf("expected key prefix %q, got %q", prefix, val)
	}
	raw := val[len(prefix):]
	if len(raw) != expectedLength {
		t.Fatalf("expected %d-byte %s private key, got %d bytes", expectedLength, expectedKeyType, len(raw))
	}
	if expectedKeyValue != "" {
		if hex.EncodeToString(raw) != expectedKeyValue {
			t.Fatalf("expected key value %q, got %q", expectedKeyValue, hex.EncodeToString(raw))
		}
	}
}
