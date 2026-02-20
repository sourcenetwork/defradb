// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

//go:build !js

package tests

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	cdc "github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocdc "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	cosmoskeyring "github.com/cosmos/cosmos-sdk/crypto/keyring"
	cosmossecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cosmostypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tclog "github.com/testcontainers/testcontainers-go/log"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/keyring"
	"github.com/sourcenetwork/defradb/tests/state"
	"github.com/sourcenetwork/sourcehub/sdk"
)

const (
	// faucetMnemonic is the mnemonic for a static faucet account present in the
	// STANDALONE version of the SourceHub image
	faucetMnemonic = "comic very pond victory suit tube ginger antique life then core warm loyal deliver iron fashion erupt husband weekend monster sunny artist empty uphold" //nolint:lll

	// faucetAddr is the account address matching the faucetMnemonic
	faucetAddr = "source12d9hjf0639k995venpv675sju9ltsvf8u5c9jt"

	sourcehubTestChainID string = "sourcehub-dev"
)

func setupSourceHub(s *state.State, testCase TestCase) (*options.NodeDocumentACPOptions, error) {
	var isDocumentACPTest bool
	for _, a := range testCase.Actions {
		switch a.(type) {
		case
			AddDACPolicy,
			AddDACActorRelationship,
			DeleteDACActorRelationship:
			isDocumentACPTest = true
		}
	}

	if !isDocumentACPTest {
		// Spinning up SourceHub instances is a bit slow, so we should be quite aggressive in trimming down the
		// runtime of the test suite when SourceHub ACP is selected.
		s.T.Skipf("test has no document ACP elements when testing with SourceHub ACP")
	}

	ctx := context.Background()
	testLogger := tclog.TestLogger(s.T)

	name := uuid.New()
	container, err := testcontainers.Run(ctx,
		sourcehubImage,
		testcontainers.WithName(name.String()),
		testcontainers.WithExposedPorts("26657/tcp"),
		testcontainers.WithExposedPorts("9090/tcp"),
		testcontainers.WithLogger(testLogger),
		testcontainers.WithEnv(map[string]string{
			// STANDALONE configures the SH container to create an isolated chain,
			// instead of connecting to an existing one.
			"STANDALONE": "1",
		}),
	)
	if err != nil {
		return nil, err
	}

	// read container logs before terminating it
	s.T.Cleanup(func() {
		s.T.Helper()
		logs, err := container.Logs(ctx)
		if err != nil {
			s.T.Logf("could not read container logs")
		} else {
			buf := bytes.Buffer{}
			// errors during cleanup don't affect anything
			buf.ReadFrom(logs) //nolint:errcheck
			s.T.Logf("container logs: %v", buf.String())
			logs.Close() //nolint:errcheck
		}
		testcontainers.TerminateContainer(container) //nolint:errcheck
	})

	grpcEndpoint, err := container.PortEndpoint(ctx, "9090", "")
	if err != nil {
		return nil, err
	}
	rpcEndpoint, err := container.PortEndpoint(ctx, "26657", "tcp")
	if err != nil {
		return nil, err
	}
	s.T.Logf("sourcehub endpoints: grpc=%v, rpc=%v", grpcEndpoint, rpcEndpoint)

	s.SourcehubAddress = faucetAddr

	err = waitForSourceHub(s.T, grpcEndpoint, rpcEndpoint, faucetAddr)
	if err != nil {
		return nil, err
	}

	privKeyBytes := getAccountDataFromMnemonic(s.T, faucetMnemonic)
	kr, err := keyring.OpenFileKeyring(
		s.T.TempDir(),
		[]byte("secret"),
	)
	if err != nil {
		return nil, err
	}

	err = kr.Set("validator", privKeyBytes)
	if err != nil {
		return nil, err
	}

	signer, err := keyring.NewTxSignerFromKeyringKey(kr, "validator")
	if err != nil {
		return nil, err
	}

	return &options.NodeDocumentACPOptions{
		DocumentACPType:          options.NodeSourceHubDocumentACPType,
		Signer:                   immutable.Some[options.NodeTxSigner](signer),
		SourceHubChainID:         sourcehubTestChainID,
		SourceHubGRPCAddress:     grpcEndpoint,
		SourceHubCometRPCAddress: rpcEndpoint,
	}, nil
}

func waitForSourceHub(t testing.TB, grpcEndpoint, cometRpcEndpoint string, accAddr string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	i := 1
	startTs := time.Now()
	for {
		// use an exponential backoff timer to adjust polling
		timer := time.After(time.Duration(i) * (10 * time.Millisecond))
		i++
		select {
		case <-ctx.Done():
			t.Logf("time out waiting for sourcehub to start")
			return fmt.Errorf("error setting up SourceHub: connection not ready after deadline")
		case <-timer:
			ok := probeSourceHub(ctx, grpcEndpoint, cometRpcEndpoint, accAddr)
			if ok {
				elapsed := time.Since(startTs)
				t.Logf("sourcehub ready to receive connections: after %v", elapsed)
				return nil
			}
		}
	}
}

// probeSourceHub is a readiness probe which tries to connect to SourceHub's
// RPC endpoint to determine if it is ready to receive connections.
// Returns true if the probe succeeded.
func probeSourceHub(ctx context.Context, grpcAddr, cometRpcAddr, knownAddr string) bool {
	client, err := sdk.NewClient(
		sdk.WithGRPCAddr(grpcAddr),
		sdk.WithCometRPCAddr(cometRpcAddr),
	)
	if err != nil {
		return false
	}
	defer client.Close()

	// probe rpc service
	height := int64(1)
	_, err = client.CometBFTRPCClient().Block(ctx, &height)
	if err != nil {
		return false
	}

	// probe grpc service
	_, err = client.AuthQueryClient().Account(ctx, &types.QueryAccountRequest{
		Address: knownAddr,
	})
	return err == nil
}

// getAccountDataFromMnemonic returns the private key bytes
// from a given sourcehub mnemonic.
// assumes the mnemonic is for a secp256k1 key
func getAccountDataFromMnemonic(t testing.TB, mnemonic string) []byte {
	registry := cdctypes.NewInterfaceRegistry()
	cryptocdc.RegisterInterfaces(registry)
	codec := cdc.NewProtoCodec(registry)

	kb := cosmoskeyring.NewInMemory(codec)
	rec, err := kb.NewAccount("key", faucetMnemonic, "", cosmostypes.GetConfig().GetFullBIP44Path(), hd.Secp256k1)
	require.NoError(t, err)
	item, ok := rec.Item.(*cosmoskeyring.Record_Local_)
	require.True(t, ok)
	privKeyRecByes := item.Local.PrivKey.Value
	privKey := cosmossecp256k1.PrivKey{}
	err = privKey.Unmarshal(privKeyRecByes)
	require.NoError(t, err)
	return privKey.Bytes()
}
