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
	errFailedToCloseCollectionVersionQuery string = "failed to close collection version prefix query"
	errFailedToCloseCollectionQuery        string = "failed to close collection prefix query"
)

// NewErrFailedToCloseCollectionVersionQuery returns a new error indicating that the query
// to get a collection version failed to close.
func NewErrFailedToCloseCollectionVersionQuery(inner error) error {
	return errors.Wrap(errFailedToCloseCollectionVersionQuery, inner)
}

// NewErrFailedToCreateCollectionQuery returns a new error indicating that the query
// to create a collection failed to close.
func NewErrFailedToCloseCollectionQuery(inner error) error {
	return errors.Wrap(errFailedToCloseCollectionQuery, inner)
}
