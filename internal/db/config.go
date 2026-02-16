// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package db

import (
	"time"

	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/lens/host-go/engine/module"

	"github.com/sourcenetwork/defradb/client/options"
	intOpts "github.com/sourcenetwork/defradb/internal/options"
)

const (
	defaultMaxTxnRetries  = 5
	updateEventBufferSize = 100
)

func defaultDBConfig() intOpts.DBOptions {
	return intOpts.DBOptions{
		NodeDBOptions: options.NodeDBOptions{
			MaxTxnRetries: immutable.Some(defaultMaxTxnRetries),
			EnableSigning: true,
			RetryIntervals: []time.Duration{
				time.Second * 30,
				time.Minute,
				time.Minute * 2,
				time.Minute * 4,
				time.Minute * 8,
				time.Minute * 16,
				time.Minute * 32,
			},
			P2PBlockSyncTimeout: time.Second * 5,
		},
	}
}

type LensRuntimeType string

const (
	// The Go-enum default LensRuntimeType.
	//
	// The actual runtime type that this resolves to depends on the build target.
	DefaultLens LensRuntimeType = ""
)

// runtimeConstructors is a map of [LensRuntimeType]s to lens runtimes.
//
// Is is populated by the `init` functions in the runtime-specific files - this
// allows it's population to be managed by build flags.
var runtimeConstructors = map[LensRuntimeType]func() module.Runtime{}

func newLensRuntime(runtimeType LensRuntimeType) (module.Runtime, error) {
	if runtimeConstructor, ok := runtimeConstructors[runtimeType]; ok {
		return runtimeConstructor(), nil
	} else {
		return nil, NewErrLensRuntimeNotSupported(runtimeType)
	}
}
