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
	errStoreBlockDAGSync       string = "failed to store block in DAG sync"
	errGenerateBlockLink       string = "failed to generate block link"
	errCheckBlockMerged        string = "failed to check if block is merged"
	errVerifyBlockSig          string = "failed to verify block signature"
	errGetEncKeysForBlock      string = "failed to get encryption keys for block"
	errLoadLinkedBlock         string = "failed to load linked block during DAG sync"
	errDecodeLinkedBlock       string = "failed to decode linked block during DAG sync"
	errProcessLinkedBlock      string = "failed to process linked block during DAG sync"
	errRetrieveEncKey          string = "failed to retrieve encryption key during DAG sync"
	errStoreP2PCollection      string = "failed to store P2P collection"
	errStoreP2PDocument        string = "failed to store P2P document"
	errDeleteP2PCollection     string = "failed to delete P2P collection"
	errDeleteP2PDocument       string = "failed to delete P2P document"
	errCheckReplicatorExists   string = "failed to check if replicator exists"
	errGetReplicator           string = "failed to get replicator"
	errUnmarshalReplicator     string = "failed to unmarshal replicator"
	errMarshalReplicator       string = "failed to marshal replicator"
	errStoreReplicator         string = "failed to store replicator"
	errDeleteReplicator        string = "failed to delete replicator"
	errListReplicators         string = "failed to list replicators"
	errCreateDocIterator       string = "failed to create document iterator for replicator"
	errIterateReplicatorDocs   string = "failed to iterate replicator documents"
	errPushDocHeads            string = "failed to push document heads"
	errGetDocHeads             string = "failed to get document heads for replication"
	errMarshalBlock            string = "failed to marshal block for replication"
	errUpdateReplicatorStatus  string = "failed to update replicator status"
	errCreateReplicatorRetry   string = "failed to create replicator retry"
	errStoreRetryDoc           string = "failed to store retry doc for replicator"
	errHandleRetryCompletion   string = "failed to handle replicator retry completion"
	errCheckRetryExists        string = "failed to check if replicator retry exists"
	errMarshalRetryInfo        string = "failed to marshal replicator retry info"
	errStoreRetryInfo          string = "failed to store replicator retry info"
	errGetRetryInfo            string = "failed to get replicator retry info"
	errUnmarshalRetryInfo      string = "failed to unmarshal replicator retry info"
	errCreateHeadstoreIterator string = "failed to create headstore iterator for replication"
	errSendReplicatorRequest   string = "failed to send replicator push request"
	errFetchRetryDocs          string = "failed to fetch retry docs for replicator"
	errDeleteRetryKey          string = "failed to delete replicator retry key"
	errDeleteRetryDoc          string = "failed to delete replicator retry doc"
	errListP2PCollections      string = "failed to list P2P collections"
	errGetAllP2PCollections    string = "failed to get all P2P collection IDs"
	errListP2PDocuments        string = "failed to list P2P documents"
	errLoadP2PDocuments        string = "failed to load P2P documents"
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

func NewErrStoreP2PCollection(inner error, collectionID string) error {
	return errors.Wrap(errStoreP2PCollection, inner, errors.NewKV("CollectionID", collectionID))
}

func NewErrStoreP2PDocument(inner error, docID string) error {
	return errors.Wrap(errStoreP2PDocument, inner, errors.NewKV("DocID", docID))
}

func NewErrCheckReplicatorExists(inner error, peerID string) error {
	return errors.Wrap(errCheckReplicatorExists, inner, errors.NewKV("PeerID", peerID))
}

func NewErrGetReplicator(inner error, peerID string) error {
	return errors.Wrap(errGetReplicator, inner, errors.NewKV("PeerID", peerID))
}

func NewErrUnmarshalReplicator(inner error, peerID string) error {
	return errors.Wrap(errUnmarshalReplicator, inner, errors.NewKV("PeerID", peerID))
}

func NewErrMarshalReplicator(inner error, peerID string) error {
	return errors.Wrap(errMarshalReplicator, inner, errors.NewKV("PeerID", peerID))
}

func NewErrStoreReplicator(inner error, peerID string) error {
	return errors.Wrap(errStoreReplicator, inner, errors.NewKV("PeerID", peerID))
}

func NewErrCreateDocIterator(inner error) error {
	return errors.Wrap(errCreateDocIterator, inner)
}

