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

	"github.com/sourcenetwork/defradb/acp/dac"
	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	acpTypes "github.com/sourcenetwork/defradb/acp/types"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/db/permission"
)

// permissionedFetcher fetcher applies access control based filtering to documents fetched.
type permissionedFetcher struct {
	ctx context.Context

	identity    immutable.Option[acpIdentity.Identity]
	documentACP dac.DocumentACP
	col         client.Collection

	fetcher fetcher
}

var _ fetcher = (*permissionedFetcher)(nil)

func newPermissionedFetcher(
	ctx context.Context,
	identity immutable.Option[acpIdentity.Identity],
	documentACP dac.DocumentACP,
	col client.Collection,
	fetcher fetcher,
) *permissionedFetcher {
	return &permissionedFetcher{
		ctx:         ctx,
		identity:    identity,
		documentACP: documentACP,
		col:         col,
		fetcher:     fetcher,
	}
}

func (f *permissionedFetcher) NextDoc() (immutable.Option[string], error) {
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
		f.documentACP,
		f.col,
		acpTypes.DocumentReadPerm,
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

func (f *permissionedFetcher) GetFields() (immutable.Option[EncodedDocument], error) {
	return f.fetcher.GetFields()
}

func (f *permissionedFetcher) Close() error {
	return f.fetcher.Close()
}
