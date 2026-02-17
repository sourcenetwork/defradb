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

import (
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp/dac"
	"github.com/sourcenetwork/defradb/client"
	publicOpts "github.com/sourcenetwork/defradb/client/options"
)

// DBOptions extends the public NodeDBOptions with internal-only fields.
type DBOptions struct {
	publicOpts.NodeDBOptions

	// DocumentACP is the optional document access control system.
	DocumentACP immutable.Option[dac.DocumentACP]
	// P2P is the optional P2P host for networking.
	P2P immutable.Option[client.Host]
}

// DBOptionsBuilder is a builder for DBOptions.
type DBOptionsBuilder struct {
	enumerableBuilder[DBOptions]
}

// DB creates a new standalone DBOptionsBuilder.
func DB() *DBOptionsBuilder {
	return &DBOptionsBuilder{}
}

// SetDocumentACP sets the document ACP system.
func (b *DBOptionsBuilder) SetDocumentACP(d dac.DocumentACP) *DBOptionsBuilder {
	b.Append(func(o *DBOptions) { o.DocumentACP = immutable.Some(d) })
	return b
}

// SetP2P sets the P2P host.
func (b *DBOptionsBuilder) SetP2P(host client.Host) *DBOptionsBuilder {
	b.Append(func(o *DBOptions) { o.P2P = immutable.Some[client.Host](host) })
	return b
}

// SetNodeDBOptions applies the public NodeDBOptions to the embedded struct.
func (b *DBOptionsBuilder) SetNodeDBOptions(cfg publicOpts.NodeDBOptions) *DBOptionsBuilder {
	b.Append(func(o *DBOptions) { o.NodeDBOptions = cfg })
	return b
}

// SetAll sets all DB options from a plain data struct.
func (b *DBOptionsBuilder) SetAll(opts DBOptions) *DBOptionsBuilder {
	b.Append(func(o *DBOptions) { *o = opts })
	return b
}
