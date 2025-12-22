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
	"testing"

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

	// Make sure that after the test, the config.yaml file is restored to its original state
	originalConfigValue, ok := getConfigValue(ctx, "keyring.backend").(string)
	if !ok {
		t.Fatal("failed to get original keyring.backend value")
	}
	t.Cleanup(func() {
		setConfigValue(ctx, "keyring.backend", originalConfigValue)
	})

	// Spoof the  model steps for this test
	stepCursor0 := &modelMultipleChoice{cursor: 0} // File
	stepCursor1 := &modelMultipleChoice{cursor: 1} // System

	// Test the callback function with choice 0 (File)
	err := callback_SetKeyringBackend(stepCursor0, &WizardContext{})
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
	err = callback_SetKeyringBackend(stepCursor1, &WizardContext{})
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
// Specifically, it will test that the keyring files are created in the correct directory,
// and that the encryption key is generated and stored in the keyring.
func Test_GenerateKeyringFiles(t *testing.T) {
	testSecretValue := "test-secret"

	// Set up a clean test environment
	unsetEnvForTest(t, "DEFRA_KEYRING_SECRET")
	os.Setenv("DEFRA_KEYRING_SECRET", testSecretValue)
	tmpDir := setupWorkingDirectoryForTest(t)
	keyringDir := tmpDir + "/keys"

	ctx := &WizardContext{
		RootDir: tmpDir,
	}
	setConfigValueForTest(t, ctx, "keyring.path", keyringDir)

	// Execute the actual function, then the first check will be that it didn't error
	err := callback_GenerateKeyringFiles(nil, &WizardContext{})
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

	// Then, check that the encryption key was generated and stored in the keyring
	kr, err := keyring.OpenFileKeyring(keyringDir, []byte(testSecretValue))
	if err != nil {
		t.Fatalf("failed to reopen keyring: %v", err)
	}
	val, err := kr.Get("encryption-key")
	if err != nil {
		t.Fatalf("expected encryption-key to exist: %v", err)
	}
	if len(val) != 32 {
		t.Fatalf("expected 32-byte AES-256 key, got %d bytes", len(val))
	}
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

	// Execute the actual function, then the first check will be that it didn't error
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
		fmt.Println("secretValue", secretValue)
		t.Fatal("DEFRA_KEYRING_SECRET environment variable was not set to the correct value")
	}
}
