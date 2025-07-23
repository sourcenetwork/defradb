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
	errLensRuntimeNotSupported  string = "the selected lens runtime is not supported by this build"
	errStoreTypeNotSupported    string = "the selected store type is not supported by this build"
	errACPTypeNotSupported      string = "the selected acp type is not supported by this build"
	errAdminACPTypeNotSupported string = "the selected admin acp type is not supported by this build"
)

var (
	ErrSignerMissingForSourceHubACP = errors.New("a txn signer must be provided for SourceHub ACP")
	ErrLensRuntimeNotSupported      = errors.New(errLensRuntimeNotSupported)
	ErrStoreTypeNotSupported        = errors.New(errStoreTypeNotSupported)
	ErrPurgeWithDevModeDisabled     = errors.New("cannot purge database when development mode is disabled")
	ErrP2PNotSupported              = errors.New("p2p networking is not supported by this build")
	ErrAdminACPTypeNotSupported     = errors.New(errAdminACPTypeNotSupported)
)

func NewErrLensRuntimeNotSupported(lens LensRuntimeType) error {
	return errors.New(errLensRuntimeNotSupported, errors.NewKV("Lens", lens))
}

func NewErrStoreTypeNotSupported(store StoreType) error {
	return errors.New(errStoreTypeNotSupported, errors.NewKV("Store", store))
}

func NewErrACPTypeNotSupported(acp DocumentACPType) error {
	return errors.New(errACPTypeNotSupported, errors.NewKV("ACP", acp))
}
