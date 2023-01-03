// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package mapper

import (
	"context"
	"encoding/json"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/errors"
)

// DescriptionsRepo is a cache of previously requested collection descriptions
// that can be used to reduce multiple reads of the same collection description.
type DescriptionsRepo struct {
	ctx context.Context
	txn datastore.Txn

	collectionDescriptionsByName map[string]client.CollectionDescription
}

// NewDescriptionsRepo instantiates a new DescriptionsRepo with the given context and transaction.
func NewDescriptionsRepo(ctx context.Context, txn datastore.Txn) *DescriptionsRepo {
	return &DescriptionsRepo{
		ctx:                          ctx,
		txn:                          txn,
		collectionDescriptionsByName: map[string]client.CollectionDescription{},
	}
}

// getCollectionDesc returns the description of the collection with the given name.
//
// Will return nil and an error if a description of the given name is not found. Will first look
// in the repo's cache for the description before querying the datastore.
func (r *DescriptionsRepo) getCollectionDesc(name string) (client.CollectionDescription, error) {
	collectionKey := core.NewCollectionKey(name)
	var desc client.CollectionDescription
	collectionBuf, err := r.txn.Systemstore().Get(r.ctx, collectionKey.ToDS())
	if err != nil {
		return desc, errors.Wrap("failed to get collection description", err)
	}

	schemaVersionId := string(collectionBuf)
	schemaVersionKey := core.NewCollectionSchemaVersionKey(schemaVersionId)
	buf, err := r.txn.Systemstore().Get(r.ctx, schemaVersionKey.ToDS())
	if err != nil {
		return desc, err
	}

	err = json.Unmarshal(buf, &desc)
	if err != nil {
		return desc, err
	}

	err = json.Unmarshal(buf, &desc)
	if err != nil {
		return desc, err
	}

	return desc, nil
}
