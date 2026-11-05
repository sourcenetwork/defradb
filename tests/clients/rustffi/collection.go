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

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/utils"
)

var _ client.Collection = (*Collection)(nil)

// Collection is one version of a collection on a Rust node, and the
// transaction it was read in, if any.
type Collection struct {
	def client.CollectionVersion
	w   *Wrapper
	txn immutable.Option[client.Txn]
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

// context carries the collection's transaction, when it has one.
func (c *Collection) context(ctx context.Context) context.Context {
	if c.txn.HasValue() {
		return datastore.CtxSetFromClientTxn(ctx, c.txn.Value())
	}
	return ctx
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
	ctx = c.context(ctx)

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
	copts := c.byName().c(&args)
	res := as(utils.NewOptions(opts...).GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_index_new(c.w.target(ctx), args.str(indexDesc.Name),
			args.str(strings.Join(fields, ",")), unique, args.str(vectorJSON), copts, ident))
	})
	return unmarshalResult[client.IndexDescription](res)
}

func (c *Collection) DeleteIndex(
	ctx context.Context,
	indexName string,
	opts ...options.Enumerable[options.DeleteCollectionIndexOptions],
) error {
	ctx = c.context(ctx)
	var args cstrings
	defer args.free()
	copts := c.byName().c(&args)
	return as(utils.NewOptions(opts...).GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_index_delete(c.w.target(ctx), args.str(indexName), copts, ident))
	}).err()
}

func (c *Collection) ListIndexes(
	ctx context.Context,
	opts ...options.Enumerable[options.ListCollectionIndexesOptions],
) ([]client.ListIndexesResult, error) {
	ctx = c.context(ctx)
	var args cstrings
	defer args.free()
	copts := c.byName().c(&args)
	res := as(utils.NewOptions(opts...).GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_index_list(c.w.target(ctx), copts, ident))
	})
	return unmarshalResult[[]client.ListIndexesResult](res)
}

func (c *Collection) NewEncryptedIndex(
	ctx context.Context,
	req client.EncryptedIndexDescription,
	opts ...options.Enumerable[options.NewEncryptedIndexOptions],
) (client.EncryptedIndexDescription, error) {
	ctx = c.context(ctx)
	var args cstrings
	defer args.free()
	res := as(utils.NewOptions(opts...).GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_encrypted_index_new(c.w.target(ctx), args.str(c.def.Name),
			args.str(req.FieldName), args.str(string(req.Type)), ident))
	})
	return unmarshalResult[client.EncryptedIndexDescription](res)
}

func (c *Collection) DeleteEncryptedIndex(
	ctx context.Context,
	fieldName string,
	opts ...options.Enumerable[options.DeleteEncryptedIndexOptions],
) error {
	ctx = c.context(ctx)
	var args cstrings
	defer args.free()
	return as(utils.NewOptions(opts...).GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_encrypted_index_delete(c.w.target(ctx), args.str(c.def.Name),
			args.str(fieldName), ident))
	}).err()
}

func (c *Collection) ListEncryptedIndexes(
	ctx context.Context, opts ...options.Enumerable[options.ListCollectionEncryptedIndexesOptions],
) ([]client.EncryptedIndexDescription, error) {
	ctx = c.context(ctx)
	var args cstrings
	defer args.free()
	res := as(utils.NewOptions(opts...).GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_encrypted_index_list(c.w.target(ctx), args.str(c.def.Name), ident))
	})
	return unmarshalResult[[]client.EncryptedIndexDescription](res)
}

func (c *Collection) Truncate(
	ctx context.Context, opts ...options.Enumerable[options.TruncateCollectionOptions],
) error {
	ctx = c.context(ctx)
	opt := utils.NewOptions(opts...)
	var args cstrings
	defer args.free()
	copts := c.byName().c(&args)
	if opt.Filter == nil {
		return as(opt.GetIdentity(), func(ident C.uintptr_t) result {
			return take(C.defra_collection_truncate(c.w.target(ctx), copts, ident))
		}).err()
	}
	filter, err := json.Marshal(opt.Filter)
	if err != nil {
		return err
	}
	return as(opt.GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_collection_truncate_with_filter(c.w.target(ctx), copts, ident,
			args.str(string(filter))))
	}).err()
}
