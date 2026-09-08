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
	"github.com/sourcenetwork/defradb/internal/utils"
)

const (
	identityFullProp = "fullIdentity"
	identityProp     = "identity"

	nodeIdentityFullProp = "fullNodeIdentity"
	nodeIdentityProp     = "nodeIdentity"
)

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

// filterArg returns the document filter at the given argument index, accepting either a
// GraphQL-style filter string or a plain filter conditions object. An object is decoded
// with the same number handling as UnmarshalJS. A missing/null/undefined object filter
// means "match all documents". Note that JS numbers are IEEE-754 doubles, so an integer
// filter condition above 2^53 will already have lost precision before reaching Go.
// To do: https://github.com/sourcenetwork/defradb/issues/5176
func filterArg(args []js.Value, index int, name string) (any, error) {
	if len(args) <= index {
		return nil, fmt.Errorf("%s argument is required", name)
	}
	arg := args[index]
	switch arg.Type() {
	case js.TypeString:
		return arg.String(), nil
	case js.TypeNull, js.TypeUndefined:
		return map[string]any{}, nil
	case js.TypeObject:
		filter, err := utils.DecodeJSONFilter([]byte(goji.JSON.Stringify(arg)))
		if err != nil {
			return nil, fmt.Errorf("%s argument is invalid: %w", name, err)
		}
		if filter == nil {
			filter = map[string]any{}
		}
		return filter, nil
	default:
		return nil, fmt.Errorf("%s argument must be a string or object", name)
	}
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

func optionsStore(db client.Store, optsVal js.Value, txns *sync.Map) (client.Store, error) {
	if optsVal.IsUndefined() || optsVal.IsNull() {
		return db, nil
	}
	id := optsVal.Get("transaction")
	if id.Type() != js.TypeNumber {
		return db, nil
	}
	txn, ok := txns.Load(uint64(id.Int()))
	if !ok {
		return nil, ErrInvalidTransactionId
	}
	return txn.(client.Txn), nil //nolint:forcetypeassert
}

// optionsIdentity parses an identity out of the JS options object from the
// given property names: fullProp holds a private key hex (full identity) and
// pubProp holds a public key hex. The private key takes precedence when both
// are present. The request identity uses `fullIdentity`/`identity` while the
// node identity uses `fullNodeIdentity`/`nodeIdentity`.
func optionsIdentity(opts js.Value, fullProp, pubProp string) (immutable.Option[acpIdentity.Identity], error) {
	if opts.IsUndefined() || opts.IsNull() {
		return immutable.None[acpIdentity.Identity](), nil
	}
	full := opts.Get(fullProp)
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
	ident := opts.Get(pubProp)
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
	ident, err := optionsIdentity(optsVal, identityFullProp, identityProp)
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
	ident, err := optionsIdentity(optsVal, nodeIdentityFullProp, nodeIdentityProp)
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
