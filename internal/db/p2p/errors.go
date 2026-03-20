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

const (
	errStoreBlockDAGSync  string = "failed to store block in DAG sync"
	errGenerateBlockLink  string = "failed to generate block link"
	errCheckBlockMerged   string = "failed to check if block is merged"
	errVerifyBlockSig     string = "failed to verify block signature"
	errGetEncKeysForBlock string = "failed to get encryption keys for block"
	errLoadLinkedBlock    string = "failed to load linked block during DAG sync"
	errDecodeLinkedBlock  string = "failed to decode linked block during DAG sync"
	errProcessLinkedBlock string = "failed to process linked block during DAG sync"
	errRetrieveEncKey     string = "failed to retrieve encryption key during DAG sync"
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

func NewErrStoreBlockDAGSync(inner error) error  { return errors.Wrap(errStoreBlockDAGSync, inner) }
func NewErrGenerateBlockLink(inner error) error  { return errors.Wrap(errGenerateBlockLink, inner) }
func NewErrCheckBlockMerged(inner error) error   { return errors.Wrap(errCheckBlockMerged, inner) }
func NewErrVerifyBlockSig(inner error) error     { return errors.Wrap(errVerifyBlockSig, inner) }
func NewErrGetEncKeysForBlock(inner error) error { return errors.Wrap(errGetEncKeysForBlock, inner) }
func NewErrLoadLinkedBlock(inner error) error    { return errors.Wrap(errLoadLinkedBlock, inner) }
func NewErrDecodeLinkedBlock(inner error) error  { return errors.Wrap(errDecodeLinkedBlock, inner) }
func NewErrProcessLinkedBlock(inner error) error { return errors.Wrap(errProcessLinkedBlock, inner) }
func NewErrRetrieveEncKey(inner error) error     { return errors.Wrap(errRetrieveEncKey, inner) }
