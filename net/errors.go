// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package net

import (
	"github.com/sourcenetwork/defradb/errors"
)

/*const (
	errPushLog                  = "failed to push log"
	errFailedToGetDocID         = "failed to get DocID from broadcast message"
	errPublishingToDocIDTopic   = "can't publish log %s for docID %s"
	errPublishingToSchemaTopic  = "can't publish log %s for schema %s"
	errCheckingForExistingBlock = "failed to check for existing block"
	errRequestingEncryptionKeys = "failed to request encryption keys with %v"
	errTopicAlreadyExist        = "topic with name \"%s\" already exists"
	errTopicDoesNotExist        = "topic with name \"%s\" does not exists"
	errFailedToGetIdentity      = "failed to get identity"
	errReplicatorCollections    = "failed to get collections for replicator"
	errPushSEArtifacts          = "failed to push SE artifacts"
	errQuerySEArtifacts         = "failed to query SE artifacts"
)*/

var (
	ErrPushLog                   = errors.New("failed to push log")
	ErrTopicAlreadyExist         = errors.New("topic already exists")
	ErrTopicDoesNotExist         = errors.New("topic does not exists")
	ErrTimeoutWaitingForPeerInfo = errors.New("timeout waiting for peer info")
	ErrContextDone               = errors.New("context done")
)

func NewErrPushLog(inner error, kv ...errors.KV) error {
	return errors.WithStack(errors.Join(inner, ErrPushLog), kv...)
}

func NewErrTopicAlreadyExist(topic string) error {
	return errors.WithStack(ErrTopicAlreadyExist, errors.NewKV("topic", topic))
}

func NewErrTopicDoesNotExist(topic string) error {
	return errors.WithStack(ErrTopicDoesNotExist, errors.NewKV("topic", topic))
}

/*func NewErrFailedToGetIdentity(inner error, kv ...errors.KV) error {
	return errors.Wrap(errFailedToGetIdentity, inner, kv...)
}

func NewErrReplicatorCollections(inner error, kv ...errors.KV) error {
	return errors.Wrap(errReplicatorCollections, inner, kv...)
}

func NewErrPushSEArtifacts(inner error, kv ...errors.KV) error {
	return errors.Wrap(errPushSEArtifacts, inner, kv...)
}

func NewErrQuerySEArtifacts(inner error, kv ...errors.KV) error {
	return errors.Wrap(errQuerySEArtifacts, inner, kv...)
}*/
