// Copyright 2026 Democratized Data Foundation
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
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// A request whose fragments spread each other in a cycle fails validation. Validation covers
// every fragment the request defines, so a cycle that no field spreads fails too.
func TestCyclicFragmentSpreadsFailValidation(t *testing.T) {
	ctx := context.Background()
	db, _ := setupUserCollection(t, ctx)

	for _, tc := range []struct {
		name    string
		request string
		wantErr string
	}{
		{
			name:    "cycle spread from a field",
			request: `query { User { ...A } } fragment A on User { name ...B } fragment B on User { name ...A }`,
			wantErr: `Cannot spread fragment "A" within itself via B.`,
		},
		{
			name:    "cycle no field spreads",
			request: `query { User { name } } fragment A on User { name ...B } fragment B on User { name ...A }`,
			wantErr: `Cannot spread fragment "A" within itself via B.`,
		},
		{
			name:    "fragment spreads itself",
			request: `query { User { ...A } } fragment A on User { name ...A }`,
			wantErr: `Cannot spread fragment "A" within itself.`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			res := db.ExecRequest(ctx, tc.request)

			messages := make([]string, len(res.GQL.Errors))
			for i, err := range res.GQL.Errors {
				messages[i] = err.Error()
			}
			require.Contains(t, messages, tc.wantErr)
			require.Nil(t, res.GQL.Data)
		})
	}
}
