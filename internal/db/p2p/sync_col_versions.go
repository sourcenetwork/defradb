// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

//go:build !js

package p2p

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime/linking"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corelog"
	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/lens/host-go/p2p"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/errors"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/description"
)

func (p *P2P) SyncCollectionVersions(ctx context.Context, versionIDs ...string) error {
	log.InfoContext(ctx, "Starting collection version sync", corelog.Int("VersionIDCount", len(versionIDs)))

	linkSys := makeLinkSystem(p.host.IPLDStore())

	for _, versionID := range versionIDs {
		_, err := p.syncCollectionVersion(ctx, versionID, linkSys)
		if err != nil {
			return err
		}
	}

	log.InfoContext(ctx, "Collection version sync completed")
	return nil
}

func (p *P2P) syncCollectionVersion(
	ctx context.Context,
	versionID string,
	linkSys linking.LinkSystem,
) (client.CollectionVersion, error) {
	col, err := description.GetCollectionByID(ctx, p.collectionRepository, versionID)
	if err != nil {
		if !errors.Is(err, client.ErrCollectionNotFound) {
			return client.CollectionVersion{}, err
		}
		// If it is not found, continue and try and sync it!
	} else {
		if col.IsPlaceholder {
			// If the collection is a placeholder locally, we should try and sync its proper definition
			// from across the network.
		} else {
			// If the collection exists locally, it is important to return it - this way
			// anything locally defined on it will be preserved.
			return col, nil
		}
	}

	cid, err := cid.Parse(versionID)
	if err != nil {
		return client.CollectionVersion{}, err
	}

	nd, err := linkSys.Load(linking.LinkContext{Ctx: ctx}, cidlink.Link{Cid: cid}, coreblock.BlockSchemaPrototype)
	if err != nil {
		return client.CollectionVersion{}, err
	}

	linkBlock, err := coreblock.GetFromNode(nd)
	if err != nil {
		return client.CollectionVersion{}, err
	}

	var collectionID string
	var previous immutable.Option[client.CollectionSource]
	if len(linkBlock.Heads) == 1 {
		// There can only ever be one or zero heads for a collection version.
		// If there is one, this must be an new version of a collection and
		// we need to sync the older version(s) recursively.
		previousID := linkBlock.Heads[0].String()
		col, err = p.syncCollectionVersion(ctx, previousID, linkSys)
		if err != nil {
			return client.CollectionVersion{}, err
		}

		previous = immutable.Some(
			client.CollectionSource{
				SourceCollectionID: previousID,
			},
		)
		collectionID = col.CollectionID
	}

	fields := make([]client.CollectionFieldDescription, 0, len(col.Fields))
	fields = append(fields, col.Fields...)

	if len(fields) == 0 && len(linkBlock.Links) > 0 {
		// Ensure the length is at least one, if this is the first time receiving fields for this
		// node/collection, the first field *must* be the docID field.
		fields = make([]client.CollectionFieldDescription, 1, len(linkBlock.Links))
	}

	for _, fieldCid := range linkBlock.Links {
		fieldNode, err := linkSys.Load(linking.LinkContext{Ctx: ctx}, fieldCid, coreblock.BlockSchemaPrototype)
		if err != nil {
			return client.CollectionVersion{}, err
		}

		fieldLinkBlock, err := coreblock.GetFromNode(fieldNode)
		if err != nil {
			return client.CollectionVersion{}, err
		}

		fieldDelta := fieldLinkBlock.Delta.FieldDefinitionDelta

		// WARNING - At the moment fields can only ever be added, the following code relies on this.
		// When we allow the mutation of fields, this code will need to change.
		var kind client.FieldKind
		if fieldDelta.RelativeID != nil {
			kind = &client.SelfKind{
				RelativeID: strconv.Itoa(*fieldDelta.RelativeID),
				// The secondary side of SelfKind relationships are never represented by
				// blocks, so we can safely hardcode this to `false`
				Array: false,
			}
		} else if fieldDelta.CollectionID != nil {
			kind = &client.CollectionKind{
				CollectionID: *fieldDelta.CollectionID,
				// The secondary side of CollectionKind relationships are never represented by
				// blocks, so we can safely hardcode this to `false`
				Array: false,
			}
		} else {
			kind = client.IntToFieldKind(*fieldDelta.ScalarKind)
		}

		field := client.CollectionFieldDescription{
			FieldID: fieldCid.String(),
			Name:    *fieldDelta.Name,
			Typ:     *fieldDelta.Crdt,
			Kind:    kind,
		}

		if *fieldDelta.Name == request.DocIDFieldName {
			// The first field *must* be the doc id field - this is a largely a cosmetic
			// decision, however by now some code does rely on this.
			fields[0] = field
		} else {
			fields = append(
				fields,
				field,
			)
		}
	}

	var query immutable.Option[client.QuerySource]
	if linkBlock.Delta.CollectionDefinitionDelta.QuerySelect != nil {
		var q request.Select
		err = json.Unmarshal(linkBlock.Delta.CollectionDefinitionDelta.QuerySelect, &q)
		if err != nil {
			return client.CollectionVersion{}, err
		}

		var transform immutable.Option[string]
		if linkBlock.Delta.CollectionDefinitionDelta.QueryTransform != nil {
			err = p.getLensP2P(ctx).SyncLens(ctx, linkBlock.Delta.CollectionDefinitionDelta.QueryTransform.String())
			if err != nil {
				return client.CollectionVersion{}, err
			}

			transform = immutable.Some(linkBlock.Delta.CollectionDefinitionDelta.QueryTransform.String())
		}

		query = immutable.Some(client.QuerySource{
			Query:     q,
			Transform: transform,
		})
	}

	// Merge the details taken from this block onto the previous

	if len(collectionID) > 0 && collectionID != client.OrphanCollectionID {
		col.CollectionID = collectionID
	} else {
		col.CollectionID = versionID
	}

	col.VersionID = versionID
	col.PreviousVersion = previous
	col.Query = query
	col.Fields = fields
	// Ensure that this newly received version is inactive, and that we have not copied
	// IsActive from a locally known version.
	col.IsActive = false

	if linkBlock.Delta.CollectionDefinitionDelta.Name != nil {
		col.Name = *linkBlock.Delta.CollectionDefinitionDelta.Name
	}

	// Non-views must be materialized.  Views are synced as non-materialized - users
	// can toggle this locally if they like.
	col.IsMaterialized = !query.HasValue()

	err = description.SaveCollection(ctx, p.collectionRepository, col)
	if err != nil {
		return client.CollectionVersion{}, err
	}

	return col, nil
}

type wrappedTxn struct {
	datastore.Txn
	corekv.ReaderWriter
}

func (p *P2P) getLensP2P(ctx context.Context) p2p.P2P {
	clientTxn := datastore.CtxMustGetClientTxn(ctx)
	dsTxn := datastore.MustGetFromClientTxn(clientTxn)
	return p.lens.P2P.Value().WithTxn(&wrappedTxn{
		Txn:          dsTxn,
		ReaderWriter: dsTxn.Rootstore(),
	})
}
