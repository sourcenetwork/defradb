// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package graphql

import (
	"github.com/sourcenetwork/defradb/errors"
)

const errSimilarityOnNonVectorField string = "similarity can only target a numeric array field"

// NewErrSimilarityOnNonVectorField returns an error indicating that SIMILARITY was given a field
// that exists but cannot hold a vector.
func NewErrSimilarityOnNonVectorField(fieldName string, fieldType string) error {
	return errors.New(
		errSimilarityOnNonVectorField,
		errors.NewKV("Field", fieldName),
		errors.NewKV("Type", fieldType),
	)
}
