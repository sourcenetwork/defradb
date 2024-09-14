// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package kms

import (
	"context"

	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/sourcenetwork/corelog"

	"github.com/sourcenetwork/defradb/internal/encryption"
)

var (
	log = corelog.NewLogger("kms")
)

type ServiceType string

const (
	PubSubServiceType ServiceType = "pubsub"
)

type Service interface {
	GetKeys(ctx context.Context, cids ...cidlink.Link) (*encryption.Results, error)
}
