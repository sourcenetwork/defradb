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

//go:build rust_ffi

// Package rustffi runs the integration tests against the Rust implementation
// of DefraDB, through its C API.
//
// The Rust C API mirrors Go's (cbindings), call for call, so this client
// mirrors cbindings.CWrapper. Run it through defradb.rs's tools/ffi-test, which
// builds the Rust library, generates defra.h, copies both next to this file,
// and runs the tests with -tags rust_ffi and DEFRA_CLIENT_RUST_FFI=true.
package rustffi

/*
#cgo CFLAGS: -I${SRCDIR}
#cgo LDFLAGS: -L${SRCDIR} -ldefra_ffi -Wl,-rpath,${SRCDIR}

#include <stdlib.h>
#include "defra.h"
*/
import "C"

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"unsafe"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/internal/datastore"
)

// subscriptionPending is the status of a poll that found nothing yet, and of
// a call that started a subscription.
const subscriptionPending = 2

// result is a C result read into Go memory.
type result struct {
	Status int
	Error  string
	Value  string
}

// err is the result's error, or nil when it succeeded.
func (r result) err() error {
	if r.Status == 1 {
		return errors.New(r.Error)
	}
	return nil
}

// take reads a C result into Go memory and frees it.
func take(cResult C.struct_DefraResult) result {
	res := result{
		Status: int(cResult.status),
		Error:  C.GoString(cResult.error),
		Value:  C.GoString(cResult.value),
	}
	C.defra_result_free(&cResult)
	return res
}

// unmarshalResult reads a call's JSON value into a T.
func unmarshalResult[T any](res result) (T, error) {
	var value T
	if err := res.err(); err != nil {
		return value, err
	}
	if err := json.Unmarshal([]byte(res.Value), &value); err != nil {
		return value, fmt.Errorf("failed to unmarshal JSON into %T: %w", value, err)
	}
	return value, nil
}

// cstrings holds the C strings one call passes, freed together after it.
type cstrings []unsafe.Pointer

// str allocates s as a C string.
func (a *cstrings) str(s string) *C.char {
	c := C.CString(s)
	*a = append(*a, unsafe.Pointer(c))
	return c
}

// bytes allocates b as C memory; nil for no bytes.
func (a *cstrings) bytes(b []byte) *C.uint8_t {
	if len(b) == 0 {
		return nil
	}
	c := C.CBytes(b)
	*a = append(*a, c)
	return (*C.uint8_t)(c)
}

// free releases every allocation.
func (a *cstrings) free() {
	for _, p := range *a {
		C.free(p)
	}
}

// identityHandle registers ident with the Rust library for one call, as
// cbindings registers a cgo handle; the caller frees it with freeIdentity.
func identityHandle(ident immutable.Option[identity.Identity]) (C.uintptr_t, error) {
	if !ident.HasValue() || ident.Value() == nil {
		return 0, nil
	}
	return newIdentityHandle(ident.Value())
}

// newIdentityHandle registers ident with the Rust library: by its private
// key when the test holds one, by its bearer token otherwise.
func newIdentityHandle(ident identity.Identity) (C.uintptr_t, error) {
	var args cstrings
	defer args.free()

	var cResult C.struct_DefraNewIdentityResult
	switch typed := ident.(type) {
	case identity.FullIdentity:
		key := typed.PrivateKey()
		cResult = C.defra_identity_new_from_typed_private_key(
			args.str(string(key.Type())), args.str(hex.EncodeToString(key.Raw())))
	case identity.TokenIdentity:
		cResult = C.defra_identity_new_from_bearer_token(
			args.str(typed.DID()), args.str(typed.BearerToken()))
	default:
		return 0, fmt.Errorf("identity %s has neither a private key nor a bearer token", ident.DID())
	}

	handle := cResult.identity_ptr
	message := C.GoString(cResult.error)
	failed := cResult.status != 0
	C.defra_new_identity_result_free(&cResult)
	if failed {
		return 0, errors.New(message)
	}
	return handle, nil
}

// freeIdentity releases a handle identityHandle returned.
func freeIdentity(handle C.uintptr_t) {
	if handle != 0 {
		C.defra_identity_free(handle)
	}
}

// collectionOptions selects a collection, as C.CollectionOptions does.
type collectionOptions struct {
	Version      string
	CollectionID string
	Name         string
	GetInactive  bool
	Signing      immutable.Option[bool]
}

// c lays the options out for C, its strings in args.
func (o collectionOptions) c(args *cstrings) C.struct_CollectionOptions {
	var copts C.struct_CollectionOptions
	copts.version = args.str(o.Version)
	copts.collection_id = args.str(o.CollectionID)
	copts.name = args.str(o.Name)
	if o.GetInactive {
		copts.get_inactive = 1
	}
	if o.Signing.HasValue() {
		if o.Signing.Value() {
			copts.enable_signing = 1
		} else {
			copts.enable_signing = -1
		}
	}
	return copts
}

// getCollectionsOptions reads Go's collection lookup options.
func getCollectionsOptions(opts *options.GetCollectionsOptions) collectionOptions {
	var copts collectionOptions
	if opts == nil {
		return copts
	}
	if opts.CollectionName.HasValue() {
		copts.Name = opts.CollectionName.Value()
	}
	if opts.VersionID.HasValue() {
		copts.Version = opts.VersionID.Value()
	}
	if opts.CollectionID.HasValue() {
		copts.CollectionID = opts.CollectionID.Value()
	}
	copts.GetInactive = opts.GetInactive.HasValue() && opts.GetInactive.Value()
	return copts
}

// target is the handle a call runs on: the transaction ctx carries, or the
// node.
func (w *Wrapper) target(ctx context.Context) C.uintptr_t {
	if txn, ok := datastore.CtxTryGetClientTxn(ctx); ok {
		if rustTxn, ok := txn.(*Transaction); ok {
			return rustTxn.handle
		}
	}
	return w.node
}

// requestStrings reads a request's operation name and variables as C takes
// them: empty when absent.
func requestStrings(opt *options.ExecRequestOptions) (string, string, error) {
	operation := ""
	if opt.OperationName.HasValue() {
		operation = opt.OperationName.Value()
	}
	variables := ""
	if opt.Variables != nil {
		data, err := json.Marshal(opt.Variables)
		if err != nil {
			return "", "", err
		}
		variables = string(data)
	}
	return operation, variables, nil
}

// optionJSON is the option's value as JSON; empty when it has none.
func optionJSON[T any](opt immutable.Option[T]) (string, error) {
	if !opt.HasValue() {
		return "", nil
	}
	data, err := json.Marshal(opt.Value())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// gqlResult reads a GraphQL result, or the call's error as one.
func gqlResult(res result) client.GQLResult {
	var gql client.GQLResult
	if err := res.err(); err != nil {
		gql.Errors = append(gql.Errors, err)
		return gql
	}
	if err := json.Unmarshal([]byte(res.Value), &gql); err != nil {
		gql.Errors = append(gql.Errors, err)
	}
	return gql
}
