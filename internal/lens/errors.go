// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package lens

import (
	"github.com/sourcenetwork/defradb/errors"
)

const (
	errStoreLensField   string = "failed to store lens migrated field"
	errStoreLensVersion string = "failed to store lens version key"
)

func NewErrStoreLensField(inner error, fieldName string) error {
	return errors.Wrap(errStoreLensField, inner, errors.NewKV("Field", fieldName))
}

func NewErrStoreLensVersion(inner error) error {
	return errors.Wrap(errStoreLensVersion, inner)
}
