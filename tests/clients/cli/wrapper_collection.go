// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cli

import (
	"context"
	"encoding/json"
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

var _ client.Collection = (*Collection)(nil)

type Collection struct {
	cmd *cliWrapper
	def client.CollectionVersion
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

func (c *Collection) AddIndex(
	ctx context.Context,
	indexDesc client.IndexAddRequest,
	opts ...options.Enumerable[options.CollectionAddIndexOptions],
) (index client.IndexDescription, err error) {
	args := []string{"client", "index", "add"}
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
	opts ...options.Enumerable[options.CollectionDeleteIndexOptions],
) error {
	args := []string{"client", "index", "delete"}
	args = append(args, "--collection", c.Version().Name)
	args = append(args, "--name", indexName)

	opt := utils.NewOptions(opts...)
	args = appendIdentityArg(args, opt.GetIdentity())

	_, err := c.cmd.execute(ctx, args)
	return err
}

func (c *Collection) ListIndexes(
	ctx context.Context,
	opts ...options.Enumerable[options.CollectionListIndexesOptions],
) ([]client.IndexDescription, error) {
	args := []string{"client", "index", "list"}
	args = append(args, "--collection", c.Version().Name)

	opt := utils.NewOptions(opts...)
	args = appendIdentityArg(args, opt.GetIdentity())

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

// AddEncryptedIndex implements client.Collection.
func (c *Collection) AddEncryptedIndex(
	ctx context.Context,
	indexDesc client.EncryptedIndexDescription,
	opts ...options.Enumerable[options.AddEncryptedIndexOptions],
) (index client.EncryptedIndexDescription, err error) {
	args := []string{"client", "encrypted-index", "add"}
	args = append(args, "--collection", c.Version().Name)
	args = append(args, "--field", indexDesc.FieldName)

	opt := utils.NewOptions(opts...)
	args = appendIdentityArg(args, opt.GetIdentity())

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
	ctx context.Context, opts ...options.Enumerable[options.CollectionListEncryptedIndexesOptions],
) ([]client.EncryptedIndexDescription, error) {
	args := []string{"client", "encrypted-index", "list"}
	args = append(args, "--collection", c.Version().Name)
	opt := utils.NewOptions(opts...)
	args = appendIdentityArg(args, opt.GetIdentity())

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

	_, err := c.cmd.execute(ctx, args)
	return err
}

func (c *Collection) Truncate(
	ctx context.Context, opts ...options.Enumerable[options.CollectionTruncateOptions],
) error {
	args := []string{"client", "collection", "truncate"}
	args = append(args, "--name", c.Version().Name)

	opt := utils.NewOptions(opts...)
	args = appendIdentityArg(args, opt.GetIdentity())

	_, err := c.cmd.execute(ctx, args)
	return err
}
