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
	"sync"
	"syscall/js"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/sourcenetwork/goji"
	"github.com/sourcenetwork/immutable"

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/internal/db"
	"github.com/sourcenetwork/defradb/internal/encryption"
)

func stringArg(args []js.Value, index int, name string) (string, error) {
	if len(args) < index {
		return "", fmt.Errorf("%s argument is required", name)
	}
	if args[index].Type() != js.TypeString {
		return "", fmt.Errorf("%s argument must be a string", name)
	}
	return args[index].String(), nil
}

func boolArg(args []js.Value, index int, name string) (bool, error) {
	if len(args) < index {
		return false, fmt.Errorf("%s argument is required", name)
	}
	if args[index].Type() != js.TypeBoolean {
		return false, fmt.Errorf("%s argument must be a bool", name)
	}
	return args[index].Bool(), nil
}

func intArg(args []js.Value, index int, name string) (int, error) {
	if len(args) < index {
		return 0, fmt.Errorf("%s argument is required", name)
	}
	if args[index].Type() != js.TypeBoolean {
		return 0, fmt.Errorf("%s argument must be an int", name)
	}
	return args[index].Int(), nil
}

func structArg(args []js.Value, index int, name string, out any) error {
	if len(args) < index {
		return fmt.Errorf("%s argument is required", name)
	}
	return goji.UnmarshalJS(args[index], out)
}

func contextArg(args []js.Value, index int, txns *sync.Map) (context.Context, error) {
	ctx := context.Background()
	if index >= len(args) {
		return ctx, nil
	}
	identity, err := contextIdentityArg(args[index])
	if err != nil {
		return ctx, err
	}
	txn, err := contextTransactionArg(args[index], txns)
	if err != nil {
		return ctx, err
	}
	encrypt, encryptFields := contextEncryptionArg(args[index])
	ctx = encryption.SetContextConfigFromParams(ctx, encrypt, encryptFields)
	ctx = acpIdentity.WithContext(ctx, identity)
	ctx = db.SetContextTxn(ctx, txn)
	return ctx, nil
}

func contextEncryptionArg(value js.Value) (bool, []string) {
	var encrypt bool
	if v := value.Get("encrypt"); v.Type() == js.TypeBoolean {
		encrypt = v.Bool()
	}
	var encryptFields []string
	if v := value.Get("encryptFields"); v.Type() == js.TypeObject {
		for i := 0; i < v.Length(); i++ {
			encryptFields = append(encryptFields, v.Index(i).String())
		}
	}
	return encrypt, encryptFields
}

func contextTransactionArg(value js.Value, txns *sync.Map) (datastore.Txn, error) {
	id := value.Get("transaction")
	if id.Type() != js.TypeNumber {
		return nil, nil
	}
	txn, ok := txns.Load(uint64(id.Int()))
	if !ok {
		return nil, ErrInvalidTransactionId
	}
	return txn.(datastore.Txn), nil //nolint:forcetypeassert
}

func contextIdentityArg(value js.Value) (immutable.Option[acpIdentity.Identity], error) {
	id := value.Get("identity")
	if id.Type() != js.TypeString {
		return immutable.None[acpIdentity.Identity](), nil
	}
	data, err := hex.DecodeString(id.String())
	if err != nil {
		return immutable.None[acpIdentity.Identity](), err
	}
	privKey := secp256k1.PrivKeyFromBytes(data)
	identity, err := acpIdentity.FromPrivateKey(privKey)
	if err != nil {
		return immutable.None[acpIdentity.Identity](), err
	}
	return immutable.Some(identity), nil
}
