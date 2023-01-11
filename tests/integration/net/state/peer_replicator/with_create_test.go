// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package peer_replicator_test

/*
import (
	"testing"

	"github.com/sourcenetwork/defradb/config"
	testUtils "github.com/sourcenetwork/defradb/tests/integration/net/state"
)

// This test fails and should be uncommented once the behaviour is corrected
// The mode of failure is somewhat flaky, possibly due to the test framework
// however I do not believe the framework should accomodate this (the bug
// looks to be a production issue).  Likely part of:
// https://github.com/sourcenetwork/defradb/issues/1000
//
// It should not fail, and should pass as is however the result toggles between
// only 1 (out of 2 expected) documents existing in node 2, and the second document
// having a nil value (really bizare, and suggests a horrible bug in collection.Get(key),
// or collection.GetAllDocKeys).
func TestP2PPeerReplicatorWithCreate(t *testing.T) {
	test := testUtils.P2PTestCase{
		NodeConfig: []*config.Config{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
		},
		NodePeers: map[int][]int{
			1: {
				0,
			},
		},
		NodeReplicators: map[int][]int{
			0: {
				2,
			},
		},
		SeedDocuments: map[int]string{
			0: `{
				"Name": "John",
				"Age": 21
			}`,
		},
		Creates: map[int]map[int]string{
			0: {
				1: `{
					"Name": "Shahzad",
					"Age": 3000
				}`,
			},
		},
		Results: map[int]map[int]map[string]any{
			0: {
				0: {
					"Age": uint64(21),
				},
				1: {
					"Age": uint64(3000),
				},
			},
			1: {
				0: {
					"Age": uint64(21),
				},
			},
			2: {
				0: {
					"Age": uint64(21),
				},
				1: {
					"Age": uint64(3000),
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
*/
