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
