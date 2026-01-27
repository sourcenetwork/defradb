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

	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/google/uuid"
	"github.com/sourcenetwork/defradb/keyring"
	"github.com/sourcenetwork/defradb/node"
	"github.com/sourcenetwork/defradb/tests/state"
	"github.com/sourcenetwork/sourcehub/sdk"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"

	"github.com/sourcenetwork/immutable"

	cdc "github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocdc "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	cosmoskeyring "github.com/cosmos/cosmos-sdk/crypto/keyring"
	cosmossecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cosmostypes "github.com/cosmos/cosmos-sdk/types"
	tclog "github.com/testcontainers/testcontainers-go/log"
)

func setupSourceHub(s *state.State, testCase TestCase) ([]node.DocumentACPOpt, error) {
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
	const chainID string = "sourcehub-dev"
	name := uuid.New()

	testLogger := tclog.TestLogger(s.T)
	ctx := context.Background()
	img := fmt.Sprintf("ghcr.io/sourcenetwork/sourcehub:refactor-df-defra") // TODO
	container, err := testcontainers.Run(ctx,
		img,
		testcontainers.WithName(name.String()),
		testcontainers.WithExposedPorts("26657/tcp"),
		testcontainers.WithExposedPorts("9090/tcp"),
		testcontainers.WithLogger(testLogger),
		testcontainers.WithEnv(map[string]string{
			"STANDALONE": "1",
		}),
	)
	if err != nil {
		return nil, err
	}

	s.T.Cleanup(func() {
		logs, err := container.Logs(context.Background())
		require.NoError(s.T, err)
		buf := bytes.Buffer{}
		buf.ReadFrom(logs)
		s.T.Logf("container logs: %v", buf.String())
		testcontainers.TerminateContainer(container)
	})
	grpcEndpoint, err := container.PortEndpoint(ctx, "9090", "")
	if err != nil {
		return nil, err
	}
	rpcEndpoint, err := container.PortEndpoint(ctx, "26657", "tcp")
	if err != nil {
		return nil, err
	}

	s.T.Logf(
		"sourcehub endpoints: grpc=%v, rpc=%v", grpcEndpoint, rpcEndpoint,
	)

	faucetMnemonic := "comic very pond victory suit tube ginger antique life then core warm loyal deliver iron fashion erupt husband weekend monster sunny artist empty uphold"

	faucetAddr := "source12d9hjf0639k995venpv675sju9ltsvf8u5c9jt"
	s.SourcehubAddress = faucetAddr
	err = waitForSourceHub(s.T, container, grpcEndpoint, rpcEndpoint, faucetAddr)
	if err != nil {
		return nil, err
	}

	registry := cdctypes.NewInterfaceRegistry()
	cryptocdc.RegisterInterfaces(registry)
	codec := cdc.NewProtoCodec(registry)
	kb := cosmoskeyring.NewInMemory(codec)
	rec, err := kb.NewAccount("key", faucetMnemonic, "", cosmostypes.GetConfig().GetFullBIP44Path(), hd.Secp256k1)
	require.NoError(s.T, err)
	keyRecBz := rec.Item.(*cosmoskeyring.Record_Local_).Local.PrivKey.Value
	privKey := cosmossecp256k1.PrivKey{}
	err = privKey.Unmarshal(keyRecBz)
	require.NoError(s.T, err)

	kr, err := keyring.OpenFileKeyring(
		s.T.TempDir(),
		[]byte("secret"),
	)
	if err != nil {
		return nil, err
	}

	// Generate the keys using the index as the seed so that multiple
	// runs yield the same private key.  This is important for stuff like
	// the change detector.
	//source := rand.NewSource(0)
	//r := rand.New(source)

	//acpKey, err := secp256k1.GeneratePrivateKeyFromRand(r)
	//require.NoError(s.T, err)
	//acpKeyHex := hex.EncodeToString(acpKey.Serialize())

	err = kr.Set("validator", privKey.Bytes())
	if err != nil {
		return nil, err
	}

	signer, err := keyring.NewTxSignerFromKeyringKey(kr, "validator")
	if err != nil {
		return nil, err
	}

	return []node.DocumentACPOpt{
		node.WithTxnSigner(immutable.Some[node.TxSigner](signer)),
		node.WithSourceHubChainID(chainID),
		node.WithSourceHubGRPCAddress(grpcEndpoint),
		node.WithSourceHubCometRPCAddress(rpcEndpoint),
	}, nil
}

func waitForSourceHub(t testing.TB, container testcontainers.Container, grpcEndpoint, cometRpcEndpoint string, valAddr string) error {
	timeout := time.After(5 * time.Second)
	i := 1
	startTs := time.Now()
	for {
		// use an exponential backoff timer to adjust polling
		timer := time.After(time.Duration(i) * (10 * time.Millisecond))
		i++
		select {
		case <-timeout:
			t.Logf("time out waiting for sourcehub to start")
			return fmt.Errorf("error setting up SourceHub: connection not ready after deadline")
		case <-timer:
			ok := probeSourceHub(grpcEndpoint, cometRpcEndpoint, valAddr)
			if ok {
				elapsed := time.Since(startTs)
				t.Logf("sourcehub ready to receive connections: after %v", elapsed)
				return nil
			}
		}
	}
}

// probeSourceHub is a rediness probe which tries to connect to SourceHub's
// RPC endpoint to determine if it is ready to receive connections
// Returns true if the probe succeeded
func probeSourceHub(grpcAddr, cometRpcAddr, valAddr string) bool {
	client, err := sdk.NewClient(
		sdk.WithGRPCAddr(grpcAddr),
		sdk.WithCometRPCAddr(cometRpcAddr),
	)
	defer func() {
		recover()
	}()
	defer client.Close()

	if err != nil {
		return false
	}

	// probe rpc service
	height := int64(1)
	_, err = client.CometBFTRPCClient().Block(context.Background(), &height)
	if err != nil {
		return false
	}

	// probe grpc service
	_, err = client.AuthQueryClient().Account(context.Background(), &types.QueryAccountRequest{
		Address: valAddr,
	})
	return err == nil
}
