// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package p2p

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/errors"
)

// The reason a DAG sync reports is derived from its error, so every step's error has to
// name itself.
func TestSyncDAGReason_NamesTheStepThatFailed(t *testing.T) {
	cause := errors.New("cause")

	for _, tc := range []struct {
		name string
		err  error
		want string
	}{
		{name: "storing the root block", err: NewErrStoreBlockDAGSync(cause), want: reasonStoreRoot},
		{name: "generating a block link", err: NewErrGenerateBlockLink(cause), want: reasonBlockLink},
		{name: "checking whether a block merged", err: NewErrCheckBlockMerged(cause), want: reasonIsMerged},
		{name: "verifying a signature", err: NewErrVerifyBlockSig(cause), want: reasonVerifySig},
		{name: "fetching encryption keys", err: NewErrGetEncKeysForBlock(cause), want: reasonEncKeys},
		{name: "retrieving one encryption key", err: NewErrRetrieveEncKey(cause), want: reasonEncKeys},
		{name: "loading a linked block", err: NewErrLoadLinkedBlock(cause), want: reasonLoadLink},
		{name: "decoding a linked block", err: NewErrDecodeLinkedBlock(cause), want: reasonDecodeLink},
		{name: "the context ending", err: context.Canceled, want: reasonContext},
		{name: "a deadline passing", err: context.DeadlineExceeded, want: reasonContext},
		{name: "an error no step names", err: cause, want: reasonOther},
		{
			// The step is more useful than the cancellation that stopped it.
			name: "a load stopped by a cancelled context",
			err:  NewErrLoadLinkedBlock(context.Canceled),
			want: reasonLoadLink,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, syncDAGReason(tc.err))
		})
	}
}
