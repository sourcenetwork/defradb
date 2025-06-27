// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

//go:build js

package js

import (
	"context"
	"syscall/js"

	"github.com/sourcenetwork/goji"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/datastore"
)

func execute(ctx context.Context, value js.Value, method string, args ...any) ([]js.Value, error) {
	contextValues := map[string]any{}
	tx, ok := datastore.CtxTryGetClientTxn(ctx)
	if ok {
		contextValues["transaction"] = tx.ID()
	}
	id := identity.FromContext(ctx)
	if id.HasValue() {
		if full, ok := id.Value().(identity.FullIdentity); ok && full.PrivateKey() != nil {
			contextValues["identity"] = full.PrivateKey().String()
		}
	}
	args = append(args, contextValues)
	prom := value.Call(method, args...)
	return goji.Await(goji.PromiseValue(prom))
}
