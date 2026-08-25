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

//go:build javaclient

package java

/*
#include <stdlib.h>
#include <jni.h>
#include "../../../cbindings/defra_structs.h"
#include "jnicall.h"
#include "../../../cbindings/libdefradb.h"
*/
import "C"

import (
	"context"
	"errors"
	"fmt"
	"unsafe"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/datastore"
)

// This file contains the shared JNI call plumbing for the java test client.

// argBuilder will assemble an argument list for a single JNI method call.
// The purpose of it is to standardize how functions get called in a simple way, rather than deal with CGO/JNI
// details at every call point. Without it, every call would need to manually allocate and free the memory
// for the C strrings, and deal with populating the data inside DefraArgKind corrrectly. With this, it is
// done with one chainable API. For example:
//
// newArgs().argStr(patch).argStr(migrationStr).argLong(idH)
type argBuilder struct {
	args  []C.DefraArg // The arguments for the function call
	cstrs []*C.char    // The C strings that will be tracked to be freed afterwarrd
}

func newArgs() *argBuilder {
	return &argBuilder{}
}

// allocCString is a helper function, that allocates memory for a new C String, and appends it to
// the builder, so that it will get cleaned up later.
func (b *argBuilder) allocCString(s string) *C.char {
	cs := C.CString(s)
	b.cstrs = append(b.cstrs, cs)
	return cs
}

// freeCStrings will be called at the end to free the memory allocated by calls to allocCString
func (b *argBuilder) freeCStrings() {
	for _, cs := range b.cstrs {
		C.free(unsafe.Pointer(cs))
	}
}

// argBool appends a bool argument.
func (b *argBuilder) argBool(v bool) *argBuilder {
	i := C.int(0)
	if v {
		i = 1
	}
	b.args = append(b.args, C.DefraArg{kind: C.DEFRA_ARG_BOOL, i: i})
	return b
}

// argInt appends an int argument.
func (b *argBuilder) argInt(v int) *argBuilder {
	b.args = append(b.args, C.DefraArg{kind: C.DEFRA_ARG_INT, i: C.int(v)})
	return b
}

// argLong appends a long argument.
func (b *argBuilder) argLong(v uintptr) *argBuilder {
	b.args = append(b.args, C.DefraArg{kind: C.DEFRA_ARG_LONG, j: C.jlong(v)})
	return b
}

// argStr appends a String argument.
func (b *argBuilder) argStr(v string) *argBuilder {
	b.args = append(b.args, C.DefraArg{kind: C.DEFRA_ARG_STRING, str: b.allocCString(v)})
	return b
}

// collOpts appends a DefraCollectionOptions argument, built on the C side from the given fields.
// enableSigning overrides the node-level signing setting for this operation; None leaves it unset.
func (b *argBuilder) collOpts(
	name, version, collectionID string,
	getInactive bool,
	enableSigning immutable.Option[bool],
) *argBuilder {
	gi := C.int(0)
	if getInactive {
		gi = 1
	}
	es := C.int(0)
	if enableSigning.HasValue() {
		if enableSigning.Value() {
			es = 1
		} else {
			es = -1
		}
	}
	b.args = append(b.args, C.DefraArg{
		kind:            C.DEFRA_ARG_COLLECTION_OPTIONS,
		str:             b.allocCString(name),
		coVersion:       b.allocCString(version),
		coCollectionID:  b.allocCString(collectionID),
		coGetInactive:   gi,
		coEnableSigning: es,
	})
	return b
}

// CGO has no way of passing a Go slice directly as a C array, so we need a way to pass it. We can
// take advantage of the fact that in Go, slices are contiguous in memory. So we can return a pointer
// to the first element in the slice, and the number off elements in the slice.

// argvPtr is the function that does this.
func (b *argBuilder) argvPtr() (*C.DefraArg, C.int) {
	if len(b.args) == 0 {
		return nil, 0
	}
	return &b.args[0], C.int(len(b.args))
}

