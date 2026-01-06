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
	"encoding/hex"
	"os"

	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/cli/config"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/keyring"
)

// This callback will set keyring.backend to either "file" or "system"
func callback_SetKeyringBackend(s step, ctx *WizardContext) error {
	mm, ok := s.(*modelMultipleChoice)
	if !ok {
		return NewErrModelTypeMismatch(s.ID(), "*modelMultipleChoice")
	}

	choice := "file"
	if mm.cursor == 1 {
		choice = "system"
	}

	return setConfigValue(ctx, "keyring.backend", choice)
}

// This callback will generate the config.yaml file
func callback_GenerateConfigYAMLFile(_ step, ctx *WizardContext) error {
	defaultCmd := &cobra.Command{}
	return config.CreateConfig(ctx.RootDir, defaultCmd.Flags())
}

// This callback will generate the keyring files
func callback_GenerateKeyringFiles(_ step, ctx *WizardContext) error {
	passwordStr, ok := os.LookupEnv("DEFRA_KEYRING_SECRET")
	if !ok {
		return errors.New(errDefraKeyringSecretNotSet)
	}
	keyringFilepath, ok := getConfigValue(ctx, "keyring.path").(string)
	if !ok {
		return errors.New(errFailedToGetKeyringFilepath)
	}
	if err := os.MkdirAll(keyringFilepath, 0755); err != nil {
		return err
	}
	keyring, err := keyring.OpenFileKeyring(keyringFilepath, []byte(passwordStr))
	if err != nil {
		return err
	}
	return generateKeysInKeyringFromStep(ctx, keyring, stepSelectKeyTypesID)
}

// This callback will generate the keys in the system keyring
func callback_GenerateKeysInSystemKeyring(_ step, ctx *WizardContext) error {
	keyringNamespace, ok := getConfigValue(ctx, "keyring.namespace").(string)
	if !ok {
		return errors.New(errFailedToGetKeyringNamespace)
	}
	keyring := keyring.OpenSystemKeyring(keyringNamespace)
	return generateKeysInKeyringFromStep(ctx, keyring, stepSelectKeyTypesID)
}

// This callback loads the environment variables from the .env file
func callback_SetAndReloadDefraKeyringSecretEnvironmentVariable(_ step, ctx *WizardContext) error {
	if len(ctx.Results[stepGetDefraKeyringSecretInputID]) == 0 {
		return NewErrNoResultValue(stepGetDefraKeyringSecretInputID)
	}
	secretValue, ok := ctx.Results[stepGetDefraKeyringSecretInputID][0].(string)
	if !ok {
		return NewErrAssertTypeFailed(ctx.Results[stepGetDefraKeyringSecretInputID][0], "string")
	}
	err := ensureEnvValue(ctx, "DEFRA_KEYRING_SECRET", secretValue)
	if err != nil {
		return err
	}
	return loadEnvVariablesFromFile(ctx)
}

// This callback will generate a new identity key
func callback_GenerateIdentityKey(_ step, ctx *WizardContext) error {
	// Open either the file or system keyring pending on the user's previous selection
	openKeyring, err := getFileOrSystemKeyring(ctx)
	if err != nil {
		return err
	}

	// Generate the identity key
	privateKey, err := crypto.GenerateKey(crypto.KeyTypeSecp256k1)
	if err != nil {
		return err
	}
	nodeIdentity, err := identity.FromPrivateKey(privateKey)
	if err != nil {
		return err
	}
	rawKey := nodeIdentity.PrivateKey().Raw()
	identityBytes := append([]byte("secp256k1:"), rawKey...)
	if err := openKeyring.Set("node-identity-key", identityBytes); err != nil {
		return err
	}
	return nil
}

// This callback will attempt to import an existing identity key
func callback_ImportIdentityKey(_ step, ctx *WizardContext) error {
	// Open either the file or system keyring
	openKeyring, err := getFileOrSystemKeyring(ctx)
	if err != nil {
		return err
	}

	// Retrieve the key value from the previous step
	if len(ctx.Results[stepGettingIdentityKeyForImportID]) == 0 {
		return NewErrNoResultValue(stepGettingIdentityKeyForImportID)
	}
	keyStr, ok := ctx.Results[stepGettingIdentityKeyForImportID][0].(string)
	if !ok {
		return NewErrAssertTypeFailed(ctx.Results[stepGettingIdentityKeyForImportID][0], "string")
	}

	// Decode the pasted hex string into raw bytes
	keyBytes, err := hex.DecodeString(keyStr)
	if err != nil {
		return NewErrInvalidHexKey(err)
	}

	// Determine the key type from a previous step
	keyTypeStep := stepQueryImportingIdentityKeyTypeID
	if len(ctx.Results[keyTypeStep]) == 0 {
		return NewErrNoResultValue(keyTypeStep)
	}
	keyTypeRaw, ok := ctx.Results[keyTypeStep][0].(int)
	if !ok {
		return NewErrAssertTypeFailed(ctx.Results[keyTypeStep][0], "int")
	}
	var keyType crypto.KeyType
	switch keyTypeRaw {
	case 0:
		keyType = crypto.KeyTypeEd25519
	case 1:
		keyType = crypto.KeyTypeSecp256k1
	case 2:
		keyType = crypto.KeyTypeSecp256r1
	}

	identityBytes := append([]byte(keyType+":"), keyBytes...)
	if err := openKeyring.Set("node-identity-key", identityBytes); err != nil {
		return err
	}

	return nil
}

