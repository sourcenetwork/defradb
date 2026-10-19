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
	"github.com/sourcenetwork/immutable"

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/utils"
)

// execute calls a method on the JS client.
//
// The final element of args, if it is a js.Value object, is treated as the
// options object for the call. The active transaction (if any) is set on
// that options object before invocation. If no options object is supplied a
// new empty one is appended.
func execute(ctx context.Context, value js.Value, method string, args ...any) ([]js.Value, error) {
	var opts js.Value
	if len(args) > 0 {
		if v, ok := args[len(args)-1].(js.Value); ok && v.Type() == js.TypeObject {
			opts = v
			args = args[:len(args)-1]
		}
	}
	if opts.IsUndefined() {
		opts = js.Global().Get("Object").New()
	}
	if tx, ok := datastore.CtxTryGetClientTxn(ctx); ok {
		opts.Set("transaction", tx.ID())
	}
	args = append(args, opts)
	prom := value.Call(method, args...)
	return goji.Await(goji.PromiseValue(prom))
}

// identityHolder is satisfied by all option pointer types in the options
// package. It is the interface used to extract identity for the JS bridge.
type identityHolder interface {
	GetIdentity() immutable.Option[acpIdentity.Identity]
}

// jsOpts merges the option enumerables into a single struct and serialises
// it as a JS options object. The Identity interface field is replaced with
// a hex-encoded `fullIdentity`/`identity` key the JS bridge can decode.
func jsOpts[T any](opts []options.Enumerable[T]) js.Value {
	opt := utils.NewOptions(opts...)
	jsObj := goji.MustMarshalJS(opt)
	jsObj.Delete("Identity")
	if h, ok := any(opt).(identityHolder); ok {
		ident := h.GetIdentity()
		if ident.HasValue() {
			if full, ok := ident.Value().(acpIdentity.FullIdentity); ok && full.PrivateKey() != nil {
				jsObj.Set("fullIdentity", full.PrivateKey().String())
			} else if pub := ident.Value().PublicKey(); pub != nil {
				jsObj.Set("identity", pub.String())
			}
		}
	}
	return jsObj
}
