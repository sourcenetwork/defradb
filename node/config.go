// Copyright 2025 Democratized Data Foundation
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
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/internal/kms"

	"github.com/sourcenetwork/immutable"
)

const (
	// 1 MB, this matches the maximum badger-in-memory value size.
	//
	// Nearly at least, badger panics if this is set to it's max for reasons not yet
	// looked into.  Going one byte smaller does not have this issue.
	defaultChunkSize = (1 << 20) - 1
)

// Config contains internal node configuration values derived from options.
type Config struct {
	disableP2P        bool
	disableAPI        bool
	enableDevelopment bool
	kmsType           immutable.Option[kms.ServiceType]
}

// DefaultConfig returns a Config with default settings.
func DefaultConfig() *Config {
	return &Config{}
}

// applyNodeOptions applies NodeOptions to the config.
func (c *Config) applyNodeOptions(opts *options.NodeOptions) {
	if opts == nil {
		return
	}
	c.disableP2P = opts.DisableP2P
	c.disableAPI = opts.DisableAPI
	c.enableDevelopment = opts.EnableDevelopment
	if opts.KMSType.HasValue() {
		c.kmsType = immutable.Some(kms.ServiceType(opts.KMSType.Value()))
	}
}
