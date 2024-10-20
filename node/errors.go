// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package node

import (
	"github.com/sourcenetwork/defradb/errors"
)

const (
	errLensRuntimeNotSupported string = "the selected lens runtime is not supported by this build"
	errStoreTypeNotSupported   string = "the selected store type is not supported by this build"
)

var (
	ErrSignerMissingForSourceHubACP = errors.New("a txn signer must be provided for SourceHub ACP")
	ErrLensRuntimeNotSupported      = errors.New(errLensRuntimeNotSupported)
	ErrStoreTypeNotSupported        = errors.New(errStoreTypeNotSupported)
	ErrPurgeWithDevModeDisabled     = errors.New("cannot purge database when development mode is disabled")
)

func NewErrLensRuntimeNotSupported(lens LensRuntimeType) error {
	return errors.New(errLensRuntimeNotSupported, errors.NewKV("Lens", lens))
}

func NewErrStoreTypeNotSupported(store StoreType) error {
	return errors.New(errStoreTypeNotSupported, errors.NewKV("Store", store))
}
