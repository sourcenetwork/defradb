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

// registernatives.c declares Java_source_defra_* functions as extern and takes their
// addresses for JNI's RegisterNatives, but those functions are implemented in
// cbindings/nativewrapper.c, not in this package. Without this import, nothing in this
// package's own dependency graph pulls cbindings' compiled code into the final binary.

import _ "github.com/sourcenetwork/defradb/cbindings"
