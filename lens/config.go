// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package lens

import "github.com/lens-vm/lens/host-go/engine/module"

// Option is a funtion that sets a config value on the lens registry.
type Option func(*lensRegistry)

// WithPoolSize sets the maximum number of cached migrations instances to preserve per schema version.
//
// Will default to `5` if not set.
func WithPoolSize(size int) Option {
	return func(r *lensRegistry) {
		r.poolSize = size
	}
}

// WithRuntime returns an option that sets the lens registry runtime.
func WithRuntime(runtime module.Runtime) Option {
	return func(r *lensRegistry) {
		r.runtime = runtime
	}
}
