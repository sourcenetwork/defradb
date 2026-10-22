// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package vectorindex

import (
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
)

const (
	errUnsupportedAlgorithm string = "unsupported vector index algorithm"
	errUnsupportedMetric    string = "unsupported vector index distance metric"
	errVectorIndexStore     string = "vector index store operation failed"
	errMalformedPending     string = "malformed vector pending entry"
	errUnknownPendingOp     string = "unknown vector pending entry op"
)

// newErrUnsupportedAlgorithm returns an error indicating the index description asked for a vector
// algorithm that has no implementation.
func newErrUnsupportedAlgorithm(a client.VectorAlgorithm) error {
	return errors.New(errUnsupportedAlgorithm, errors.NewKV("Algorithm", a))
}

// newErrUnsupportedMetric returns an error indicating the index asked for a distance metric the
// selected algorithm does not support.
func newErrUnsupportedMetric(m client.DistanceMetric) error {
	return errors.New(errUnsupportedMetric, errors.NewKV("Metric", m))
}

// newErrMalformedPending returns an error indicating a stored pending entry for the node could not be
// decoded because its length does not fit an op byte followed by float32s.
func newErrMalformedPending(nodeID uint64, length int) error {
	return errors.New(errMalformedPending, errors.NewKV("NodeID", nodeID), errors.NewKV("Length", length))
}

// newErrUnknownPendingOp returns an error indicating a stored pending entry for the node carries an op
// byte that is neither an upsert nor a delete.
func newErrUnknownPendingOp(nodeID uint64, op PendingOp) error {
	return errors.New(errUnknownPendingOp, errors.NewKV("NodeID", nodeID), errors.NewKV("Op", op))
}

// newErrVectorIndexStore returns a new error indicating that a vector index store operation
// (against its datastore-backed node/meta keyspace) failed. The keyspace coordinates are attached so
// the failure names its index when a backfill runs many at once.
func newErrVectorIndexStore(inner error, collectionShortID, indexID, epoch uint32) error {
	return errors.Wrap(
		errVectorIndexStore,
		inner,
		errors.NewKV("CollectionShortID", collectionShortID),
		errors.NewKV("IndexID", indexID),
		errors.NewKV("Epoch", epoch),
	)
}
