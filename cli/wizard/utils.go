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
	"os"
	"strings"

	"github.com/sourcenetwork/defradb/errors"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/keyring"
)

// generateKeysInKeyringFromStep is a helper function to generate the keys in the keyring from the results of a step
// It contains behavior that is common to both callback_GenerateKeyringFiles and callback_GenerateKeysInSystemKeyring
func generateKeysInKeyringFromStep(ctx *WizardContext, kr keyring.Keyring, stepname string) error {
	// Get and cast the results of the step to a []bool
	resultsRaw, ok := ctx.Results[stepname]
	if !ok {
		return NewErrFailedToRetrieveResultValue(stepname)
	}
	results, ok := resultsRaw[0].([]bool)
	if !ok {
		return NewErrAssertTypeFailed(resultsRaw[0], "[]bool")
	}

	// Always generate the identity key
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
	if err := kr.Set("node-identity-key", identityBytes); err != nil {
		return err
	}

	// Generate the other keys if the user has selected to do so
	// Peer key
	if results[0] {
		key, err := crypto.GenerateEd25519()
		if err != nil {
			return err
		}
		if err := kr.Set("peer-key", key); err != nil {
			return err
		}
	}
	// Encryption key
	if results[1] {
		key, err := crypto.GenerateAES256()
		if err != nil {
			return err
		}
		if err := kr.Set("encryption-key", key); err != nil {
			return err
		}
	}
	// Searchable encryption key
	if results[2] {
		key, err := crypto.GenerateAES256()
		if err != nil {
			return err
		}
		if err := kr.Set("searchable-encryption-key", key); err != nil {
			return err
		}
	}

	// If we made it this far, we successfully generated all of the keys
	return nil
}

// getFileOrSystemKeyring is a helper function to get the file or system keyring based
// on the user's previous selection
func getFileOrSystemKeyring(ctx *WizardContext) (keyring.Keyring, error) {
	// Get the user's previous selection for the keyring storage location
	keyringStorageLocationStepName := "stepKeyringStorageLocation"
	valRaw, ok := ctx.Results[keyringStorageLocationStepName]
	if !ok {
		return nil, NewErrFailedToRetrieveResultValue(keyringStorageLocationStepName)
	}
	storageValue, ok := valRaw[0].(int)
	if !ok {
		return nil, NewErrAssertTypeFailed(valRaw[0], "int")
	}

	var kr keyring.Keyring
	// Open the file keyring
	if storageValue == 0 {
		passwordStr, ok := os.LookupEnv("DEFRA_KEYRING_SECRET")
		if !ok {
			return nil, errors.New(errDefraKeyringSecretNotSet)
		}
		keyringFilepath, ok := getConfigValue(ctx, "keyring.path").(string)
		if !ok {
			return nil, errors.New(errFailedToGetKeyringFilepath)
		}
		kr, err := keyring.OpenFileKeyring(keyringFilepath, []byte(passwordStr))
		if err != nil {
			return nil, err
		}
		return kr, nil
		// Open the system keyring
	} else {
		keyringNamespace, ok := getConfigValue(ctx, "keyring.namespace").(string)
		if !ok {
			return nil, errors.New(errFailedToGetKeyringNamespace)
		}
		kr = keyring.OpenSystemKeyring(keyringNamespace)
		return kr, nil
	}
}

// printToTerminal is a helper function to print text to the terminal
// This will discard any errors that may occur from writing
func printToTerminal(text string) {
	_, _ = os.Stdout.WriteString(text)
}

// extractMeaningfulError is a helper function to extract the most meaningful
// error from the output of a CLI command process
func extractMeaningfulError(output string) string {
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Walk backwards to find the most relevant error
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])

		if line == "" {
			continue
		}

		// Common CLI error patterns
		if strings.HasPrefix(line, "Error:") {
			return strings.TrimPrefix(line, "Error:")
		}

		if strings.Contains(strings.ToLower(line), "error") {
			return line
		}
	}

	// Fallback: return entire output if no clear error found
	return output
}
