// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package description

import "github.com/sourcenetwork/defradb/errors"

const (
	errFailedToCloseSchemaQuery     string = "failed to close schema prefix query"
	errFailedToCloseCollectionQuery string = "failed to close collection prefix query"
)

// NewErrFailedToCreateSchemaQuery returns a new error indicating that the query
// to create a schema failed to close.
func NewErrFailedToCloseSchemaQuery(inner error) error {
	return errors.Wrap(errFailedToCloseSchemaQuery, inner)
}

// NewErrFailedToCreateCollectionQuery returns a new error indicating that the query
// to create a collection failed to close.
func NewErrFailedToCloseCollectionQuery(inner error) error {
	return errors.Wrap(errFailedToCloseCollectionQuery, inner)
}
