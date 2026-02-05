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
	if checkedValue != keyring.KeyringBackendFile {
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
	if checkedValue != keyring.KeyringBackendSystem {
		t.Fatal("keyring.backend is not set to system")
	}
}

// This test will test the callback_GenerateKeyringFiles function.
// Specifically, it will test that the keyring file is created in the correct directory for
// the node identity key, but none of the other keys.
func Test_GenerateMultipleKeyringFiles_OnlyIdentityKey(t *testing.T) {
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
	requireKeyInKeyring(t, kr, "node-identity-key", "secp256k1", 32, "")

	// Then, check that none of the other keys were generated
	for _, keyname := range []string{"peer-key", "encryption-key", "searchable-encryption-key"} {
		_, err := kr.Get(keyname)
		if err == nil {
			t.Fatalf("expected %s to not exist, but it does", keyname)
		}
	}
}

// This test will test the callback_GenerateKeyringFiles function.
// Specifically, it will test that the keyring files are created in the correct
// directory for all of the key types.
func Test_GenerateMultipleKeyringFiles_AllKeys(t *testing.T) {
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
	openKeyring, err := keyring.OpenFileKeyring(keyringDir, []byte(testSecretValue))
	if err != nil {
		t.Fatalf("failed to reopen keyring: %v", err)
	}

	// Check that each key was generated and stored in the keyring
	requireKeyInKeyring(t, openKeyring, "node-identity-key", "secp256k1", 32, "")
	requireKeyInKeyring(t, openKeyring, "peer-key", "", 64, "")
	requireKeyInKeyring(t, openKeyring, "encryption-key", "", 32, "")
	requireKeyInKeyring(t, openKeyring, "searchable-encryption-key", "", 32, "")
}

