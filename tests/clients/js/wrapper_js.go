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

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/internal/db/txnctx"

	"github.com/sourcenetwork/goji"
)

func execute(ctx context.Context, value js.Value, method string, args ...any) ([]js.Value, error) {
	contextValues := map[string]any{}
	tx, ok := txnctx.TryGet(ctx)
	if ok {
		contextValues["transaction"] = tx.ID()
	}
	id := identity.FromContext(ctx)
	if id.HasValue() && id.Value().PrivateKey != nil {
		contextValues["identity"] = id.Value().PrivateKey.String()
	}
	args = append(args, contextValues)
	prom := value.Call(method, args...)
	return goji.Await(goji.PromiseValue(prom))
}
