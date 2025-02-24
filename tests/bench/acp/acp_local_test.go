// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package bench_acp

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/acp"
	"github.com/sourcenetwork/defradb/acp/identity"
)

var identity1 = identity.Identity{
	DID: "did:key:z7r8os2G88XXBNBTLj3kFR5rzUJ4VAesbX7PgsA68ak9B5RYcXF5EZEmjRzzinZndPSSwujXb4XKHG6vmKEFG6ZfsfcQn",
}

//var identity2 = identity.Identity{
//	DID: "did:key:z7r8ooUiNXK8TT8Xjg1EWStR2ZdfxbzVfvGWbA2FjmzcnmDxz71QkP1Er8PP3zyLZpBLVgaXbZPGJPS4ppXJDPRcqrx4F",
//}
//
//var invalidIdentity = identity.Identity{
//	DID: "did:something",
//}

var validPolicyID string = "d59f91ba65fe142d35fc7df34482eafc7e99fed7c144961ba32c4664634e61b7"
var validPolicy string = `
name: test
description: a policy

actor:
  name: actor

resources:
  users:
    permissions:
      write:
        expr: owner
      read:
        expr: owner + reader

    relations:
      owner:
        types:
          - actor
      reader:
        types:
          - actor
 `

// newLocalACPSetup will setup localACP instance in memory if inMem is true and a persistent store otherwise.
// Additionally it will also start the acp instance.
// The caller is responsible to call `Close()` on the returned [acp.ACP] instance.
func newLocalACPSetup(b *testing.B, inMem bool) acp.ACP {
	ctx := context.Background()
	localACP := acp.NewLocalACP()

	if inMem {
		localACP.Init(ctx, "")
	} else {
		acpPath := b.TempDir()
		localACP.Init(ctx, acpPath)
	}

	err := localACP.Start(ctx)
	require.Nil(b, err)

	return localACP
}

// resetLocalACPKeepPolicy resets the local acp instance then adds our desired policy back.
func resetLocalACPKeepPolicy(b *testing.B, ctx context.Context, localACP acp.ACP) {
	resetErr := localACP.ResetState(ctx)
	require.Nil(b, resetErr)

	policyID, errAddPolicy := localACP.AddPolicy(
		ctx,
		identity1,
		validPolicy,
	)
	require.Nil(b, errAddPolicy)
	require.Equal(
		b,
		validPolicyID,
		policyID,
	)
}

func registerXDocObjects(b *testing.B, ctx context.Context, count int, localACP acp.ACP) {
	for index := 0; index < count; index++ {
		err := localACP.RegisterDocObject(
			ctx,
			identity1,
			validPolicyID,
			"users",
			strconv.Itoa(index),
		)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkACPRegister(b *testing.B) {
	for _, inMemoryOrPersistent := range []bool{true, false} {
		for _, scaleBy := range []int{256, 512, 1024, 2048, 4096, 8192} {
			b.Run(
				fmt.Sprintf("scale=%d,inMem=%t", scaleBy, inMemoryOrPersistent),
				func(b *testing.B) {
					localACP := newLocalACPSetup(b, inMemoryOrPersistent)
					defer localACP.Close()

					b.ResetTimer()
					for bNIndex := 0; bNIndex < b.N; bNIndex++ {
						// Since we need to re-initialize for every run use stop-start.
						b.StopTimer()
						ctx := context.Background()
						resetLocalACPKeepPolicy(b, ctx, localACP)

						b.StartTimer()
						registerXDocObjects(b, ctx, scaleBy, localACP)
					}
				},
			)
		}
	}
}

func BenchmarkACPIsDocRegistered(b *testing.B) {
	for _, inMemoryOrPersistent := range []bool{true, false} {
		for _, scaleBy := range []int{256, 512, 1024, 2048, 4096, 8192} {
			b.Run(
				fmt.Sprintf("scale=%d,inMem=%t", scaleBy, inMemoryOrPersistent),
				func(b *testing.B) {
					localACP := newLocalACPSetup(b, inMemoryOrPersistent)
					defer localACP.Close()

					b.ResetTimer()
					for bNIndex := 0; bNIndex < b.N; bNIndex++ {
						// Since we need to re-initialize for every run use stop-start.
						b.StopTimer()
						ctx := context.Background()
						resetLocalACPKeepPolicy(b, ctx, localACP)
						registerXDocObjects(b, ctx, scaleBy, localACP)

						b.StartTimer()
						_, err := localACP.IsDocRegistered(ctx, validPolicyID, "users", "1")
						if err != nil {
							b.Fatal(err)
						}
					}
				},
			)
		}
	}
}

func BenchmarkACPCheckDocAccess(b *testing.B) {
	for _, inMemoryOrPersistent := range []bool{true, false} {
		for _, scaleBy := range []int{256, 512, 1024, 2048, 4096, 8192} {
			b.Run(
				fmt.Sprintf("scale=%d,inMem=%t", scaleBy, inMemoryOrPersistent),
				func(b *testing.B) {
					localACP := newLocalACPSetup(b, inMemoryOrPersistent)
					defer localACP.Close()

					b.ResetTimer()
					for bNIndex := 0; bNIndex < b.N; bNIndex++ {
						// Since we need to re-initialize for every run use stop-start.
						b.StopTimer()
						ctx := context.Background()
						resetLocalACPKeepPolicy(b, ctx, localACP)
						registerXDocObjects(b, ctx, scaleBy, localACP)

						b.StartTimer()
						_, err := localACP.CheckDocAccess(
							ctx,
							acp.ReadPermission,
							identity1.DID,
							validPolicyID,
							"users",
							"1",
						)
						if err != nil {
							b.Fatal(err)
						}
					}
				},
			)
		}
	}
}
