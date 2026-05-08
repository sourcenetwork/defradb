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
	"encoding/hex"
	"fmt"
	"reflect"
	"sync"
	"syscall/js"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/sourcenetwork/goji"
	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/immutable/enumerable"

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/internal/db"
)

func stringArg(args []js.Value, index int, name string) (string, error) {
	if len(args) <= index {
		return "", fmt.Errorf("%s argument is required", name)
	}
	if args[index].Type() != js.TypeString {
		return "", fmt.Errorf("%s argument must be a string", name)
	}
	return args[index].String(), nil
}

func optionalStringArg(args []js.Value, index int) string {
	if len(args) <= index || args[index].Type() != js.TypeString {
		return ""
	}
	return args[index].String()
}

func boolArg(args []js.Value, index int, name string) (bool, error) {
	if len(args) <= index {
		return false, fmt.Errorf("%s argument is required", name)
	}
	if args[index].Type() != js.TypeBoolean {
		return false, fmt.Errorf("%s argument must be a bool", name)
	}
	return args[index].Bool(), nil
}

func structArg(args []js.Value, index int, name string, out any) error {
	if len(args) <= index {
		return fmt.Errorf("%s argument is required", name)
	}
	return goji.UnmarshalJS(args[index], out)
}

func stringSliceArg(args []js.Value, index int, name string) ([]string, error) {
	if len(args) <= index {
		return nil, fmt.Errorf("%s argument is required", name)
	}
	var out []string
	if err := goji.UnmarshalJS(args[index], &out); err != nil {
		return nil, fmt.Errorf("%s argument must be an array of strings: %w", name, err)
	}
	return out, nil
}

// optionsValue returns the JS options object at the given argument index,
// or js.Undefined() if not present or not an object. Options are always
// optional and always appear as the last argument.
func optionsValue(args []js.Value, index int) js.Value {
	if len(args) > index && args[index].Type() == js.TypeObject {
		return args[index]
	}
	return js.Undefined()
}

// makeContext builds the context for an operation from the JS options object.
//
// Only the transaction binding is read here. Identity is decoded into the
// per-operation options struct by parseOptions, which the DB layer copies
// into the context itself.
func makeContext(optsVal js.Value, txns *sync.Map) (context.Context, error) {
	ctx := context.Background()
	txn, err := optionsTransaction(optsVal, txns)
	if err != nil {
		return ctx, err
	}
	return db.InitContext(ctx, txn), nil
}

func optionsTransaction(opts js.Value, txns *sync.Map) (client.Txn, error) {
	if opts.IsUndefined() || opts.IsNull() {
		return nil, nil
	}
	id := opts.Get("transaction")
	if id.Type() != js.TypeNumber {
		return nil, nil
	}
	txn, ok := txns.Load(uint64(id.Int()))
	if !ok {
		return nil, ErrInvalidTransactionId
	}
	return txn.(client.Txn), nil //nolint:forcetypeassert
}

// optionsIdentity parses the `identity` (public key hex) or `fullIdentity`
// (private key hex) property out of the JS options object.
func optionsIdentity(opts js.Value) (immutable.Option[acpIdentity.Identity], error) {
	return parseIdentityKeys(opts, "fullIdentity", "identity")
}

// parseIdentityKeys reads a private/public key hex pair from the given JS
// options object using the supplied property names. The private key takes
// precedence when both are present.
func parseIdentityKeys(opts js.Value, fullKey, identityKey string) (immutable.Option[acpIdentity.Identity], error) {
	if opts.IsUndefined() || opts.IsNull() {
		return immutable.None[acpIdentity.Identity](), nil
	}
	full := opts.Get(fullKey)
	if full.Type() == js.TypeString {
		data, err := hex.DecodeString(full.String())
		if err != nil {
			return immutable.None[acpIdentity.Identity](), err
		}
		privKey := secp256k1.PrivKeyFromBytes(data)
		identity, err := acpIdentity.FromPrivateKey(crypto.NewPrivateKey(privKey))
		if err != nil {
			return immutable.None[acpIdentity.Identity](), err
		}
		return immutable.Some[acpIdentity.Identity](identity), nil
	}
	ident := opts.Get(identityKey)
	if ident.Type() != js.TypeString {
		return immutable.None[acpIdentity.Identity](), nil
	}
	publicKey, err := crypto.PublicKeyFromString(crypto.KeyTypeSecp256r1, ident.String())
	if err != nil {
		return immutable.None[acpIdentity.Identity](), err
	}
	identity, err := acpIdentity.FromPublicKey(publicKey)
	if err != nil {
		return immutable.None[acpIdentity.Identity](), err
	}
	return immutable.Some(identity), nil
}

// transientKeys are JS option keys that are handled out-of-band and stripped
// before JSON-decoding into typed options structs. Both camelCase (used by
// hand-written JS callers) and PascalCase (produced by Go-side struct
// marshalling in the test wrapper) variants are listed.
var transientKeys = []any{
	"identity", "Identity",
	"fullIdentity", "FullIdentity",
	"nodeIdentity", "NodeIdentity",
	"fullNodeIdentity", "FullNodeIdentity",
	"transaction", "Transaction",
}

// parseOptions decodes the JS options object directly into the typed options
// struct. Identity is parsed separately and assigned to the struct's `Identity`
// field via reflection (the field is an interface and so cannot be decoded by
// the JSON pipeline). The `transaction` key is consumed by makeContext.
//
// If optsVal is undefined the struct is left zero-valued.
func parseOptions[T any](optsVal js.Value, out *T) error {
	if optsVal.IsUndefined() || optsVal.IsNull() {
		return nil
	}
	// Copy the options object without identity/transaction keys, which need
	// special handling and would otherwise fail to decode.
	cleaned := js.Global().Get("Object").Call("assign", map[string]any{}, optsVal)
	for _, k := range transientKeys {
		cleaned.Delete(k.(string)) //nolint:forcetypeassert
	}
	if err := goji.UnmarshalJS(cleaned, out); err != nil {
		return err
	}
	ident, err := optionsIdentity(optsVal)
	if err != nil {
		return err
	}
	if ident.HasValue() {
		v := reflect.ValueOf(out).Elem().FieldByName("Identity")
		if v.IsValid() && v.CanSet() {
			v.Set(reflect.ValueOf(ident))
		}
	}
	return nil
}

// parseNodeOptions decodes the JS options object into the given NodeOptions
// struct, layering present fields on top of any values already set. The
// `nodeIdentity`/`fullNodeIdentity` keys are decoded separately and assigned
// to DB.Identity (the field is an interface and cannot be JSON-decoded).
func parseNodeOptions(optsVal js.Value, out *options.NodeOptions) error {
	if optsVal.IsUndefined() || optsVal.IsNull() {
		return nil
	}
	cleaned := js.Global().Get("Object").Call("assign", map[string]any{}, optsVal)
	for _, k := range transientKeys {
		cleaned.Delete(k.(string)) //nolint:forcetypeassert
	}
	if err := goji.UnmarshalJS(cleaned, out); err != nil {
		return err
	}
	ident, err := optionsIdentity(optsVal)
	if err != nil {
		return err
	}
	if ident.HasValue() {
		out.DB.Identity = ident
	}
	return nil
}

// asOpts wraps a parsed options struct as an Enumerable that, when applied,
// overwrites the target with the parsed value.
func asOpts[T any](v T) options.Enumerable[T] {
	return enumerable.New([]func(*T){
		func(target *T) { *target = v },
	})
}
