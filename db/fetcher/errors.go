// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package fetcher

import (
	"github.com/sourcenetwork/defradb/errors"
)

const (
	errFieldIdNotFound              string = "unable to find FieldDescription for given FieldId"
	errFailedToDecodeCIDForVFetcher string = "failed to decode CID for VersionedFetcher"
	errFailedToSeek                 string = "seek failed"
	errFailedToMergeState           string = "failed merging state"
	errVFetcherFailedToFindBlock    string = "(version fetcher) failed to find block in blockstore"
	errVFetcherFailedToGetBlock     string = "(version fetcher) failed to get block in blockstore"
	errVFetcherFailedToWriteBlock   string = "(version fetcher) failed to write block to blockstore"
	errVFetcherFailedToDecodeNode   string = "(version fetcher) failed to decode protobuf"
	errVFetcherFailedToGetDagLink   string = "(version fetcher) failed to get node link from DAG"
	errFailedToGetDagNode           string = "failed to get DAG Node"
)

var (
	ErrFieldIdNotFound              = errors.New(errFieldIdNotFound)
	ErrFailedToDecodeCIDForVFetcher = errors.New(errFailedToDecodeCIDForVFetcher)
	ErrFailedToSeek                 = errors.New(errFailedToSeek)
	ErrFailedToMergeState           = errors.New(errFailedToMergeState)
	ErrVFetcherFailedToFindBlock    = errors.New(errVFetcherFailedToFindBlock)
	ErrVFetcherFailedToGetBlock     = errors.New(errVFetcherFailedToGetBlock)
	ErrVFetcherFailedToWriteBlock   = errors.New(errVFetcherFailedToWriteBlock)
	ErrVFetcherFailedToDecodeNode   = errors.New(errVFetcherFailedToDecodeNode)
	ErrVFetcherFailedToGetDagLink   = errors.New(errVFetcherFailedToGetDagLink)
	ErrFailedToGetDagNode           = errors.New(errFailedToGetDagNode)
	ErrSingleSpanOnly               = errors.New("spans must contain only a single entry")
)

func NewErrFieldIdNotFound(fieldId uint32) error {
	return errors.New(errFieldIdNotFound, errors.NewKV("FieldId", fieldId))
}

func NewErrFailedToDecodeCIDForVFetcher(inner error) error {
	return errors.Wrap(errFailedToDecodeCIDForVFetcher, inner)
}

func NewErrFailedToSeek(target any, inner error) error {
	return errors.Wrap(errFailedToSeek, inner, errors.NewKV("Target", target))
}

func NewErrFailedToMergeState(inner error) error {
	return errors.Wrap(errFailedToMergeState, inner)
}

func NewErrVFetcherFailedToFindBlock(inner error) error {
	return errors.Wrap(errVFetcherFailedToFindBlock, inner)
}

func NewErrVFetcherFailedToGetBlock(inner error) error {
	return errors.Wrap(errVFetcherFailedToGetBlock, inner)
}

func NewErrVFetcherFailedToWriteBlock(inner error) error {
	return errors.Wrap(errVFetcherFailedToWriteBlock, inner)
}

func NewErrVFetcherFailedToDecodeNode(inner error) error {
	return errors.Wrap(errVFetcherFailedToDecodeNode, inner)
}

func NewErrVFetcherFailedToGetDagLink(inner error) error {
	return errors.Wrap(errVFetcherFailedToGetDagLink, inner)
}

func NewErrFailedToGetDagNode(inner error) error {
	return errors.Wrap(errFailedToGetDagNode, inner)
}
