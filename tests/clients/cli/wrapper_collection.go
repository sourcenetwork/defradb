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

package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/internal/utils"
)

// appendIdentityArg extracts identity from an immutable.Option and appends --identity flag to args
// if identity is present and is a FullIdentity with a private key.
func appendIdentityArg(args []string, ident immutable.Option[identity.Identity]) []string {
	if !ident.HasValue() {
		return args
	}
	if fullIdent, ok := ident.Value().(identity.FullIdentity); ok {
		rawIdent := fullIdent.IntoRawIdentity()
		if rawIdent.PrivateKey != "" {
			args = append(args, "--identity", rawIdent.PrivateKey)
		}
	}
	return args
}

// appendTxnArg extracts a transaction from an immutable.Option and appends --tx flag to args
// if transaction is present.
func appendTxnArg(args []string, txn immutable.Option[client.Txn]) []string {
	if !txn.HasValue() {
		return args
	}
	args = append(args, "--tx", fmt.Sprintf("%d", txn.Value().ID()))
	return args
}

var _ client.Collection = (*Collection)(nil)

type Collection struct {
	cmd *cliWrapper
	def client.CollectionVersion
	txn immutable.Option[client.Txn]
}

func (c *Collection) Version() client.CollectionVersion {
	return c.def
}

func (c *Collection) Name() string {
	return c.Version().Name
}

func (c *Collection) VersionID() string {
	return c.Version().VersionID
}

func (c *Collection) CollectionID() string {
	return c.Version().CollectionID
}

func (c *Collection) NewIndex(
	ctx context.Context,
	indexDesc client.NewIndexRequest,
	opts ...options.Enumerable[options.NewCollectionIndexOptions],
) (index client.IndexDescription, err error) {
	args := []string{"client", "index", "new"}
	args = append(args, "--collection", c.Version().Name)
	if indexDesc.Name != "" {
		args = append(args, "--name", indexDesc.Name)
	}
	if indexDesc.Unique {
		args = append(args, "--unique")
	}

	fields := make([]string, len(indexDesc.Fields))
	orders := make([]bool, len(indexDesc.Fields))

	for i := range indexDesc.Fields {
		fields[i] = indexDesc.Fields[i].Name
		orders[i] = indexDesc.Fields[i].Descending
	}

	orderedFields := make([]string, len(fields))

	for i := range fields {
		if orders[i] {
			orderedFields[i] = fields[i] + ":DESC"
		} else {
			orderedFields[i] = fields[i] + ":ASC"
		}
	}

	args = append(args, "--fields", strings.Join(orderedFields, ","))

	opt := utils.NewOptions(opts...)
	args = appendIdentityArg(args, opt.GetIdentity())
	args = appendTxnArg(args, c.txn)

	data, err := c.cmd.execute(ctx, args)
	if err != nil {
		return index, err
	}
	if err := json.Unmarshal(data, &index); err != nil {
		return index, err
	}
	return index, nil
}

func (c *Collection) DeleteIndex(
	ctx context.Context,
	indexName string,
	opts ...options.Enumerable[options.DeleteCollectionIndexOptions],
) error {
	args := []string{"client", "index", "delete"}
	args = append(args, "--collection", c.Version().Name)
	args = append(args, "--name", indexName)

	opt := utils.NewOptions(opts...)
	args = appendIdentityArg(args, opt.GetIdentity())
	args = appendTxnArg(args, c.txn)

	_, err := c.cmd.execute(ctx, args)
	return err
}

func (c *Collection) ListIndexes(
	ctx context.Context,
	opts ...options.Enumerable[options.ListCollectionIndexesOptions],
) ([]client.IndexDescription, error) {
	args := []string{"client", "index", "list"}
	args = append(args, "--collection", c.Version().Name)

	opt := utils.NewOptions(opts...)
	args = appendIdentityArg(args, opt.GetIdentity())
	args = appendTxnArg(args, c.txn)

	data, err := c.cmd.execute(ctx, args)
	if err != nil {
		return nil, err
	}
	var indexes []client.IndexDescription
	if err := json.Unmarshal(data, &indexes); err != nil {
		return nil, err
	}
	return indexes, nil
}

// NewEncryptedIndex implements client.Collection.
func (c *Collection) NewEncryptedIndex(
	ctx context.Context,
	indexDesc client.EncryptedIndexDescription,
	opts ...options.Enumerable[options.NewEncryptedIndexOptions],
) (index client.EncryptedIndexDescription, err error) {
	args := []string{"client", "encrypted-index", "new"}
	args = append(args, "--collection", c.Version().Name)
	args = append(args, "--field", indexDesc.FieldName)

	opt := utils.NewOptions(opts...)
	args = appendIdentityArg(args, opt.GetIdentity())
	args = appendTxnArg(args, c.txn)

	data, err := c.cmd.execute(ctx, args)
	if err != nil {
		return index, err
	}
	if err := json.Unmarshal(data, &index); err != nil {
		return index, err
	}
	return index, nil
}

// ListEncryptedIndexes implements client.Collection.
func (c *Collection) ListEncryptedIndexes(
	ctx context.Context, opts ...options.Enumerable[options.ListCollectionEncryptedIndexesOptions],
) ([]client.EncryptedIndexDescription, error) {
	args := []string{"client", "encrypted-index", "list"}
	args = append(args, "--collection", c.Version().Name)
	opt := utils.NewOptions(opts...)
	args = appendIdentityArg(args, opt.GetIdentity())
	args = appendTxnArg(args, c.txn)

	data, err := c.cmd.execute(ctx, args)
	if err != nil {
		return nil, err
	}
	var indexes []client.EncryptedIndexDescription
	if err := json.Unmarshal(data, &indexes); err != nil {
		return nil, err
	}
	return indexes, nil
}

// DeleteEncryptedIndex implements client.Collection.
func (c *Collection) DeleteEncryptedIndex(
	ctx context.Context,
	fieldName string,
	opts ...options.Enumerable[options.DeleteEncryptedIndexOptions],
) error {
	args := []string{"client", "encrypted-index", "delete"}
	args = append(args, "--collection", c.Version().Name)
	args = append(args, "--field", fieldName)

	opt := utils.NewOptions(opts...)
	args = appendIdentityArg(args, opt.GetIdentity())
	args = appendTxnArg(args, c.txn)

	_, err := c.cmd.execute(ctx, args)
	return err
}

func (c *Collection) Truncate(
	ctx context.Context, opts ...options.Enumerable[options.TruncateCollectionOptions],
) error {
	args := []string{"client", "collection", "truncate"}
	args = append(args, "--name", c.Version().Name)

	opt := utils.NewOptions(opts...)
	args = appendIdentityArg(args, opt.GetIdentity())
	args = appendTxnArg(args, c.txn)

	_, err := c.cmd.execute(ctx, args)
	return err
}
