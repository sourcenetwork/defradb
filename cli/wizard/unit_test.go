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
	"fmt"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/sourcenetwork/defradb/keyring"
)

// This will test the callback_GenerateConfigYAMLFile function.
// Specifically, it will test that the config.yaml file is created in the correct directory,
// and that it is not empty.
func Test_GenerateConfigYAMLFile(t *testing.T) {
	// Set up a clean test environment
	tmpDir := setupWorkingDirectoryForTest(t)

	ctx := &WizardContext{
		RootDir: tmpDir,
	}

	// Execute the actual callback, then the first check will be that it didn't error
	err := callback_GenerateConfigYAMLFile(nil, ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Then, check that the config.yaml file was created, and is not-empty
	info, err := os.Stat(tmpDir + "/config.yaml")
	if err != nil {
		t.Fatalf("expected config.yaml file to exist: %v", err)
	}
	if info.Size() == 0 {
		t.Fatalf("expected %s to be a non-empty file", tmpDir+"/config.yaml")
	}
}

// This will test the callback_SetKeyringBackend function.
// Specifically, it will test that the keyring backend is set to "file" or "system"
// depending on the user's choice (spoofed in the context).
func Test_SetKeyringBackend(t *testing.T) {
	// Set up a clean test environment
	tmpDir := setupWorkingDirectoryForTest(t)

	ctx := &WizardContext{
		RootDir: tmpDir,
	}

	// Spoof the  model steps for this test
	stepCursor0 := &modelMultipleChoice{cursor: 0} // File
	stepCursor1 := &modelMultipleChoice{cursor: 1} // System

	// Test the callback function with choice 0 (File)
	err := callback_SetKeyringBackend(stepCursor0, ctx)
	if err != nil {
		t.Fatal(err)
	}
	checkedValue, ok := getConfigValue(ctx, "keyring.backend").(string)
	if !ok {
		t.Fatal("failed to type assert keyring.backend value")
	}
	if checkedValue != "file" {
		t.Fatal("keyring.backend is not set to file")
	}

	// Test the callback function with choice 1 (System)
	err = callback_SetKeyringBackend(stepCursor1, ctx)
	if err != nil {
		t.Fatal(err)
	}
	checkedValue, ok = getConfigValue(ctx, "keyring.backend").(string)
	if !ok {
		t.Fatal("failed to type assert keyring.backend value")
	}
	if checkedValue != "system" {
		t.Fatal("keyring.backend is not set to system")
	}
}

// This test will test the callback_GenerateKeyringFiles function.
// Specifically, it will test that the keyring file is created in the correct directory for
// the node identity key, but none of the other keys.
func Test_GenerateKeyringFiles_OnlyIdentityKey(t *testing.T) {
	testSecretValue := "test-secret"

	// Set up a clean test environment
	unsetEnvForTest(t, "DEFRA_KEYRING_SECRET")
	os.Setenv("DEFRA_KEYRING_SECRET", testSecretValue)
	tmpDir := setupWorkingDirectoryForTest(t)
	keyringDir := tmpDir + "/keys"

	ctx := &WizardContext{
		Results: map[string][]any{
			"stepSelectKeyTypes": {[]bool{false, false, false}},
		},
		RootDir: tmpDir,
	}
	setConfigValueForTest(t, ctx, "keyring.path", keyringDir)

	// Execute the actual function, then the first check will be that it didn't error
	err := callback_GenerateKeyringFiles(nil, ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Next, check that the keyring directory was created by the callback
	info, err := os.Stat(keyringDir)
	if err != nil {
		t.Fatalf("expected keyring directory to exist: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("expected %s to be a directory", keyringDir)
	}

	// Then, check that the node identity key was generated and stored in the keyring
	kr, err := keyring.OpenFileKeyring(keyringDir, []byte(testSecretValue))
	if err != nil {
		t.Fatalf("failed to reopen keyring: %v", err)
	}
	val, err := kr.Get("node-identity-key")
	if err != nil {
		t.Fatalf("expected encryption-key to exist: %v", err)
	}
	if len(val) != 32 {
		t.Fatalf("expected 32-byte AES-256 key, got %d bytes", len(val))
	}

	// Then, check that none of the other keys were generated
	for _, keyname := range []string{"peer-key", "encryption-key", "searchable-encryption-key"} {
		_, err := kr.Get(keyname)
		if err == nil {
			t.Fatalf("expected %s to not exist, but it does", keyname)
		}
	}

	// Finally, cleanup the entry in the keyring we made for this test
	_ = kr.Delete("node-identity-key")
}

// This test will test the callback_GenerateKeyringFiles function.
// Specifically, it will test that the keyring files are created in the correct
// directory for all of the key types.
func Test_GenerateKeyringFiles_AllKeys(t *testing.T) {
	testSecretValue := "test-secret"

	// Set up a clean test environment
	unsetEnvForTest(t, "DEFRA_KEYRING_SECRET")
	os.Setenv("DEFRA_KEYRING_SECRET", testSecretValue)
	tmpDir := setupWorkingDirectoryForTest(t)
	keyringDir := tmpDir + "/keys"

	ctx := &WizardContext{
		Results: map[string][]any{
			"stepSelectKeyTypes": {[]bool{true, true, true}},
		},
		RootDir: tmpDir,
	}
	setConfigValueForTest(t, ctx, "keyring.path", keyringDir)

	// Execute the actual function, then the first check will be that it didn't error
	err := callback_GenerateKeyringFiles(nil, ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Next, check that the keyring directory was created by the callback
	info, err := os.Stat(keyringDir)
	if err != nil {
		t.Fatalf("expected keyring directory to exist: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("expected %s to be a directory", keyringDir)
	}

	// Open the keyring
	kr, err := keyring.OpenFileKeyring(keyringDir, []byte(testSecretValue))
	if err != nil {
		t.Fatalf("failed to reopen keyring: %v", err)
	}

	// Then, check that all of the keys were generated and stored in the keyring
	keysToCheck := []string{
		"node-identity-key",
		"peer-key",
		"encryption-key",
		"searchable-encryption-key",
	}

	for _, keyname := range keysToCheck {
		val, err := kr.Get(keyname)
		if err != nil {
			t.Fatalf("expected %s to exist: %v", keyname, err)
		}
		if len(val) != 32 {
			t.Fatalf("expected %s to be 32-byte AES-256 key, got %d bytes", keyname, len(val))
		}
	}

	// Finally, cleanup the entries in the keyring we made for this test
	_ = kr.Delete("node-identity-key")
	_ = kr.Delete("peer-key")
	_ = kr.Delete("encryption-key")
	_ = kr.Delete("searchable-encryption-key")
}

// This will test the callback_GenerateKeysInSystemKeyring function.
// Specifically, it will test that the node identity key generated and stored in the system
// keyring, and that none of the other keys are generated.
// Note that this test will not work on WSL.
func Test_GenerateKeysInSystemKeyring_OnlyIdentityKey(t *testing.T) {
	// Skip the test on Linux CI / WSL due to missing dbus-launch
	if runtime.GOOS == "linux" {
		t.Skip("system keyring tests are skipped on Linux CI / WSL due to missing dbus-launch")
	}

	// Set up a clean test environment, and create a context for the test
	tmpDir := setupWorkingDirectoryForTest(t)
	ctx := &WizardContext{
		Results: map[string][]any{
			"stepSelectKeyTypes": {[]bool{false, false, false}},
		},
		RootDir: tmpDir,
	}

	// Assign a unique namespace for the test keyring so we can remove it afterwards
	keyringNamespace := fmt.Sprintf("test-system-keyring-%d", time.Now().UnixNano())
	setConfigValueForTest(t, ctx, "keyring.namespace", keyringNamespace)

	// Execute the callback, then the first check will be that it didn't error
	err := callback_GenerateKeysInSystemKeyring(nil, ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Open the keyring and check that the node identity key was generated and stored
	openKeyring := keyring.OpenSystemKeyring(keyringNamespace)
	val, err := openKeyring.Get("node-identity-key")
	if err != nil {
		t.Fatalf("expected encryption-key to exist: %v", err)
	}
	if len(val) != 32 {
		t.Fatalf("expected 32-byte AES-256 key, got %d bytes", len(val))
	}

	// Then, check that none of the other keys were generated
	for _, keyname := range []string{"peer-key", "encryption-key", "searchable-encryption-key"} {
		_, err := openKeyring.Get(keyname)
		if err == nil {
			t.Fatalf("expected %s to not exist, but it does", keyname)
		}
	}

	// Finally, cleanup the entry in the keyring we made for this test
	_ = openKeyring.Delete("node-identity-key")
}

// This will test the callback_GenerateKeysInSystemKeyring function.
// Specifically, it will test that all of the keys are generated and stored.
// Note that this test will not work on WSL.
func Test_GenerateKeysInSystemKeyring_AllKeys(t *testing.T) {
	// Skip the test on Linux CI / WSL due to missing dbus-launch
	if runtime.GOOS == "linux" {
		t.Skip("system keyring tests are skipped on Linux CI / WSL due to missing dbus-launch")
	}

	// Set up a clean test environment, and create a context for the test
	tmpDir := setupWorkingDirectoryForTest(t)
	ctx := &WizardContext{
		Results: map[string][]any{
			"stepSelectKeyTypes": {[]bool{true, true, true}},
		},
		RootDir: tmpDir,
	}

	// Assign a unique namespace for the test keyring so we can remove it afterwards
	keyringNamespace := fmt.Sprintf("test-system-keyring-%d", time.Now().UnixNano())
	setConfigValueForTest(t, ctx, "keyring.namespace", keyringNamespace)

	// Execute the callback, then the first check will be that it didn't error
	err := callback_GenerateKeysInSystemKeyring(nil, ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Open the keyring
	openKeyring := keyring.OpenSystemKeyring(keyringNamespace)

	// Then, check that all of the keys were generated and stored in the keyring
	keysToCheck := []string{
		"node-identity-key",
		"peer-key",
		"encryption-key",
		"searchable-encryption-key",
	}

	for _, keyname := range keysToCheck {
		val, err := openKeyring.Get(keyname)
		if err != nil {
			t.Fatalf("expected %s to exist: %v", keyname, err)
		}
		if len(val) != 32 {
			t.Fatalf("expected %s to be 32-byte AES-256 key, got %d bytes", keyname, len(val))
		}
	}

	// Finally, cleanup the entries in the keyring we made for this test
	_ = openKeyring.Delete("node-identity-key")
	_ = openKeyring.Delete("peer-key")
	_ = openKeyring.Delete("encryption-key")
	_ = openKeyring.Delete("searchable-encryption-key")
}

// This will test the callback_SetAndReloadDefraKeyringSecretEnvironmentVariable function.
// Specifically, it will test that the DEFRA_KEYRING_SECRET environment variable can correctly
// be inserted into an .env file, then that .env file can be loaded into the environment variables.
func Test_SetAndReloadDefraKeyringSecretEnvironmentVariable(t *testing.T) {
	testSecretValue := "new-secret-value"

	// Set up a clean test environment
	unsetEnvForTest(t, "DEFRA_KEYRING_SECRET")
	tmpDir := setupWorkingDirectoryForTest(t)

	// Spoof the wizard context to contain the secret value from a previous step, as well as
	// the root directory of the test environment
	ctx := &WizardContext{
		Results: map[string][]any{
			"stepGetDefraKeyringSecretInput": {testSecretValue},
		},
		RootDir: tmpDir,
	}
	setConfigValueForTest(t, ctx, "secretfile", tmpDir+"/.env")

	// Execute the callback, then the first check will be that it didn't error
	err := callback_SetAndReloadDefraKeyringSecretEnvironmentVariable(nil, ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Then, check that the environment variable was set correctly
	secretValue, ok := os.LookupEnv("DEFRA_KEYRING_SECRET")
	if !ok {
		t.Fatal("DEFRA_KEYRING_SECRET environment variable was not set")
	}
	if secretValue != testSecretValue {
		t.Fatal("DEFRA_KEYRING_SECRET environment variable was not set to the correct value")
	}
}

// This will test the evaluator_IsEnvironmentVariableDefraKeyringSecretSet function.
// Specifically, it will test that it returns 0 or 1, correctly, depending on whether or not
// the DEFRA_KEYRING_SECRET environment variable is set.
func Test_IsEnvironmentVariableDefraKeyringSecretSet(t *testing.T) {
	testSecretValue := "test-secret"
	unsetEnvForTest(t, "DEFRA_KEYRING_SECRET")

	// Test that the result is 0, because the environment variable should not be set
	result, err := evaluator_IsEnvironmentVariableDefraKeyringSecretSet(&WizardContext{})
	if err != nil {
		t.Fatal(err)
	}
	if result != 0 {
		t.Fatal("expected result to be 0")
	}

	// Test that the result is 1, because the environment variable should be set
	os.Setenv("DEFRA_KEYRING_SECRET", testSecretValue)
	result, err = evaluator_IsEnvironmentVariableDefraKeyringSecretSet(&WizardContext{})
	if err != nil {
		t.Fatal(err)
	}
	if result != 1 {
		t.Fatal("expected result to be 1")
	}
}
