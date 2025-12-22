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
	encryptionKey, err := crypto.GenerateAES256()
	if err != nil {
		return err
	}
	err = keyring.Set("encryption-key", encryptionKey)
	if err != nil {
		return err
	}
	return nil
}

// This callback will generate the keys in the system keyring
func callback_GenerateKeysInSystemKeyring(_ step, ctx *WizardContext) error {
	keyringNamespace, ok := getConfigValue(ctx, "keyring.namespace").(string)
	if !ok {
		return errors.New(errFailedToGetKeyringNamespace)
	}
	keyring := keyring.OpenSystemKeyring(keyringNamespace)
	encryptionKey, err := crypto.GenerateAES256()
	if err != nil {
		return err
	}
	err = keyring.Set("encryption-key", encryptionKey)
	if err != nil {
		return err
	}
	return nil
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
	envFilename, ok := getConfigValue(ctx, "secretfile").(string)
	if !ok {
		return errors.New(errFailedToGetEnvFilename)
	}
	if envFilename == "" {
		envFilename = DefaultEnvFilename
	}
	err := ensureEnvValue(ctx, "DEFRA_KEYRING_SECRET", secretValue)
	if err != nil {
		return err
	}
	return loadEnvVariablesFromFile(ctx)
}