// defraResult is the parsed shape of a Java-side DefraResult object.
type defraResult struct {
	Status int
	Error  string
	Value  string
}

// goStringField reads a java.lang.String field off a Java object and converts it into a Go string
func goStringField(env *C.JNIEnv, obj C.jobject, fid C.jfieldID) string {
	cstr := C.defra_get_string_field_copy(env, obj, fid)
	if cstr == nil {
		return ""
	}
	defer C.free(unsafe.Pointer(cstr))
	return C.GoString(cstr)
}

// callNode invokes the named registered DefraNode native method (handle is
// either a node's or a transaction's cgo.Handle, matching cbindings'
// getNodeOrTxnHandle convention, and is passed as the method's first
// argument) and parses the returned DefraResult.
func callNode(nodeObj C.jobject, name string, handle uintptr, b *argBuilder) (defraResult, error) {
	full := newArgs()
	full.args = append(full.args, C.DefraArg{kind: C.DEFRA_ARG_LONG, j: C.jlong(handle)})
	full.args = append(full.args, b.args...)
	full.cstrs = b.cstrs
	b.cstrs = nil // ownership moved to full, avoid double free
	return callNodeRaw(nodeObj, name, full)
}

// callNodeNoHandle is like callNode but for the two DefraNode native methods
// that take no handle argument at all (such as PollSubscriptionNative)
func callNodeNoHandle(nodeObj C.jobject, name string, b *argBuilder) (defraResult, error) {
	return callNodeRaw(nodeObj, name, b)
}

// callNodeRaw looks up the named native method, attaches to the JVM, invokes
// it with the builder's arguments, and parses the returned DefraResult.
func callNodeRaw(nodeObj C.jobject, name string, b *argBuilder) (defraResult, error) {
	defer b.freeCStrings()

	// Look up the jmethodID for the requested native method name, cached in
	// the table built once at JVM setup (see: jvm.go).
	mid, ok := nodeMethodIDs[name]
	if !ok {
		return defraResult{}, fmt.Errorf(errFmtNoRegisteredMethod, name)
	}

	// Pin this goroutine to its OS thread.
	// Doing this prevents a SIGSEGV crash, because a JNIEnv is only valid on the thread that obtained it.
	env, detach, err := attach()
	if err != nil {
		return defraResult{}, err
	}
	defer detach()

	// Turn the builder's []C.DefraArg into the (*C.DefraArg, count) pair, and make the JNI call.
	argv, n := b.argvPtr()
	var errbuf [C.DEFRA_ERRBUF_LEN]C.char
	obj := C.defra_call_object_method(env, nodeObj, mid, argv, n, &errbuf[0], C.int(len(errbuf)))
	if obj == 0 {
		// A null resullt means a Java exception was thrown. But, we can examine what it was.
		return defraResult{}, fmt.Errorf(errFmtJNICall, name, C.GoString(&errbuf[0]))
	}

	return defraResult{
		Status: int(C.defra_get_int_field(env, obj, resultStatusField)),
		Error:  goStringField(env, obj, resultErrorField),
		Value:  goStringField(env, obj, resultValueField),
	}, nil
}

// callTxn invokes the named registered DefraTransaction native method with a given name on a
// transaction object. The transaction handle is passed explicitly.
func callTxn(txnObj C.jobject, name string, handle uintptr, b *argBuilder) (defraResult, error) {
	full := newArgs()
	full.args = append(full.args, C.DefraArg{kind: C.DEFRA_ARG_LONG, j: C.jlong(handle)})
	full.args = append(full.args, b.args...)
	full.cstrs = b.cstrs
	b.cstrs = nil // ownership moved to full, avoid double free
	return callTxnRaw(txnObj, name, full)
}

