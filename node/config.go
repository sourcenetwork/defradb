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
	"github.com/sourcenetwork/defradb/internal/kms"

	"github.com/sourcenetwork/immutable"
)

// Option is a generic option that applies to any subsystem.
//
// Invalid option types will be silently ignored. Valid option types are:
// - `ACPOpt`
// - `NodeOpt`
// - `StoreOpt`
// - `db.Option`
// - `http.ServerOpt`
// - `net.NodeOpt`
type Option any

// Config contains node configuration values.
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

// NodeOpt is a function for setting configuration values.
type NodeOpt func(*Config)

// WithDisableP2P sets the disable p2p flag.
func WithDisableP2P(disable bool) NodeOpt {
	return func(o *Config) {
		o.disableP2P = disable
	}
}

// WithDisableAPI sets the disable api flag.
func WithDisableAPI(disable bool) NodeOpt {
	return func(o *Config) {
		o.disableAPI = disable
	}
}

func WithKMS(kms kms.ServiceType) NodeOpt {
	return func(o *Config) {
		o.kmsType = immutable.Some(kms)
	}
}

// WithEnableDevelopment sets the enable development mode flag.
func WithEnableDevelopment(enable bool) NodeOpt {
	return func(o *Config) {
		o.enableDevelopment = enable
	}
}

// filterOptions returns a list of options containing
// only options that match the given generic type.
func filterOptions[T any](options []Option) []T {
	var out []T
	for _, o := range options {
		switch t := o.(type) {
		case T:
			out = append(out, t)
		}
	}
	return out
}
