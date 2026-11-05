// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package hnsw

import "errors"

// ErrInvalidNodeEncoding is returned by UnmarshalNode when the input bytes
// are truncated, corrupt, or use an unsupported encoding version.
var ErrInvalidNodeEncoding = errors.New("hnsw: invalid node encoding")

// ErrInvalidMetaEncoding is returned by UnmarshalMeta when the input bytes
// are truncated, corrupt, or use an unsupported encoding version.
var ErrInvalidMetaEncoding = errors.New("hnsw: invalid meta encoding")

// ErrEntryPointNotFound is returned when the graph's meta points at an entry-point node that the
// store does not have. The graph is torn: meta and the node entries disagree.
var ErrEntryPointNotFound = errors.New("hnsw: entry point node not found")