// callTxnRaw is callNodeRaw's counterpart for DefraTransaction natives
func callTxnRaw(txnObj C.jobject, name string, b *argBuilder) (defraResult, error) {
	defer b.freeCStrings()

	// Firrst, get the transaction method ID
	mid, ok := transactionMethodIDs[name]
	if !ok {
		return defraResult{}, fmt.Errorf(errFmtNoRegisteredMethod, name)
	}

	// Pin this goroutine to its OS thread.
	// Doing this prevents a SIGSEGV crash, because a JNIEnv is only valid on the thread that obtained it.
	env, detach, err := attach()
	if err != nil {
		return defraResult{}, err
	}
	defer detach()

	// Turn the builder's []C.DefraArg into the (*C.DefraArg, count) pair, and make the JNI call.
	argv, n := b.argvPtr()
	var errbuf [C.DEFRA_ERRBUF_LEN]C.char
	obj := C.defra_call_object_method(env, txnObj, mid, argv, n, &errbuf[0], C.int(len(errbuf)))
	if obj == 0 {
		return defraResult{}, fmt.Errorf(errFmtJNICall, name, C.GoString(&errbuf[0]))
	}

	return defraResult{
		Status: int(C.defra_get_int_field(env, obj, resultStatusField)),
		Error:  goStringField(env, obj, resultErrorField),
		Value:  goStringField(env, obj, resultValueField),
	}, nil
}

// callStore invokes the named native method appropriately based on context. If ctx carries an
// active transaction (a *Txn) AND DefraTransaction actually has this native method registered, it
// gets invoked on the transaction's own Java object. Otherwise, it gets invoked on the node.
//
// Guards against Close having deleted w.nodeObj's JNI global ref out from under it: nodeMu is held
// for the whole call, so this can't run concurrently with Close deleting the ref, and w.closed is
// checked so a call that arrives after Close has already finished returns a clean error instead of
// using a stale/deleted nodeObj.
//
// Also guards against a ctx that still carries a *Txn whose Commit/Discard has already run. That
// txn's handle/txnObj are zeroed on finalization (see Txn.finalize), so without this check a call
// made through a finished transaction would either look up a stale native handle (already reused
// or deleted, since 0 is never a valid cgo.Handle) or silently fall back to operating on the whole
// node instead of failing clearly.
func callStore(w *Wrapper, ctx context.Context, name string, b *argBuilder) (defraResult, error) {
	w.nodeMu.RLock()
	defer w.nodeMu.RUnlock()
	if w.closed {
		return defraResult{}, errors.New(errWrapperClosed)
	}

	if activeTxn, hadTxn := datastore.CtxTryGetTxn(ctx); hadTxn {
		if t, ok := activeTxn.(*Txn); ok {
			if t.isFinalized() {
				return defraResult{}, client.ErrTransactionNotFound
			}
			if _, hasMethod := transactionMethodIDs[name]; hasMethod {
				return callTxn(t.txnObj, name, t.handle, b)
			}
		}
	}
	handle := getNodeOrTxnHandle(w.handle, ctx)
	return callNode(w.nodeObj, name, handle, b)
}

// callGuarded invokes a DefraNode native method directly on this Wrapper's own node object,
// bypassing callStore's transaction dispatch. Guards against Close the same way callStore does.
func (w *Wrapper) callGuarded(name string, handle uintptr, b *argBuilder) (defraResult, error) {
	w.nodeMu.RLock()
	defer w.nodeMu.RUnlock()
	if w.closed {
		return defraResult{}, errors.New(errWrapperClosed)
	}
	return callNode(w.nodeObj, name, handle, b)
}

// asError converts a defraResult with a non-zero status into a Go error, reviving
// well-known DefraDB sentinel errors so callers can use errors.Is against them.
func (r defraResult) asError() error {
	if r.Status == 0 {
		return nil
	}
	return client.ReviveError(r.Error)
}

