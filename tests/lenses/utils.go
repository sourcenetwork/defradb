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
var SetDefaultModulePath string = getPathRelativeToProjectRoot(
	"/tests/modules/rust_wasm32_set_default/target/wasm32-unknown-unknown/debug/rust_wasm32_set_default.wasm",
)

func getPathRelativeToProjectRoot(relativePath string) string {
	_, filename, _, _ := runtime.Caller(0)
	root := path.Dir(path.Dir(path.Dir(filename)))
	return path.Join(root, relativePath)
}
