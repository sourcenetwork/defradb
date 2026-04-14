// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cache

import "context"

type Repository[T any, TV any] interface {
	TryGet(context.Context, T) (TV, bool, error)
	Write(context.Context, TV) error
	Delete(context.Context, T) error
	Forbid(TV)
}

type Cache[T any, TV any] interface {
	TryGet(T) (TV, bool)
	Cache(TV)
	Remove(T)
	Forbid(TV)
}