func NewErrIterateReplicatorDocs(inner error) error {
	return errors.Wrap(errIterateReplicatorDocs, inner)
}

func NewErrPushDocHeads(inner error, docID string) error {
	return errors.Wrap(errPushDocHeads, inner, errors.NewKV("DocID", docID))
}

func NewErrGetDocHeads(inner error, docID string) error {
	return errors.Wrap(errGetDocHeads, inner, errors.NewKV("DocID", docID))
}

func NewErrMarshalBlock(inner error, docID string, cid string) error {
	return errors.Wrap(errMarshalBlock, inner, errors.NewKV("DocID", docID), errors.NewKV("CID", cid))
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

func NewErrDeleteReplicator(inner error, peerID string) error {
	return errors.Wrap(errDeleteReplicator, inner, errors.NewKV("PeerID", peerID))
}

func NewErrListReplicators(inner error) error {
	return errors.Wrap(errListReplicators, inner)
}

func NewErrUpdateReplicatorStatus(inner error, peerID string) error {
	return errors.Wrap(errUpdateReplicatorStatus, inner, errors.NewKV("PeerID", peerID))
}

func NewErrCreateReplicatorRetry(inner error, peerID string) error {
	return errors.Wrap(errCreateReplicatorRetry, inner, errors.NewKV("PeerID", peerID))
}

func NewErrStoreRetryDoc(inner error, peerID, docID string) error {
	return errors.Wrap(errStoreRetryDoc, inner, errors.NewKV("PeerID", peerID), errors.NewKV("DocID", docID))
}

func NewErrHandleRetryCompletion(inner error, peerID string) error {
	return errors.Wrap(errHandleRetryCompletion, inner, errors.NewKV("PeerID", peerID))
}

func NewErrCheckRetryExists(inner error, peerID string) error {
	return errors.Wrap(errCheckRetryExists, inner, errors.NewKV("PeerID", peerID))
}

func NewErrMarshalRetryInfo(inner error, peerID string) error {
	return errors.Wrap(errMarshalRetryInfo, inner, errors.NewKV("PeerID", peerID))
}

func NewErrStoreRetryInfo(inner error, peerID string) error {
	return errors.Wrap(errStoreRetryInfo, inner, errors.NewKV("PeerID", peerID))
}

func NewErrGetRetryInfo(inner error, peerID string) error {
	return errors.Wrap(errGetRetryInfo, inner, errors.NewKV("PeerID", peerID))
}

func NewErrUnmarshalRetryInfo(inner error, peerID string) error {
	return errors.Wrap(errUnmarshalRetryInfo, inner, errors.NewKV("PeerID", peerID))
}

func NewErrCreateHeadstoreIterator(inner error, docID string) error {
	return errors.Wrap(errCreateHeadstoreIterator, inner, errors.NewKV("DocID", docID))
}

func NewErrSendReplicatorRequest(inner error, peerID, docID string) error {
	return errors.Wrap(errSendReplicatorRequest, inner, errors.NewKV("PeerID", peerID), errors.NewKV("DocID", docID))
}

func NewErrFetchRetryDocs(inner error, peerID string) error {
	return errors.Wrap(errFetchRetryDocs, inner, errors.NewKV("PeerID", peerID))
}

func NewErrDeleteRetryKey(inner error, peerID string) error {
	return errors.Wrap(errDeleteRetryKey, inner, errors.NewKV("PeerID", peerID))
}

func NewErrDeleteRetryDoc(inner error, peerID string) error {
	return errors.Wrap(errDeleteRetryDoc, inner, errors.NewKV("PeerID", peerID))
}

func NewErrDeleteP2PCollection(inner error, collectionID string) error {
	return errors.Wrap(errDeleteP2PCollection, inner, errors.NewKV("CollectionID", collectionID))
}

func NewErrDeleteP2PDocument(inner error, docID string) error {
	return errors.Wrap(errDeleteP2PDocument, inner, errors.NewKV("DocID", docID))
}

func NewErrListP2PCollections(inner error) error {
	return errors.Wrap(errListP2PCollections, inner)
}

func NewErrGetAllP2PCollections(inner error) error {
	return errors.Wrap(errGetAllP2PCollections, inner)
}

func NewErrListP2PDocuments(inner error) error {
	return errors.Wrap(errListP2PDocuments, inner)
}

func NewErrLoadP2PDocuments(inner error) error {
	return errors.Wrap(errLoadP2PDocuments, inner)
}
