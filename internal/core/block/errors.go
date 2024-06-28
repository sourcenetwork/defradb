// Copyright 2024 Democratized Data Foundation
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

	"github.com/sourcenetwork/defradb/errors"
)

const (
	errNodeToBlock        string = "failed to convert node to block"
	errEncodingBlock      string = "failed to encode block"
	errUnmarshallingBlock string = "failed to unmarshal block"
	errGeneratingLink     string = "failed to generate link"
)

// Errors returnable from this package.
//
// This list is incomplete and undefined errors may also be returned.
// Errors returned from this package may be tested against these errors with errors.Is.
var (
	ErrNodeToBlock        = errors.New(errNodeToBlock)
	ErrEncodingBlock      = errors.New(errEncodingBlock)
	ErrUnmarshallingBlock = errors.New(errUnmarshallingBlock)
	ErrGeneratingLink     = errors.New(errGeneratingLink)
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
