// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package se

import (
	"context"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/internal/db/txnctx"
	secore "github.com/sourcenetwork/defradb/internal/se/core"
)

type contextKey struct{}

// Config represents SE configuration
type Config struct {
	Key             []byte
	CollectionID    string
	EncryptedFields []client.EncryptedIndexDescription
}

// Context represents SE context stored in context.Context
type Context struct {
	config    Config
	artifacts []secore.Artifact
	doc       *client.Document
	txn       datastore.Txn
}

// PrepareContextIfConfigured checks collection configuration and prepares SE context if needed
func PrepareContextIfConfigured(
	ctx context.Context,
	col client.Collection,
	doc *client.Document,
	seKey []byte,
) (context.Context, error) {
	encryptedIndexes, err := col.GetEncryptedIndexes(ctx)
	if err != nil {
		return ctx, err
	}
	if len(encryptedIndexes) == 0 || len(seKey) == 0 {
		return ctx, nil
	}

	txn := txnctx.MustGet(ctx)

	seCtx := &Context{
		config: Config{
			Key:             seKey,
			CollectionID:    col.VersionID(),
			EncryptedFields: encryptedIndexes,
		},
		artifacts: make([]secore.Artifact, 0),
		doc:       doc,
		txn:       txn,
	}

	seCtx.registerReplicationCallback()

	return context.WithValue(ctx, contextKey{}, seCtx), nil
}

// registerReplicationCallback sets up transaction callback for artifact replication
func (c *Context) registerReplicationCallback() {
	c.txn.OnSuccess(func() {
		if len(c.artifacts) == 0 {
			return
		}

		docID := c.doc.ID().String()
		for i := range c.artifacts {
			c.artifacts[i].DocID = docID
		}

		// c.db.QueueSEArtifacts(c.artifacts)
	})
}
