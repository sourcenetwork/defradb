// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package options

// Lister is implemented by option builders to provide functional options.
type Lister[T any] interface {
	// List returns the slice of functional options that will be applied
	// to configure an options struct of type T.
	List() []func(*T)
}
