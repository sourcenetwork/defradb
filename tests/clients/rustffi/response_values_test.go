// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

//go:build rust_ffi

package rustffi

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/immutable"
)

func TestNormalizeCollectionDateTimesPreservesNullableElements(t *testing.T) {
	first := time.Date(2026, time.July, 30, 12, 0, 0, 0, time.UTC)
	second := first.Add(time.Hour)
	value := map[string]any{
		"times": []any{first, nil, second},
	}
	version := client.CollectionVersion{
		Fields: []client.CollectionFieldDescription{{
			Name: "times",
			Kind: client.FieldKind_NILLABLE_DATETIME_ARRAY,
		}},
	}

	normalizeCollectionDateTimes(value, version)

	require.Equal(t, []immutable.Option[time.Time]{
		immutable.Some(first),
		immutable.None[time.Time](),
		immutable.Some(second),
	}, value["times"])
}
