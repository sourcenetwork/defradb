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
	"fmt"

	"github.com/sourcenetwork/defradb/errors"
)

const (
	errPushLog                  = "failed to push log"
	errFailedToGetDocID         = "failed to get DocID from broadcast message"
	errPublishingToDocIDTopic   = "can't publish log %s for docID %s"
	errPublishingToSchemaTopic  = "can't publish log %s for schema %s"
	errCheckingForExistingBlock = "failed to check for existing block"
	errRequestingEncryptionKeys = "failed to request encryption keys with %v"
	errTopicAlreadyExist        = "topic with name \"%s\" already exists"
	errTopicDoesNotExist        = "topic with name \"%s\" does not exists"
)

var (
	ErrPeerConnectionWaitTimout = errors.New("waiting for peer connection timed out")
	ErrPubSubWaitTimeout        = errors.New("waiting for pubsub timed out")
	ErrPushLogWaitTimeout       = errors.New("waiting for pushlog timed out")
	ErrNilDB                    = errors.New("database object can't be nil")
	ErrNilUpdateChannel         = errors.New("tried to subscribe to update channel, but update channel is nil")
	ErrCheckingForExistingBlock = errors.New(errCheckingForExistingBlock)
)

func NewErrPushLog(inner error, kv ...errors.KV) error {
	return errors.Wrap(errPushLog, inner, kv...)
}

func NewErrFailedToGetDocID(inner error, kv ...errors.KV) error {
	return errors.Wrap(errFailedToGetDocID, inner, kv...)
}

func NewErrPublishingToDocIDTopic(inner error, cid, docID string, kv ...errors.KV) error {
	return errors.Wrap(fmt.Sprintf(errPublishingToDocIDTopic, cid, docID), inner, kv...)
}

func NewErrPublishingToSchemaTopic(inner error, cid, docID string, kv ...errors.KV) error {
	return errors.Wrap(fmt.Sprintf(errPublishingToSchemaTopic, cid, docID), inner, kv...)
}

func NewErrTopicAlreadyExist(topic string) error {
	return errors.New(fmt.Sprintf(errTopicAlreadyExist, topic))
}

func NewErrTopicDoesNotExist(topic string) error {
	return errors.New(fmt.Sprintf(errTopicDoesNotExist, topic))
}
