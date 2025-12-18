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

	"github.com/sourcenetwork/defradb/errors"
)

const (
	errDefraKeyringSecretNotSet    = "DEFRA_KEYRING_SECRET environment variable is not set"
	errFailedToGetKeyringFilepath  = "failed to get keyring filepath"
	errFailedToGetKeyringNamespace = "failed to get keyring namespace"
	errFailedToGetEnvFilename      = "failed to get env filename"
)

var (
	errModelTypeMismatch           = errors.New("model type mismatch")
	errFailedToRetrieveResultValue = errors.New("failed to retrieve result value from previous step")
)

func NewErrModelTypeMismatch(stepID, expectedType string) error {
	return fmt.Errorf(
		"%w: "+"a type assertion failed when trying to cast step %s to model type %s",
		errModelTypeMismatch,
		stepID,
		expectedType,
	)
}

func NewErrFailedToRetrieveResultValue(stepID string) error {
	return fmt.Errorf(
		"%w: "+"failed to retrieve result value from previous step %s",
		errFailedToRetrieveResultValue,
		stepID,
	)
}
