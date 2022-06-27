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
	"fmt"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore"
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
	if desc, hasDesc := r.collectionDescriptionsByName[name]; hasDesc {
		return desc, nil
	}

	key := core.NewCollectionKey(name)
	buf, err := r.txn.Systemstore().Get(r.ctx, key.ToDS())
	if err != nil {
		return client.CollectionDescription{}, fmt.Errorf("Failed to get collection description: %w", err)
	}

	desc := client.CollectionDescription{}
	err = json.Unmarshal(buf, &desc)
	if err != nil {
		return client.CollectionDescription{}, err
	}

	r.collectionDescriptionsByName[name] = desc
	return desc, nil
}
