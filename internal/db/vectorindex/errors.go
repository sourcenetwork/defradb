// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package vectorindex

import (
	"github.com/sourcenetwork/defradb/errors"
)

const (
	errVectorIndexStore string = "vector index store operation failed"
)

// newErrVectorIndexStore returns a new error indicating that a vector index store operation
// (against its datastore-backed node/meta keyspace) failed.
func newErrVectorIndexStore(inner error) error {
	return errors.Wrap(errVectorIndexStore, inner)
}
