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

import "os"

// This callback will return 0 if DEFRA_KEYRING_SECRET is not set, and 1 if it is
func evaluator_IsEnvironmentVariableDefraKeyringSecretSet(ctx *WizardContext) int {
	_ = loadEnvVariablesFromFile(ctx)
	val, ok := os.LookupEnv("DEFRA_KEYRING_SECRET")
	if !ok || val == "" {
		return 0
	}
	return 1
}