// createTransactionWithHandle creates a new txn using the native method, retuurning the new txn's cgo.Handle value.
func createTransactionWithHandle(nodeObj C.jobject, nodePtr uintptr, isReadOnly bool) (uintptr, error) {
	mid, ok := nodeMethodIDs["TransactionCreateNative"]
	if !ok {
		return 0, errors.New(errNoRegisteredTxnCreateMethod)
	}

	// Pin this goroutine to its OS thread.
	// Doing this prevents a SIGSEGV crash, because a JNIEnv is only valid on the thread that obtained it.
	env, detach, err := attach()
	if err != nil {
		return 0, err
	}
	defer detach()

	// Create and populate the DefraArg object for use in the JNI call
	readOnlyCFlag := C.int(0)
	if isReadOnly {
		readOnlyCFlag = 1
	}
	args := []C.DefraArg{
		{kind: C.DEFRA_ARG_LONG, j: C.jlong(nodePtr)},
		{kind: C.DEFRA_ARG_BOOL, i: readOnlyCFlag},
	}

	// Make the JNI call
	var errbuf [C.DEFRA_ERRBUF_LEN]C.char
	obj := C.defra_call_object_method(env, nodeObj, mid, &args[0], C.int(len(args)), &errbuf[0], C.int(len(errbuf)))
	if obj == 0 {
		// A null resullt means a Java exception was thrown. But, we can examine what it was.
		return 0, fmt.Errorf(errFmtTransactionCreateFailed, C.GoString(&errbuf[0]))
	}

	// Examine the status, and if it worked return the pointer, otherwise return 0 and an error
	status := int(C.defra_get_int_field(env, obj, txnResultStatusField))
	if status != 0 {
		return 0, client.ReviveError(goStringField(env, obj, txnResultErrorField))
	}
	ptr := uintptr(C.defra_get_long_field(env, obj, txnResultPtrField))
	return ptr, nil
}

// commitTransaction commits a transaction by way of the exposed native method
func commitTransaction(txnObj C.jobject, txnPtr uintptr) error {
	// Pin this goroutine to its OS thread.
	// Doing this prevents a SIGSEGV crash, because a JNIEnv is only valid on the thread that obtained it.
	env, detach, err := attach()
	if err != nil {
		return err
	}
	defer detach()

	// Look up the method ID, and use it to make the method call
	mid := transactionMethodIDs["TransactionCommitNative"]
	args := []C.DefraArg{{kind: C.DEFRA_ARG_LONG, j: C.jlong(txnPtr)}}
	var errbuf [C.DEFRA_ERRBUF_LEN]C.char
	obj := C.defra_call_object_method(env, txnObj, mid, &args[0], 1, &errbuf[0], C.int(len(errbuf)))
	if obj == 0 {
		// A null resullt means a Java exception was thrown. But, we can examine what it was.
		return fmt.Errorf(errFmtTransactionCommitFailed, C.GoString(&errbuf[0]))
	}
	if status := int(C.defra_get_int_field(env, obj, resultStatusField)); status != 0 {
		// In this case, there was not a Java exception, but the function call returned an error
		return client.ReviveError(goStringField(env, obj, resultErrorField))
	}
	return nil
}

// discardTransaction discards a transaction by way of the exposed native method
func discardTransaction(txnObj C.jobject, txnPtr uintptr) {
	// Pin this goroutine to its OS thread.
	// Doing this prevents a SIGSEGV crash, because a JNIEnv is only valid on the thread that obtained it.
	env, detach, err := attach()
	if err != nil {
		return
	}
	defer detach()

	// Look up the method ID, and use it to make the method call
	mid := transactionMethodIDs["TransactionDiscardNative"]
	args := []C.DefraArg{{kind: C.DEFRA_ARG_LONG, j: C.jlong(txnPtr)}}
	var errbuf [C.DEFRA_ERRBUF_LEN]C.char
	C.defra_call_void_method(env, txnObj, mid, &args[0], 1, &errbuf[0], C.int(len(errbuf)))
}
