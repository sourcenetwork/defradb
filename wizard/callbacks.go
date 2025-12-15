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

	return setYAMLValueInFile(getConfigFile(), []string{"keyring", "backend"}, choice)
}

// This callback will generate the config.yaml file
func callback_GenerateConfigYAMLFile(_ step, ctx *WizardContext) error {
	return ctx.CreateConfigCallback(os.Getenv("HOME") + "/.defradb")
}

// This callback will generate the keyring files
func callback_GenerateKeyringFiles(_ step, ctx *WizardContext) error {
	passwordStr, ok := os.LookupEnv("DEFRA_KEYRING_SECRET")
	if !ok {
		return errors.New(errDefraKeyringSecretNotSet)
	}
	cfgFile := getConfigFile()
	keyringFilepath, ok := getYAMLValueInFile(cfgFile, []string{"keyring", "path"}).(string)
	if !ok {
		return errors.New(errFailedToGetKeyringFilepath)
	}
	fullKeyringFilepath := os.Getenv("HOME") + "/.defradb/" + keyringFilepath
	if err := os.MkdirAll(fullKeyringFilepath, 0755); err != nil {
		return err
	}
	keyring, err := keyring.OpenFileKeyring(fullKeyringFilepath, []byte(passwordStr))
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

// This callback will generate the keys in the system keyrind
func callback_GenerateKeyringFilesInSystemKeyring(_ step, ctx *WizardContext) error {
	cfgFile := getConfigFile()
	keyringNamespace, ok := getYAMLValueInFile(cfgFile, []string{"keyring", "namespace"}).(string)
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
