// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package fetcher

import (
	"context"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp"
	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/db/permission"
)

// permissioned fetcher applies access control based filtering to documents fetched.
type permissioned struct {
	ctx context.Context

	identity immutable.Option[acpIdentity.Identity]
	acp      acp.ACP
	col      client.Collection

	fetcher fetcher
}

var _ fetcher = (*permissioned)(nil)

func newPermissionedFetcher(
	ctx context.Context,
	identity immutable.Option[acpIdentity.Identity],
	acp acp.ACP,
	col client.Collection,
	fetcher fetcher,
) *permissioned {
	return &permissioned{
		ctx:      ctx,
		identity: identity,
		acp:      acp,
		col:      col,
		fetcher:  fetcher,
	}
}

func (f *permissioned) NextDoc() (immutable.Option[string], error) {
	docID, err := f.fetcher.NextDoc()
	if err != nil {
		return immutable.None[string](), err
	}

	if !docID.HasValue() {
		return immutable.None[string](), nil
	}

	hasPermission, err := permission.CheckAccessOfDocOnCollectionWithACP(
		f.ctx,
		f.identity,
		f.acp,
		f.col,
		acp.ReadPermission,
		docID.Value(),
	)
	if err != nil {
		return immutable.None[string](), err
	}

	if !hasPermission {
		return f.NextDoc()
	}

	return docID, nil
}

func (f *permissioned) GetFields() (immutable.Option[EncodedDocument], error) {
	return f.fetcher.GetFields()
}

func (f *permissioned) Close() error {
	return f.fetcher.Close()
}
