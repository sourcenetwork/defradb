// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package coreblock

import (
	"fmt"

	"github.com/ipfs/go-cid"

	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/errors"
)

const (
	errNodeToBlock                 string = "failed to convert node to block"
	errEncodingBlock               string = "failed to encode block"
	errUnmarshallingBlock          string = "failed to unmarshal block"
	errGeneratingLink              string = "failed to generate link"
	errInvalidBlockEncryptionType  string = "invalid block encryption type"
	errInvalidBlockEncryptionKeyID string = "invalid block encryption key id"
	errCouldNotLoadSignatureBlock  string = "could not load signature block"
	errSignatureVerification       string = "signature verification failed"
	errSignaturePubKeyMismatch     string = "signature was created by a different key"
	errCreatingBlock                      = "error creating block"
	errWritingBlock                       = "error writing block"
	errGettingHeads                       = "error getting heads"
	errMergingDelta                       = "error merging delta"
	errAddingHead                         = "error adding head"
	errCheckingHead                       = "error checking if is head"
	errReplacingHead                      = "error replacing head"
	errCouldNotFindBlock                  = "error checking for known block "
	errFailedToGetNextQResult             = "failed to get next query result"
	errCouldNotGetEncKey                  = "could not get encryption key"
	errUnsupportedKeyForSigning           = "unsupported key type for signing"
)

// Errors returnable from this package.
//
// This list is incomplete and undefined errors may also be returned.
// Errors returned from this package may be tested against these errors with errors.Is.
var (
	ErrNodeToBlock                 = errors.New(errNodeToBlock)
	ErrEncodingBlock               = errors.New(errEncodingBlock)
	ErrUnmarshallingBlock          = errors.New(errUnmarshallingBlock)
	ErrGeneratingLink              = errors.New(errGeneratingLink)
	ErrInvalidBlockEncryptionType  = errors.New(errInvalidBlockEncryptionType)
	ErrInvalidBlockEncryptionKeyID = errors.New(errInvalidBlockEncryptionKeyID)
	ErrSignatureVerification       = errors.New(errSignatureVerification)
	ErrSignaturePubKeyMismatch     = errors.New(errSignaturePubKeyMismatch)
	ErrCreatingBlock               = errors.New(errCreatingBlock)
	ErrWritingBlock                = errors.New(errWritingBlock)
	ErrGettingHeads                = errors.New(errGettingHeads)
	ErrMergingDelta                = errors.New(errMergingDelta)
	ErrAddingHead                  = errors.New(errAddingHead)
	ErrCheckingHead                = errors.New(errCheckingHead)
	ErrReplacingHead               = errors.New(errReplacingHead)
	ErrCouldNotFindBlock           = errors.New(errCouldNotFindBlock)
	ErrFailedToGetNextQResult      = errors.New(errFailedToGetNextQResult)
	ErrDecodingHeight              = errors.New("error decoding height")
	ErrCouldNotGetEncKey           = errors.New(errCouldNotGetEncKey)
)

// NewErrFailedToGetPriority returns an error indicating that the priority could not be retrieved.
func NewErrNodeToBlock(node any) error {
	return errors.New(
		errNodeToBlock,
		errors.NewKV("Expected", fmt.Sprintf("%T", &Block{})),
		errors.NewKV("Actual", fmt.Sprintf("%T", node)),
	)
}

// NewErrEncodingBlock returns an error indicating that the block could not be encoded.
func NewErrEncodingBlock(err error) error {
	return errors.Wrap(
		errEncodingBlock,
		err,
	)
}

// NewErrUnmarshallingBlock returns an error indicating that the block could not be unmarshalled.
func NewErrUnmarshallingBlock(err error) error {
	return errors.Wrap(
		errUnmarshallingBlock,
		err,
	)
}

// NewErrGeneratingLink returns an error indicating that the link could not be generated.
func NewErrGeneratingLink(err error) error {
	return errors.Wrap(
		errGeneratingLink,
		err,
	)
}

// NewErrCouldNotLoadSignatureBlock returns an error indicating that the signature block could not be found.
func NewErrCouldNotLoadSignatureBlock(err error) error {
	return errors.Wrap(
		errCouldNotLoadSignatureBlock,
		err,
	)
}

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

func NewErrUnsupportedKeyForSigning(keyType crypto.KeyType) error {
	return errors.New(errUnsupportedKeyForSigning, errors.NewKV("KeyType", keyType))
}
