// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package clock

import (
	cid "github.com/ipfs/go-cid"

	"github.com/sourcenetwork/defradb/errors"
)

const (
	errCreatingBlock          = "error creating block"
	errWritingBlock           = "error writing block"
	errGettingHeads           = "error getting heads"
	errMergingDelta           = "error merging delta"
	errAddingHead             = "error adding head"
	errCheckingHead           = "error checking if is head"
	errReplacingHead          = "error replacing head"
	errCouldNotFindBlock      = "error checking for known block "
	errFailedToGetNextQResult = "failed to get next query result"
	errCouldNotGetEncKey      = "could not get encryption key"
)

var (
	ErrCreatingBlock          = errors.New(errCreatingBlock)
	ErrWritingBlock           = errors.New(errWritingBlock)
	ErrGettingHeads           = errors.New(errGettingHeads)
	ErrMergingDelta           = errors.New(errMergingDelta)
	ErrAddingHead             = errors.New(errAddingHead)
	ErrCheckingHead           = errors.New(errCheckingHead)
	ErrReplacingHead          = errors.New(errReplacingHead)
	ErrCouldNotFindBlock      = errors.New(errCouldNotFindBlock)
	ErrFailedToGetNextQResult = errors.New(errFailedToGetNextQResult)
	ErrDecodingHeight         = errors.New("error decoding height")
	ErrCouldNotGetEncKey      = errors.New(errCouldNotGetEncKey)
)

func NewErrCreatingBlock(inner error) error {
	return errors.Wrap(errCreatingBlock, inner)
}

func NewErrWritingBlock(inner error) error {
	return errors.Wrap(errWritingBlock, inner)
}

func NewErrGettingHeads(inner error) error {
	return errors.Wrap(errGettingHeads, inner)
}

func NewErrMergingDelta(cid cid.Cid, inner error) error {
	return errors.Wrap(errMergingDelta, inner, errors.NewKV("Cid", cid))
}

func NewErrAddingHead(cid cid.Cid, inner error) error {
	return errors.Wrap(errAddingHead, inner, errors.NewKV("Cid", cid))
}

func NewErrCheckingHead(cid cid.Cid, inner error) error {
	return errors.Wrap(errCheckingHead, inner, errors.NewKV("Cid", cid))
}

func NewErrReplacingHead(cid cid.Cid, root cid.Cid, inner error) error {
	return errors.Wrap(
		errReplacingHead,
		inner,
		errors.NewKV("Cid", cid),
		errors.NewKV("Root", root),
	)
}

func NewErrCouldNotFindBlock(cid cid.Cid, inner error) error {
	return errors.Wrap(errCouldNotFindBlock, inner, errors.NewKV("Cid", cid))
}

func NewErrFailedToGetNextQResult(inner error) error {
	return errors.Wrap(errFailedToGetNextQResult, inner)
}
