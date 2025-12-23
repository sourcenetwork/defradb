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
	"errors"
	"os"

	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/cli/config"
	"github.com/sourcenetwork/defradb/crypto"
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
	// Get the DEFRA_KEYRING_SECRET and keyring.path, and open the keyring
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
	return generateKeysInKeyringFromStep(ctx, keyring, "stepSelectKeyTypes")
}

// This callback will generate the keys in the system keyring
func callback_GenerateKeysInSystemKeyring(_ step, ctx *WizardContext) error {
	keyringNamespace, ok := getConfigValue(ctx, "keyring.namespace").(string)
	if !ok {
		return errors.New(errFailedToGetKeyringNamespace)
	}
	keyring := keyring.OpenSystemKeyring(keyringNamespace)
	return generateKeysInKeyringFromStep(ctx, keyring, "stepSelectKeyTypes")
}

// This callback loads the environment variables from the .env file
func callback_SetAndReloadDefraKeyringSecretEnvironmentVariable(_ step, ctx *WizardContext) error {
	stepToRetrieveResultFrom := "stepGetDefraKeyringSecretInput"
	if len(ctx.Results[stepToRetrieveResultFrom]) == 0 {
		return NewErrNoResultValue(stepToRetrieveResultFrom)
	}
	secretValue, ok := ctx.Results[stepToRetrieveResultFrom][0].(string)
	if !ok {
		return NewErrAssertTypeFailed(ctx.Results[stepToRetrieveResultFrom][0], "string")
	}
	err := ensureEnvValue(ctx, "DEFRA_KEYRING_SECRET", secretValue)
	if err != nil {
		return err
	}
	return loadEnvVariablesFromFile(ctx)
}

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

	// Create an anonymous function to generate a key of a specific name
	generateKeyFunction := func(keyName string) error {
		key, err := crypto.GenerateAES256()
		if err != nil {
			return err
		}
		if err := kr.Set(keyName, key); err != nil {
			return err
		}
		return nil
	}

	// Always generate the identity key
	if err := generateKeyFunction("node-identity-key"); err != nil {
		return err
	}

	// Generate the other keys if the user has selected to do so
	for i, keyname := range []string{"peer-key", "encryption-key", "searchable-encryption-key"} {
		if results[i] {
			if err := generateKeyFunction(keyname); err != nil {
				return err
			}
		}
	}

	// If we made it this far, we successfully generated all of the keys
	return nil
}
