// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package lenses

import (
	"path"
	"runtime"
)

// SetDefaultModulePath is the path to the `SetDefault` lens module compiled to wasm.
//
// The module has two parameters:
//   - `dst` is a string and is the name of the property you wish to set
//   - `value` can be any valid json value and is the value that you wish the `dst` property
//     of all documents being transformed by this module to have.
//
// This module has an inverse, which will clear any value in the `dst` field.
var SetDefaultModulePath string = getPathRelativeToProjectRoot(
	"/tests/lenses/rust_wasm32_set_default/target/wasm32-unknown-unknown/debug/rust_wasm32_set_default.wasm",
)

// RemoveModulePath is the path to the `Remove` lens module compiled to wasm.
//
// The module has one parameter:
//   - `target` is a string and is the name of the property you wish to remove.
var RemoveModulePath string = getPathRelativeToProjectRoot(
	"/tests/lenses/rust_wasm32_remove/target/wasm32-unknown-unknown/debug/rust_wasm32_remove.wasm",
)

// CopyModulePath is the path to the `Copy` lens module compiled to wasm.
//
// The module has two parameters:
//   - `src` is a string and is the name of the property you wish to copy values from.
//   - `dst` is a string and is the name of the property you wish to copy the `src` value to.
var CopyModulePath string = getPathRelativeToProjectRoot(
	"/tests/lenses/rust_wasm32_copy/target/wasm32-unknown-unknown/debug/rust_wasm32_copy.wasm",
)

// PrependModulePath is the path to the `Prepend` lens module compiled to wasm.
//
// The module has one parameter:
//   - `values` is an array of `map[string]string`s, the module will yield these documents before
//     any documents fed to it (from Defra).
var PrependModulePath string = getPathRelativeToProjectRoot(
	"/tests/lenses/rust_wasm32_prepend/target/wasm32-unknown-unknown/debug/rust_wasm32_prepend.wasm",
)

// FilterModulePath is the path to the `Filter` lens module compiled to wasm.
//
// The module has two parameters:
//   - `src` is a string and is the name of the property you wish to evaluate
//   - `value` can be any valid json value and will be compared to the document value at the `src` location
//     only documents with values that match this given value will be returned.
var FilterModulePath string = getPathRelativeToProjectRoot(
	"/tests/lenses/rust_wasm32_filter/target/wasm32-unknown-unknown/debug/rust_wasm32_filter.wasm",
)

func getPathRelativeToProjectRoot(relativePath string) string {
	_, filename, _, _ := runtime.Caller(0)
	root := path.Dir(path.Dir(path.Dir(filename)))
	return path.Join(root, relativePath)
}
