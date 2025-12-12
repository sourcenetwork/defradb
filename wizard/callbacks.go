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

	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/keyring"
)

// This callback will set keyring.backend to either "file" or "system"
func callback_SetKeyringBackend(s step, ctx *WizardContext) {
	mm := s.(*modelMultipleChoice) //nolint:forcetypeassert

	choice := "file"
	if mm.cursor == 1 {
		choice = "system"
	}

	_ = setYAMLValueInFile(getConfigFile(), []string{"keyring", "backend"}, choice)
}

// This callback will generate the config.yaml file
func callback_GenerateConfigYAMLFile(_ step, ctx *WizardContext) {
	_ = ctx.CreateConfigCallback(os.Getenv("HOME") + "/.defradb")
}

// This callback will generate the keyring files
func callback_GenerateKeyringFiles(_ step, ctx *WizardContext) {
	cfgFile := getConfigFile()
	keyringFilepath := getYAMLValueInFile(cfgFile, []string{"keyring", "path"})
	passwordStr, ok := os.LookupEnv("DEFRA_KEYRING_SECRET")
	if !ok {
		return
	}
	fullKeyringFilepath := os.Getenv("HOME") + "/.defradb/keys/" + keyringFilepath.(string) //nolint:forcetypeassert
	keyring, err := keyring.OpenFileKeyring(fullKeyringFilepath, []byte(passwordStr))
	if err != nil {
		return
	}
	encryptionKey, err := crypto.GenerateAES256()
	if err != nil {
		return
	}
	err = keyring.Set("encryption-key", encryptionKey)
	if err != nil {
		return
	}
}
