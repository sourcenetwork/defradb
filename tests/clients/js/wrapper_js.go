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

//go:build js

package js

import (
	"context"
	"syscall/js"

	"github.com/sourcenetwork/goji"

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/internal/datastore"
	iIdentity "github.com/sourcenetwork/defradb/internal/identity"
)

func execute(ctx context.Context, value js.Value, method string, args ...any) ([]js.Value, error) {
	contextValues := map[string]any{}
	tx, ok := datastore.CtxTryGetClientTxn(ctx)
	if ok {
		contextValues["transaction"] = tx.ID()
	}
	ident := iIdentity.FromContext(ctx)
	if ident.HasValue() {
		if full, ok := ident.Value().(acpIdentity.FullIdentity); ok && full.PrivateKey() != nil {
			contextValues["full_identity"] = full.PrivateKey().String()
		}
	}
	args = append(args, contextValues)
	prom := value.Call(method, args...)
	return goji.Await(goji.PromiseValue(prom))
}
