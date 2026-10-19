// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package keys

// IndexStateKey identifies an index by its collection and index ID. The db package uses it to
// key index state in memory, for example during startup recovery; the persisted state lives in
// an action record (see ActionStatusKey).
//
// CollectionID is the stable collection ID (not a version ID), so state survives collection
// version switches.
type IndexStateKey struct {
	CollectionID string
	IndexID      uint32
}