// This test will test the callback_GenerateIdentityKey, callback_GeneratePeerKey, callback_GenerateEncryptionKey,
//
//	and callback_GenerateSearchableEncryptionKey functions using the file keyring.
func Test_GenerateIndividualKeyrings_FileKeyring(t *testing.T) {
	testSecretValue := "test-secret"

	// Set up a clean test environment
	unsetEnvForTest(t, "DEFRA_KEYRING_SECRET")
	os.Setenv("DEFRA_KEYRING_SECRET", testSecretValue)
	tmpDir := setupWorkingDirectoryForTest(t)
	keyringDir := tmpDir + "/keys"

	ctx := &WizardContext{
		Results: map[string][]any{
			"stepKeyringStorageLocation": {0},
		},
		RootDir: tmpDir,
	}
	setConfigValueForTest(t, ctx, "keyring.path", keyringDir)

	// Execute the actual functions, then the first check will be that it didn't error
	err := callback_GenerateIdentityKey(nil, ctx)
	if err != nil {
		t.Fatal(err)
	}
	err = callback_GeneratePeerKey(nil, ctx)
	if err != nil {
		t.Fatal(err)
	}
	err = callback_GenerateEncryptionKey(nil, ctx)
	if err != nil {
		t.Fatal(err)
	}
	err = callback_GenerateSearchableEncryptionKey(nil, ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Then, check that the keys were generated and stored in the keyring
	kr, err := keyring.OpenFileKeyring(keyringDir, []byte(testSecretValue))
	if err != nil {
		t.Fatalf("failed to reopen keyring: %v", err)
	}

	requireKeyInKeyring(t, kr, "node-identity-key", "secp256k1", 32, "")
	requireKeyInKeyring(t, kr, "peer-key", "", 64, "")
	requireKeyInKeyring(t, kr, "encryption-key", "", 32, "")
	requireKeyInKeyring(t, kr, "searchable-encryption-key", "", 32, "")
}

// This test will test the callback_GenerateIdentityKey, callback_GeneratePeerKey, callback_GenerateEncryptionKey,
//
//	and callback_GenerateSearchableEncryptionKey functions using the system keyring.
func Test_GenerateIndividualKeyrings_SystemKeyring(t *testing.T) {
	// Skip the test on Linux CI / WSL due to missing dbus-launch
	if runtime.GOOS == "linux" {
		t.Skip("system keyring tests are skipped on Linux CI / WSL due to missing dbus-launch")
	}

	ctx := &WizardContext{
		Results: map[string][]any{
			"stepKeyringStorageLocation": {1},
		},
	}

	// Assign a unique namespace for the test keyring so we can remove it afterwards
	keyringNamespace := fmt.Sprintf("test-system-keyring-%d", time.Now().UnixNano())
	setConfigValueForTest(t, ctx, "keyring.namespace", keyringNamespace)

	// Execute the actual functions, then the first check will be that it didn't error
	err := callback_GenerateIdentityKey(nil, ctx)
	if err != nil {
		t.Fatal(err)
	}
	err = callback_GeneratePeerKey(nil, ctx)
	if err != nil {
		t.Fatal(err)
	}
	err = callback_GenerateEncryptionKey(nil, ctx)
	if err != nil {
		t.Fatal(err)
	}
	err = callback_GenerateSearchableEncryptionKey(nil, ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Open the keyring and check that the keys were generated and stored
	openKeyring := keyring.OpenSystemKeyring(keyringNamespace)
	if err != nil {
		t.Fatalf("failed to reopen keyring: %v", err)
	}

	requireKeyInKeyring(t, openKeyring, "node-identity-key", "secp256k1", 32, "")
	requireKeyInKeyring(t, openKeyring, "peer-key", "", 64, "")
	requireKeyInKeyring(t, openKeyring, "encryption-key", "", 32, "")
	requireKeyInKeyring(t, openKeyring, "searchable-encryption-key", "", 32, "")
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
	requireKeyInKeyring(t, openKeyring, "node-identity-key", "secp256k1", 32, "")

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

	// Check each key separately
	requireKeyInKeyring(t, openKeyring, "node-identity-key", "secp256k1", 32, "")
	requireKeyInKeyring(t, openKeyring, "peer-key", "", 64, "")
	requireKeyInKeyring(t, openKeyring, "encryption-key", "", 32, "")
	requireKeyInKeyring(t, openKeyring, "searchable-encryption-key", "", 32, "")

	// Finally, cleanup the entries in the keyring we made for this test
	_ = openKeyring.Delete("node-identity-key")
	_ = openKeyring.Delete("peer-key")
	_ = openKeyring.Delete("encryption-key")
	_ = openKeyring.Delete("searchable-encryption-key")
}

// This will test the callback_ImportIdentityKey function using the file keyring, and a secp256r1 key.
func Test_ImportIdentityKey_Secp256r1_FileKeyring(t *testing.T) {
	testSecretValue := "test-secret"

	// Set up a clean test environment
	unsetEnvForTest(t, "DEFRA_KEYRING_SECRET")
	os.Setenv("DEFRA_KEYRING_SECRET", testSecretValue)
	tmpDir := setupWorkingDirectoryForTest(t)
	keyringDir := tmpDir + "/keys"

	dummyKey_secp256r1 := "75f22540e27d2f47680982acc22fc7b7976b92cddcf1a7846518482d4f463139"

	ctx := &WizardContext{
		Results: map[string][]any{
			"stepKeyringStorageLocation":        {0},
			"stepQueryImportingIdentityKeyType": {2},
			"stepGettingIdentityKeyForImport":   {dummyKey_secp256r1},
		},
		RootDir: tmpDir,
	}
	setConfigValueForTest(t, ctx, "keyring.path", keyringDir)

	// Execute the callback, then the first check will be that it didn't error
	err := callback_ImportIdentityKey(nil, ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Then, check that the key was imported correctly
	openKeyring, err := keyring.OpenFileKeyring(keyringDir, []byte(testSecretValue))
	if err != nil {
		t.Fatalf("failed to reopen keyring: %v", err)
	}
	requireKeyInKeyring(t, openKeyring, "node-identity-key", "secp256r1", 32, dummyKey_secp256r1)
}

// This will test the callback_ImportIdentityKey function using the system keyring, and a secp256r1 key.
func Test_ImportIdentityKey_Secp256r1_SystemKeyring(t *testing.T) {
	// Skip the test on Linux CI / WSL due to missing dbus-launch
	if runtime.GOOS == "linux" {
		t.Skip("system keyring tests are skipped on Linux CI / WSL due to missing dbus-launch")
	}

	dummyKey_secp256r1 := "75f22540e27d2f47680982acc22fc7b7976b92cddcf1a7846518482d4f463139"

	ctx := &WizardContext{
		Results: map[string][]any{
			"stepKeyringStorageLocation":        {1},
			"stepQueryImportingIdentityKeyType": {2},
			"stepGettingIdentityKeyForImport":   {dummyKey_secp256r1},
		},
	}

	// Assign a unique namespace for the test keyring so we can remove it afterwards
	keyringNamespace := fmt.Sprintf("test-system-keyring-%d", time.Now().UnixNano())
	setConfigValueForTest(t, ctx, "keyring.namespace", keyringNamespace)

	// Execute the callback, then the first check will be that it didn't error
	err := callback_ImportIdentityKey(nil, ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Then, check that the key was imported correctly
	// Open the keyring and check that the keys were generated and stored
	openKeyring := keyring.OpenSystemKeyring(keyringNamespace)
	if err != nil {
		t.Fatalf("failed to reopen keyring: %v", err)
	}
	requireKeyInKeyring(t, openKeyring, "node-identity-key", "secp256r1", 32, dummyKey_secp256r1)
}

// This will test the callback_ImportIdentityKey function using the file keyring, and a secp256k1 key.
func Test_ImportIdentityKey_Secp256k1_FileKeyring(t *testing.T) {
	testSecretValue := "test-secret"

	// Set up a clean test environment
	unsetEnvForTest(t, "DEFRA_KEYRING_SECRET")
	os.Setenv("DEFRA_KEYRING_SECRET", testSecretValue)
	tmpDir := setupWorkingDirectoryForTest(t)
	keyringDir := tmpDir + "/keys"

	dummyKey_secp256k1 := "1cf0c5b2af63ade9020b0f1d38e927ae2f384e1b635e601f18f281e53b981a22"

	ctx := &WizardContext{
		Results: map[string][]any{
			"stepKeyringStorageLocation":        {0},
			"stepQueryImportingIdentityKeyType": {1},
			"stepGettingIdentityKeyForImport":   {dummyKey_secp256k1},
		},
		RootDir: tmpDir,
	}
	setConfigValueForTest(t, ctx, "keyring.path", keyringDir)

	// Execute the callback, then the first check will be that it didn't error
	err := callback_ImportIdentityKey(nil, ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Then, check that the key was imported correctly
	openKeyring, err := keyring.OpenFileKeyring(keyringDir, []byte(testSecretValue))
	if err != nil {
		t.Fatalf("failed to reopen keyring: %v", err)
	}
	requireKeyInKeyring(t, openKeyring, "node-identity-key", "secp256k1", 32, dummyKey_secp256k1)
}

// This will test the callback_ImportIdentityKey function using the system keyring, and a secp256k1 key.
func Test_ImportIdentityKey_Secp256k1_SystemKeyring(t *testing.T) {
	// Skip the test on Linux CI / WSL due to missing dbus-launch
	if runtime.GOOS == "linux" {
		t.Skip("system keyring tests are skipped on Linux CI / WSL due to missing dbus-launch")
	}

	dummyKey_secp256k1 := "1cf0c5b2af63ade9020b0f1d38e927ae2f384e1b635e601f18f281e53b981a22"

	ctx := &WizardContext{
		Results: map[string][]any{
			"stepKeyringStorageLocation":        {1},
			"stepQueryImportingIdentityKeyType": {2},
			"stepGettingIdentityKeyForImport":   {dummyKey_secp256k1},
		},
	}

	// Assign a unique namespace for the test keyring so we can remove it afterwards
	keyringNamespace := fmt.Sprintf("test-system-keyring-%d", time.Now().UnixNano())
	setConfigValueForTest(t, ctx, "keyring.namespace", keyringNamespace)

	// Execute the callback, then the first check will be that it didn't error
	err := callback_ImportIdentityKey(nil, ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Then, check that the key was imported correctly
	// Open the keyring and check that the keys were generated and stored
	openKeyring := keyring.OpenSystemKeyring(keyringNamespace)
	if err != nil {
		t.Fatalf("failed to reopen keyring: %v", err)
	}
	requireKeyInKeyring(t, openKeyring, "node-identity-key", "secp256k1", 32, dummyKey_secp256k1)
}

// This will test the callback_ImportIdentityKey function using the file keyring, and a ed25519 key.
func Test_ImportIdentityKey_Ed25519_FileKeyring(t *testing.T) {
	testSecretValue := "test-secret"

	// Set up a clean test environment
	unsetEnvForTest(t, "DEFRA_KEYRING_SECRET")
	os.Setenv("DEFRA_KEYRING_SECRET", testSecretValue)
	tmpDir := setupWorkingDirectoryForTest(t)
	keyringDir := tmpDir + "/keys"

	dummyKey_ed25519 := "f0a804f0ab5d6bd49c6e55f27b433a8b28d23f1290a930fb1f16f6e433710" +
		"0b5638f8e118d2c2c3d21a0c5e56b78756de96ca96f0ae0e54e7055ea67f93d84c2"

	ctx := &WizardContext{
		Results: map[string][]any{
			"stepKeyringStorageLocation":        {0},
			"stepQueryImportingIdentityKeyType": {0},
			"stepGettingIdentityKeyForImport":   {dummyKey_ed25519},
		},
		RootDir: tmpDir,
	}
	setConfigValueForTest(t, ctx, "keyring.path", keyringDir)

	// Execute the callback, then the first check will be that it didn't error
	err := callback_ImportIdentityKey(nil, ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Then, check that the key was imported correctly
	openKeyring, err := keyring.OpenFileKeyring(keyringDir, []byte(testSecretValue))
	if err != nil {
		t.Fatalf("failed to reopen keyring: %v", err)
	}
	requireKeyInKeyring(t, openKeyring, "node-identity-key", "ed25519", 64, dummyKey_ed25519)
}

// This will test the callback_ImportIdentityKey function using the system keyring, and a ed25519 key.
func Test_ImportIdentityKey_Ed25519_SystemKeyring(t *testing.T) {
	// Skip the test on Linux CI / WSL due to missing dbus-launch
	if runtime.GOOS == "linux" {
		t.Skip("system keyring tests are skipped on Linux CI / WSL due to missing dbus-launch")
	}

	dummyKey_ed25519 := "f0a804f0ab5d6bd49c6e55f27b433a8b28d23f1290a930fb1f16f6e433710" +
		"0b5638f8e118d2c2c3d21a0c5e56b78756de96ca96f0ae0e54e7055ea67f93d84c2"

	ctx := &WizardContext{
		Results: map[string][]any{
			"stepKeyringStorageLocation":        {1},
			"stepQueryImportingIdentityKeyType": {2},
			"stepGettingIdentityKeyForImport":   {dummyKey_ed25519},
		},
	}

	// Assign a unique namespace for the test keyring so we can remove it afterwards
	keyringNamespace := fmt.Sprintf("test-system-keyring-%d", time.Now().UnixNano())
	setConfigValueForTest(t, ctx, "keyring.namespace", keyringNamespace)

	// Execute the callback, then the first check will be that it didn't error
	err := callback_ImportIdentityKey(nil, ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Then, check that the key was imported correctly
	// Open the keyring and check that the keys were generated and stored
	openKeyring := keyring.OpenSystemKeyring(keyringNamespace)
	if err != nil {
		t.Fatalf("failed to reopen keyring: %v", err)
	}
	requireKeyInKeyring(t, openKeyring, "node-identity-key", "ed25519", 64, dummyKey_ed25519)
}

// This will test the callback_ImportPeerKey, callback_ImportEncryptionKey, and
// callback_ImportSearchableEncryptionKey functions using the file keyring.
func Test_ImportMultipleKeys_FileKeyring(t *testing.T) {
	testSecretValue := "test-secret"

	// Set up a clean test environment
	unsetEnvForTest(t, "DEFRA_KEYRING_SECRET")
	os.Setenv("DEFRA_KEYRING_SECRET", testSecretValue)
	tmpDir := setupWorkingDirectoryForTest(t)
	keyringDir := tmpDir + "/keys"

	dummyKey_peer := "ecce81a027d6dc6bddef226aa719719453ef315c5d991860d8f4763df564" +
		"c8bc78b95ea264b812ba1b99e3572d4be0344f6c27876767308df8ead1be0e5659cd"
	dummyKey_encryption := "f9046c13ba264b115a96bba9cc6cbf48129242290eac7bc5d760001f1fb65eac"
	dummyKey_searchableEncryption := "273d5cfd18561f8aee3851327630acbf5bc7c660d2a13998950125cded2bf444"

	ctx := &WizardContext{
		Results: map[string][]any{
			"stepKeyringStorageLocation":                  {0},
			"stepGettingPeerKeyForImport":                 {dummyKey_peer},
			"stepGettingEncryptionKeyForImport":           {dummyKey_encryption},
			"stepGettingSearchableEncryptionKeyForImport": {dummyKey_searchableEncryption},
		},
		RootDir: tmpDir,
	}
	setConfigValueForTest(t, ctx, "keyring.path", keyringDir)

	// Execute the callbacks, then the first check will be that they didn't error
	err := callback_ImportPeerKey(nil, ctx)
	if err != nil {
		t.Fatal(err)
	}
	err = callback_ImportEncryptionKey(nil, ctx)
	if err != nil {
		t.Fatal(err)
	}
	err = callback_ImportSearchableEncryptionKey(nil, ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Then, check that the keys were imported correctly
	openKeyring, err := keyring.OpenFileKeyring(keyringDir, []byte(testSecretValue))
	if err != nil {
		t.Fatalf("failed to reopen keyring: %v", err)
	}
	requireKeyInKeyring(t, openKeyring, "peer-key", "", 64, dummyKey_peer)
	requireKeyInKeyring(t, openKeyring, "encryption-key", "", 32, dummyKey_encryption)
	requireKeyInKeyring(t, openKeyring, "searchable-encryption-key", "", 32, dummyKey_searchableEncryption)
}

// This will test the callback_ImportPeerKey, callback_ImportEncryptionKey, and
// callback_ImportSearchableEncryptionKey functions using the system keyring.
func Test_ImportMultipleKeys_SystemKeyring(t *testing.T) {
	// Skip the test on Linux CI / WSL due to missing dbus-launch
	if runtime.GOOS == "linux" {
		t.Skip("system keyring tests are skipped on Linux CI / WSL due to missing dbus-launch")
	}

	testSecretValue := "test-secret"

	// Set up a clean test environment
	unsetEnvForTest(t, "DEFRA_KEYRING_SECRET")
	os.Setenv("DEFRA_KEYRING_SECRET", testSecretValue)
	tmpDir := setupWorkingDirectoryForTest(t)
	keyringDir := tmpDir + "/keys"

	dummyKey_peer := "ecce81a027d6dc6bddef226aa719719453ef315c5d991860d8f4763df564" +
		"c8bc78b95ea264b812ba1b99e3572d4be0344f6c27876767308df8ead1be0e5659cd"
	dummyKey_encryption := "f9046c13ba264b115a96bba9cc6cbf48129242290eac7bc5d760001f1fb65eac"
	dummyKey_searchableEncryption := "273d5cfd18561f8aee3851327630acbf5bc7c660d2a13998950125cded2bf444"

	ctx := &WizardContext{
		Results: map[string][]any{
			"stepKeyringStorageLocation":                  {0},
			"stepGettingPeerKeyForImport":                 {dummyKey_peer},
			"stepGettingEncryptionKeyForImport":           {dummyKey_encryption},
			"stepGettingSearchableEncryptionKeyForImport": {dummyKey_searchableEncryption},
		},
		RootDir: tmpDir,
	}
	setConfigValueForTest(t, ctx, "keyring.path", keyringDir)

	// Assign a unique namespace for the test keyring so we can remove it afterwards
	keyringNamespace := fmt.Sprintf("test-system-keyring-%d", time.Now().UnixNano())
	setConfigValueForTest(t, ctx, "keyring.namespace", keyringNamespace)

	// Execute the callbacks, then the first check will be that they didn't error
	err := callback_ImportPeerKey(nil, ctx)
	if err != nil {
		t.Fatal(err)
	}
	err = callback_ImportEncryptionKey(nil, ctx)
	if err != nil {
		t.Fatal(err)
	}
	err = callback_ImportSearchableEncryptionKey(nil, ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Then, check that the keys were imported correctly
	openKeyring := keyring.OpenSystemKeyring(keyringNamespace)
	if err != nil {
		t.Fatalf("failed to reopen keyring: %v", err)
	}
	requireKeyInKeyring(t, openKeyring, "peer-key", "", 64, dummyKey_peer)
	requireKeyInKeyring(t, openKeyring, "encryption-key", "", 32, dummyKey_encryption)
	requireKeyInKeyring(t, openKeyring, "searchable-encryption-key", "", 32, dummyKey_searchableEncryption)
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
