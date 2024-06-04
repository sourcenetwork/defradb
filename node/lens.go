// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package node

import (
	"context"

	"github.com/lens-vm/lens/host-go/engine/module"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/lens"
)

type LensRuntimeType string

const (
	DefaultLens LensRuntimeType = ""
)

var runtimeConstructors = map[LensRuntimeType]func() module.Runtime{}

// LensOptions contains Lens configuration values.
type LensOptions struct {
	lensRuntime LensRuntimeType

	// The maximum number of cached migrations instances to preserve per schema version.
	lensPoolSize int
}

// DefaultACPOptions returns new options with default values.
func DefaultLensOptions() *LensOptions {
	return &LensOptions{
		lensPoolSize: lens.DefaultPoolSize,
	}
}

type LenOpt func(*LensOptions)

// WithLensRuntime returns an option that sets the lens registry runtime.
func WithLensRuntime(runtime LensRuntimeType) Option {
	return func(o *LensOptions) {
		o.lensRuntime = runtime
	}
}

// WithLensPoolSize sets the maximum number of cached migrations instances to preserve per schema version.
//
// Will default to `5` if not set.
func WithLensPoolSize(size int) Option {
	return func(o *LensOptions) {
		o.lensPoolSize = size
	}
}

func NewLens(
	ctx context.Context,
	opts ...LenOpt,
) (client.LensRegistry, error) {
	options := DefaultLensOptions()
	for _, opt := range opts {
		opt(options)
	}

	var runtime module.Runtime
	if runtimeConstructor, ok := runtimeConstructors[options.lensRuntime]; ok {
		runtime = runtimeConstructor()
	} else {
		return nil, NewErrLensRuntimeNotSupported(options.lensRuntime)
	}

	return lens.NewRegistry(
		options.lensPoolSize,
		runtime,
	), nil
}
