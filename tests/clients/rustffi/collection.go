// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

//go:build rust_ffi

package rustffi

/*
#include <stdlib.h>
#include "defra.h"
*/
import "C"

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/utils"
)

var _ client.Collection = (*Collection)(nil)

// Collection is one version of a collection on a Rust node. A collection
// fetched through a transaction keeps it, as Go's do: each call runs in the
// transaction its context carries, else in that one, else on the node.
type Collection struct {
	def client.CollectionVersion
	w   *Wrapper
	txn *Transaction
}

// in is ctx carrying the collection's transaction when it carries none.
func (c *Collection) in(ctx context.Context) context.Context {
	if _, ok := datastore.CtxTryGetClientTxn(ctx); ok || c.txn == nil {
		return ctx
	}
	return c.txn.in(ctx)
}

// target is the handle a call made with ctx runs on.
func (c *Collection) target(ctx context.Context) C.uintptr_t {
	return c.w.target(c.in(ctx))
}

func (c *Collection) Version() client.CollectionVersion {
	return c.def
}

func (c *Collection) Name() string {
	return c.def.Name
}

func (c *Collection) VersionID() string {
	return c.def.VersionID
}

func (c *Collection) CollectionID() string {
	return c.def.CollectionID
}

// byName selects this collection by name.
func (c *Collection) byName() collectionOptions {
	return collectionOptions{Name: c.def.Name}
}

func (c *Collection) NewIndex(
	ctx context.Context,
	indexDesc client.NewIndexRequest,
	opts ...options.Enumerable[options.NewCollectionIndexOptions],
) (client.IndexDescription, error) {

	// The deprecated Fields wins when set, so the node still sees it disagree
	// with the kind config and rejects it.
	//nolint:staticcheck // the deprecated field is still supported until v2.0.0
	requested := indexDesc.Fields
	if len(requested) == 0 {
		requested = indexDesc.GetFields()
	}
	fields := make([]string, len(requested))
	for i, field := range requested {
		order := "ASC"
		if field.Descending {
			order = "DESC"
		}
		fields[i] = field.Name + ":" + order
	}

	unique := C.int(0)
	//nolint:staticcheck // the deprecated field is still supported until v2.0.0
	if indexDesc.Unique || (indexDesc.Ordered != nil && indexDesc.Ordered.Unique) {
		unique = 1
	}

	vectorJSON := ""
	if indexDesc.Vector != nil {
		data, err := json.Marshal(indexDesc.Vector)
		if err != nil {
			return client.IndexDescription{}, err
		}
		vectorJSON = string(data)
	}

	var args cstrings
	defer args.free()
	sel := c.byName().c(&args)
	res := as(utils.NewOptions(opts...).GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_index_new(c.target(ctx), args.str(strings.Join(fields, ",")),
			C.DefraIndexNewOptions{
				identity:                ident,
				collection_name:         sel.name,
				collection_version_id:   sel.version,
				collection_id:           sel.id,
				collection_get_inactive: sel.inactive,
				name:                    args.str(indexDesc.Name),
				unique:                  unique,
				vector_json:             args.str(vectorJSON),
			}))
	})
	return unmarshalResult[client.IndexDescription](res)
}

func (c *Collection) DeleteIndex(
	ctx context.Context,
	indexName string,
	opts ...options.Enumerable[options.DeleteCollectionIndexOptions],
) error {
	var args cstrings
	defer args.free()
	sel := c.byName().c(&args)
	return as(utils.NewOptions(opts...).GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_index_delete(c.target(ctx), args.str(indexName),
			C.DefraIndexDeleteOptions{
				identity:                ident,
				collection_name:         sel.name,
				collection_version_id:   sel.version,
				collection_id:           sel.id,
				collection_get_inactive: sel.inactive,
			}))
	}).err()
}

func (c *Collection) ListIndexes(
	ctx context.Context,
	opts ...options.Enumerable[options.ListCollectionIndexesOptions],
) ([]client.ListIndexesResult, error) {
	var args cstrings
	defer args.free()
	sel := c.byName().c(&args)
	res := as(utils.NewOptions(opts...).GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_index_list(c.target(ctx), C.DefraIndexListOptions{
			identity:                ident,
			collection_name:         sel.name,
			collection_version_id:   sel.version,
			collection_id:           sel.id,
			collection_get_inactive: sel.inactive,
		}))
	})
	return unmarshalResult[[]client.ListIndexesResult](res)
}

func (c *Collection) NewEncryptedIndex(
	ctx context.Context,
	req client.EncryptedIndexDescription,
	opts ...options.Enumerable[options.NewEncryptedIndexOptions],
) (client.EncryptedIndexDescription, error) {
	var args cstrings
	defer args.free()
	res := as(utils.NewOptions(opts...).GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_encrypted_index_new(c.target(ctx), args.str(req.FieldName),
			C.DefraEncryptedIndexNewOptions{
				identity:        ident,
				collection_name: args.str(c.def.Name),
				index_type:      args.str(string(req.Type)),
			}))
	})
	return unmarshalResult[client.EncryptedIndexDescription](res)
}

func (c *Collection) DeleteEncryptedIndex(
	ctx context.Context,
	fieldName string,
	opts ...options.Enumerable[options.DeleteEncryptedIndexOptions],
) error {
	var args cstrings
	defer args.free()
	return as(utils.NewOptions(opts...).GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_encrypted_index_delete(c.target(ctx), args.str(fieldName),
			C.DefraEncryptedIndexDeleteOptions{identity: ident, collection_name: args.str(c.def.Name)}))
	}).err()
}

func (c *Collection) ListEncryptedIndexes(
	ctx context.Context, opts ...options.Enumerable[options.ListCollectionEncryptedIndexesOptions],
) ([]client.EncryptedIndexDescription, error) {
	var args cstrings
	defer args.free()
	res := as(utils.NewOptions(opts...).GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_encrypted_index_list(c.target(ctx),
			C.DefraEncryptedIndexListOptions{identity: ident, collection_name: args.str(c.def.Name)}))
	})
	return unmarshalResult[[]client.EncryptedIndexDescription](res)
}

func (c *Collection) Truncate(
	ctx context.Context, opts ...options.Enumerable[options.TruncateCollectionOptions],
) error {
	opt := utils.NewOptions(opts...)
	var args cstrings
	defer args.free()
	sel := c.byName().c(&args)
	var filter *C.char
	if opt.Filter != nil {
		data, err := json.Marshal(opt.Filter)
		if err != nil {
			return err
		}
		filter = args.str(string(data))
	}
	return as(opt.GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_collection_truncate(c.target(ctx), C.DefraCollectionTruncateOptions{
			identity:                ident,
			collection_name:         sel.name,
			collection_version_id:   sel.version,
			collection_id:           sel.id,
			collection_get_inactive: sel.inactive,
			filter_json:             filter,
		}))
	}).err()
}
