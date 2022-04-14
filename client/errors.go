// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package client

import "errors"

// Errors returnable from this package.
//
// This list is incomplete and undefined errors may also be returned.
// Errors returned from this package may be tested against these errors with errors.Is.
var (
	ErrFieldNotExist       = errors.New("The given field does not exist")
	ErrFieldNotObject      = errors.New("Trying to access field on a non object type")
	ErrValueTypeMismatch   = errors.New("Value does not match indicated type")
	ErrIndexNotFound       = errors.New("No index found for given ID")
	ErrDocumentNotFound    = errors.New("No document for the given key exists")
	ErrInvalidUpdateTarget = errors.New("The target document to update is of invalid type")
	ErrInvalidUpdater      = errors.New("The updater of a document is of invalid type")
	ErrInvalidDeleteTarget = errors.New("The target document to delete is of invalid type")
)
