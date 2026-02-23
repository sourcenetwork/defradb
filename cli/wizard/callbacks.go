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
	"context"
	"encoding/hex"
	"net/http"
	"os"
	"os/exec"
	"time"

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

	choice := keyring.KeyringBackendFile
	if mm.cursor == 1 {
		choice = keyring.KeyringBackendSystem
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

// This callback will attempt to add an existing identity key
func callback_AddIdentityKey(_ step, ctx *WizardContext) error {
	// Open either the file or system keyring
	openKeyring, err := getFileOrSystemKeyring(ctx)
	if err != nil {
		return err
	}

	// Retrieve the key value from the previous step
	if len(ctx.Results[stepGettingIdentityKeyForAddID]) == 0 {
		return NewErrNoResultValue(stepGettingIdentityKeyForAddID)
	}
	keyStr, ok := ctx.Results[stepGettingIdentityKeyForAddID][0].(string)
	if !ok {
		return NewErrAssertTypeFailed(ctx.Results[stepGettingIdentityKeyForAddID][0], "string")
	}

	// Decode the pasted hex string into raw bytes
	keyBytes, err := hex.DecodeString(keyStr)
	if err != nil {
		return NewErrInvalidHexKey(err)
	}

	// Determine the key type from a previous step
	keyTypeStep := stepQueryAddingIdentityKeyTypeID
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

// This callback will attempt to add an existing peer key
func callback_AddPeerKey(_ step, ctx *WizardContext) error {
	// Open either the file or system keyring pending on the user's previous selection
	openKeyring, err := getFileOrSystemKeyring(ctx)
	if err != nil {
		return err
	}

	// Try to retrieve the key value from a previous step
	if len(ctx.Results[stepGettingPeerKeyForAddID]) == 0 {
		return NewErrNoResultValue(stepGettingPeerKeyForAddID)
	}
	keyValue, ok := ctx.Results[stepGettingPeerKeyForAddID][0].(string)
	if !ok {
		return NewErrAssertTypeFailed(ctx.Results[stepGettingPeerKeyForAddID][0], "string")
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

	// Add the key into the keyring
	err = openKeyring.Set("peer-key", keyBytes)
	if err != nil {
		return err
	}

	// If we made it this far, we successfully added the key
	return nil
}

// This callback will attempt to add an existing AES-256 encryption key
func callback_AddEncryptionKey(_ step, ctx *WizardContext) error {
	// Open either the file or system keyring
	openKeyring, err := getFileOrSystemKeyring(ctx)
	if err != nil {
		return err
	}

	// Retrieve the key value from the previous step
	keyStep := "stepGettingEncryptionKeyForAdd"
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

// This callback will attempt to add an existing AES-256 searchable encryption key
func callback_AddSearchableEncryptionKey(_ step, ctx *WizardContext) error {
	// Open either the file or system keyring
	openKeyring, err := getFileOrSystemKeyring(ctx)
	if err != nil {
		return err
	}

	// Retrieve the key value from the previous step
	if len(ctx.Results[stepGettingSearchableEncryptionKeyForAddID]) == 0 {
		return NewErrNoResultValue(stepGettingSearchableEncryptionKeyForAddID)
	}
	keyStr, ok := ctx.Results[stepGettingSearchableEncryptionKeyForAddID][0].(string)
	if !ok {
		return NewErrAssertTypeFailed(ctx.Results[stepGettingSearchableEncryptionKeyForAddID][0], "string")
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

// This callback will start a DefraDB instance and perform a health check on it
func callback_PerformHealthcheck(_ step, ctx *WizardContext) error {
	printToTerminal(TerminalClearANSICode)
	printToTerminal("Performing health check...")
	defer printToTerminal(TerminalClearANSICode)

	// Entire health check must finish within a finite amount of time
	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), HealthCheckTimeoutTimeInSeconds*time.Second)
	defer cancel()

	// Resolve the binary path to the one that launched the wizard
	// The reason we do this, is because the purpose of the wizard is to help the user configure
	// DefraDB after installing it. If this function is being called by the wizard, we know
	// that the binary has been built successfully. This function will use that binary to
	// start DefraDB and perform the health check. This ensures that we are testing a specific
	// installation of Defra, rather than testing the behavior of our code in a general sense.
	binPath, err := os.Executable()
	if err != nil {
		return NewErrFailedToResolveBinary(err)
	}

	cmd := exec.CommandContext(
		ctxWithTimeout,
		binPath,
		"start",
	)

	// Capture the output of the command
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output

	// Start DefraDB, checking that it started successfully
	if err := cmd.Start(); err != nil {
		return NewErrFailedToStartDefraDB(err)
	}

	// Defer shutting down defra after the health check
	defer func() {
		if cmd.Process != nil {
			_ = cmd.Process.Signal(os.Interrupt)
			_, _ = cmd.Process.Wait()
		}
	}()

	// Poll the health endpoint
	healthURL := "http://localhost:9181/health-check"
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctxWithTimeout.Done():
			return NewErrFailedToStartDefraDB(errors.New(extractMeaningfulError(output.String())))

		case <-ticker.C:
			resp, err := http.Get(healthURL)
			if err != nil {
				continue // server not up yet
			}

			_ = resp.Body.Close()

			// The health check is successful
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
	}
}