// This callback will attempt to import an existing peer key
func callback_ImportPeerKey(_ step, ctx *WizardContext) error {
	// Open either the file or system keyring pending on the user's previous selection
	openKeyring, err := getFileOrSystemKeyring(ctx)
	if err != nil {
		return err
	}

	// Try to retrieve the key value from a previous step
	if len(ctx.Results[stepGettingPeerKeyForImportID]) == 0 {
		return NewErrNoResultValue(stepGettingPeerKeyForImportID)
	}
	keyValue, ok := ctx.Results[stepGettingPeerKeyForImportID][0].(string)
	if !ok {
		return NewErrAssertTypeFailed(ctx.Results[stepGettingPeerKeyForImportID][0], "string")
	}

	// Decode the hex string into raw bytes
	keyBytes, err := hex.DecodeString(keyValue)
	if err != nil {
		return NewErrInvalidHexKey(err)
	}

	// Sanity check length: ed25519 private keys are 64 bytes (or sometimes 96)
	if len(keyBytes) != 64 && len(keyBytes) != 96 {
		return NewErrInvalidEd25519KeyLength(len(keyBytes))
	}

	// Import the key into the keyring
	err = openKeyring.Set("peer-key", keyBytes)
	if err != nil {
		return err
	}

	// If we made it this far, we successfully imported the key
	return nil
}

// This callback will attempt to import an existing AES-256 encryption key
func callback_ImportEncryptionKey(_ step, ctx *WizardContext) error {
	// Open either the file or system keyring
	openKeyring, err := getFileOrSystemKeyring(ctx)
	if err != nil {
		return err
	}

	// Retrieve the key value from the previous step
	keyStep := "stepGettingEncryptionKeyForImport"
	if len(ctx.Results[keyStep]) == 0 {
		return NewErrNoResultValue(keyStep)
	}
	keyStr, ok := ctx.Results[keyStep][0].(string)
	if !ok {
		return NewErrAssertTypeFailed(ctx.Results[keyStep][0], "string")
	}

	// Decode the hex string into raw bytes
	keyBytes, err := hex.DecodeString(keyStr)
	if err != nil {
		return NewErrInvalidHexKey(err)
	}

	// Validate AES-256 key length
	if len(keyBytes) != 32 {
		return NewErrInvalidAES256KeyLength(len(keyBytes))
	}

	if err := openKeyring.Set("encryption-key", keyBytes); err != nil {
		return err
	}

	return nil
}

// This callback will attempt to import an existing AES-256 searchable encryption key
func callback_ImportSearchableEncryptionKey(_ step, ctx *WizardContext) error {
	// Open either the file or system keyring
	openKeyring, err := getFileOrSystemKeyring(ctx)
	if err != nil {
		return err
	}

	// Retrieve the key value from the previous step
	if len(ctx.Results[stepGettingSearchableEncryptionKeyForImportID]) == 0 {
		return NewErrNoResultValue(stepGettingSearchableEncryptionKeyForImportID)
	}
	keyStr, ok := ctx.Results[stepGettingSearchableEncryptionKeyForImportID][0].(string)
	if !ok {
		return NewErrAssertTypeFailed(ctx.Results[stepGettingSearchableEncryptionKeyForImportID][0], "string")
	}

	// Decode the hex string into raw bytes
	keyBytes, err := hex.DecodeString(keyStr)
	if err != nil {
		return NewErrInvalidHexKey(err)
	}

	// Validate AES-256 key length
	if len(keyBytes) != 32 {
		return NewErrInvalidAES256KeyLength(len(keyBytes))
	}

	if err := openKeyring.Set("searchable-encryption-key", keyBytes); err != nil {
		return err
	}

	return nil
}

// This callback will generate a new peer key
func callback_GeneratePeerKey(_ step, ctx *WizardContext) error {
	// Open either the file or system keyring pending on the user's previous selection
	openKeyring, err := getFileOrSystemKeyring(ctx)
	if err != nil {
		return err
	}

	// Generate the peer key
	key, err := crypto.GenerateEd25519()
	if err != nil {
		return err
	}
	if err := openKeyring.Set("peer-key", key); err != nil {
		return err
	}
	return nil
}

// This callback will generate a new encryption key
func callback_GenerateEncryptionKey(_ step, ctx *WizardContext) error {
	// Open either the file or system keyring pending on the user's previous selection
	openKeyring, err := getFileOrSystemKeyring(ctx)
	if err != nil {
		return err
	}

	// Generate the encryption key
	key, err := crypto.GenerateAES256()
	if err != nil {
		return err
	}
	if err := openKeyring.Set("encryption-key", key); err != nil {
		return err
	}
	return nil
}

// This callback will generate a new encryption key
func callback_GenerateSearchableEncryptionKey(_ step, ctx *WizardContext) error {
	// Open either the file or system keyring pending on the user's previous selection
	openKeyring, err := getFileOrSystemKeyring(ctx)
	if err != nil {
		return err
	}

	// Generate the encryption key
	key, err := crypto.GenerateAES256()
	if err != nil {
		return err
	}
	if err := openKeyring.Set("searchable-encryption-key", key); err != nil {
		return err
	}
	return nil
}
