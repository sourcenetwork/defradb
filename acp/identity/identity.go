// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

/*
Package identity provides defradb identity.
*/

package identity

import (
	"github.com/sourcenetwork/immutable"
)

// Identity is the unique identifier for an actor.
type Identity immutable.Option[string]

var (
	// NoIdentity is an empty identity.
	NoIdentity = Identity(immutable.None[string]())
)

// NewIdentity makes a new identity if the input is not empty otherwise, returns an empty Option.
func NewIdentity(identity string) Identity {
	// TODO-ACP: There will be more validation once sourcehub gets some utilities.
	// Then a validation function would do the validation, will likely do outside this function.
	// https://github.com/sourcenetwork/defradb/issues/2358
	if identity == "" {
		return NoIdentity
	}
	return Identity(immutable.Some[string](identity))
}

// HasValue returns true if the identity is not empty.
func (i Identity) HasValue() bool {
	return immutable.Option[string](i).HasValue()
}

// Value returns the identity value.
func (i Identity) Value() string {
	return immutable.Option[string](i).Value()
}
