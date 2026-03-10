// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package p2p

import (
	"github.com/sourcenetwork/defradb/errors"
)

var (
	ErrSelfTargetForReplicator     = errors.New("can't target ourselves as a replicator")
	ErrReplicatorNotFound          = errors.New("replicator not found")
	ErrReplicatorCollections       = errors.New("failed to get collections for replicator")
	ErrContextDone                 = errors.New("context done")
	errPublishingToDocIDTopic      = errors.New("can't publish log for document")
	errPublishingToCollectionTopic = errors.New("can't publish log for collection")
	ErrTimeoutDocSync              = errors.New("timeout while syncing doc")
	ErrTimeoutCollectionSync       = errors.New("timeout while syncing branchable collection")
	ErrCollectionNotBranchable     = errors.New("collection is not branchable")
	ErrNoHeadsForBranchableCol     = errors.New("no heads found for branchable collection")
)

func NewErrReplicatorCollections(inner error, kv ...errors.KV) error {
	return errors.WithStack(errors.Join(ErrReplicatorCollections, inner), kv...)
}

func NewErrPublishingToDocIDTopic(inner error, cid, docID string) error {
	return errors.WithStack(
		errors.Join(inner, errPublishingToDocIDTopic),
		errors.NewKV("CID", cid),
		errors.NewKV("DocID", docID),
	)
}

func NewErrPublishingToCollectionTopic(inner error, cid, colID string) error {
	return errors.WithStack(
		errors.Join(inner, errPublishingToCollectionTopic),
		errors.NewKV("CID", cid),
		errors.NewKV("CollectionID", colID),
	)
}

func NewErrCollectionNotBranchable(collectionID string) error {
	return errors.WithStack(
		ErrCollectionNotBranchable,
		errors.NewKV("CollectionID", collectionID),
	)
}

func NewErrNoHeadsForBranchableCol(collectionID string) error {
	return errors.WithStack(
		ErrNoHeadsForBranchableCol,
		errors.NewKV("CollectionID", collectionID),
	)
}
