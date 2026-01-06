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
)

// This callback will return 0 if DEFRA_KEYRING_SECRET is not set, and 1 if it is
func evaluator_IsEnvironmentVariableDefraKeyringSecretSet(ctx *WizardContext) (int, error) {
	_ = loadEnvVariablesFromFile(ctx)
	val, ok := os.LookupEnv("DEFRA_KEYRING_SECRET")
	if !ok || val == "" {
		return 0, nil
	}
	return 1, nil
}

// This callback will return 0 if the user previously selected to store the keyring
// in the filesystem, and 1 if they selected the OS keychain
func evaluator_ResultOfStepKeyringStorageLocation(ctx *WizardContext) (int, error) {
	valRaw, ok := ctx.Results[stepKeyringStorageLocationID]
	if !ok {
		return -1, NewErrFailedToRetrieveResultValue(stepKeyringStorageLocationID)
	}
	val, ok := valRaw[0].(int)
	if !ok {
		return -1, NewErrAssertTypeFailed(valRaw[0], "int")
	}
	return val, nil
}
