// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package tests

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	toml "github.com/pelletier/go-toml"
	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/sourcehub/sdk"
	"github.com/stretchr/testify/require"

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/internal/db"
	"github.com/sourcenetwork/defradb/keyring"
	"github.com/sourcenetwork/defradb/node"
	"github.com/sourcenetwork/defradb/tests/clients/cli"
	"github.com/sourcenetwork/defradb/tests/clients/http"
)

type ACPType string

const (
	acpTypeEnvName = "DEFRA_ACP_TYPE"
)

const (
	SourceHubACPType ACPType = "source-hub"
	LocalACPType     ACPType = "local"
)

const (
	// authTokenExpiration is the expiration time for auth tokens.
	authTokenExpiration = time.Minute * 1
)

var (
	acpType ACPType
)

func init() {
	acpType = ACPType(os.Getenv(acpTypeEnvName))
	if acpType == "" {
		acpType = LocalACPType
	}
}

// AddPolicy will attempt to add the given policy using DefraDB's ACP system.
type AddPolicy struct {
	// NodeID may hold the ID (index) of the node we want to add policy to.
	//
	// If a value is not provided the policy will be added in all nodes.
	NodeID immutable.Option[int]

	// The raw policy string.
	Policy string

	// The policy creator identity, i.e. actor creating the policy.
	Identity immutable.Option[int]

	// The expected policyID generated based on the Policy loaded in to the ACP system.
	ExpectedPolicyID string

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

// addPolicyACP will attempt to add the given policy using DefraDB's ACP system.
func addPolicyACP(
	s *state,
	action AddPolicy,
) {
	// If we expect an error, then ExpectedPolicyID should be empty.
	if action.ExpectedError != "" && action.ExpectedPolicyID != "" {
		require.Fail(s.t, "Expected error should not have an expected policyID with it.", s.testCase.Description)
	}

	for i, node := range getNodes(action.NodeID, s.nodes) {
		identity := getIdentity(s, i, action.Identity)
		ctx := db.SetContextIdentity(s.ctx, identity)
		policyResult, err := node.AddPolicy(ctx, action.Policy)

		if err == nil {
			require.Equal(s.t, action.ExpectedError, "")
			require.Equal(s.t, action.ExpectedPolicyID, policyResult.PolicyID)
		}

		expectedErrorRaised := AssertError(s.t, s.testCase.Description, err, action.ExpectedError)
		assertExpectedErrorRaised(s.t, s.testCase.Description, action.ExpectedError, expectedErrorRaised)
	}
}

func setupSourceHub(s *state) ([]node.ACPOpt, error) {
	var isACPTest bool
	for _, a := range s.testCase.Actions {
		if _, ok := a.(AddPolicy); ok {
			isACPTest = true
			break
		}
	}
	if !isACPTest {
		// Spinning up SourceHub instances is a bit slow, so we should be quite aggressive in trimming down the
		// runtime of the test suite when SourceHub ACP is selected.
		s.t.Skipf("test has no ACP elements when testing with SourceHub ACP")
	}

	const moniker string = "foo"
	const chainID string = "sourcehub-test"
	const validatorName string = "test-validator"
	const keyringBackend string = "test"
	directory := s.t.TempDir()

	kr, err := keyring.OpenFileKeyring(
		directory,
		keyring.PromptFunc(func(s string) ([]byte, error) {
			return []byte("secret"), nil
		}),
	)
	if err != nil {
		return nil, err
	}

	// Generate the keys using the index as the seed so that multiple
	// runs yield the same private key.  This is important for stuff like
	// the change detector.
	source := rand.NewSource(0)
	r := rand.New(source)

	acpKey, err := secp256k1.GeneratePrivateKeyFromRand(r)
	require.NoError(s.t, err)
	acpKeyHex := hex.EncodeToString(acpKey.Serialize())

	err = kr.Set(validatorName, acpKey.Serialize())
	if err != nil {
		return nil, err
	}

	out, err := exec.Command("sourcehubd", "init", moniker, "--chain-id", chainID, "--home", directory).CombinedOutput()
	s.t.Log(string(out))
	if err != nil {
		return nil, err
	}

	// Annoyingly, the CLI does not support changing the comet config params that we need,
	// so we have to manually rewrite the config file.
	cfg, err := toml.LoadFile(filepath.Join(directory, "config", "config.toml"))
	if err != nil {
		return nil, err
	}

	fo, err := os.Create(filepath.Join(directory, "config", "config.toml"))
	if err != nil {
		return nil, err
	}

	// Speed up the rate at which the blocks are created, this is particularly important for getting
	// the first block created on the `sourcehubd start` call at the end of this function as
	// we cannot use the node until the first block has been created.
	cfg.Set("consensus.timeout_propose", "0.5s")
	cfg.Set("consensus.timeout_commit", "1s")

	_, err = cfg.WriteTo(fo)
	if err != nil {
		return nil, err
	}
	err = fo.Close()
	if err != nil {
		return nil, err
	}

	out, err = exec.Command(
		"sourcehubd", "keys", "import-hex", validatorName, acpKeyHex,
		"--keyring-backend", keyringBackend,
		"--home", directory,
	).CombinedOutput()
	s.t.Log(string(out))
	if err != nil {
		return nil, err
	}

	out, err = exec.Command(
		"sourcehubd", "keys", "show", validatorName,
		"--address",
		"--keyring-backend", keyringBackend,
		"--home", directory,
	).CombinedOutput()
	s.t.Log(string(out))
	if err != nil {
		return nil, err
	}

	// The result is suffexed with a newline char so we must trim the whitespace
	validatorAddress := strings.TrimSpace(string(out))
	s.sourcehubAddress = validatorAddress

	out, err = exec.Command(
		"sourcehubd", "genesis", "add-genesis-account", validatorAddress, "900000000stake",
		"--keyring-backend", keyringBackend,
		"--home", directory,
	).CombinedOutput()
	s.t.Log(string(out))
	if err != nil {
		return nil, err
	}

	out, err = exec.Command(
		"sourcehubd", "genesis", "gentx", validatorName, "10000000stake",
		"--chain-id", chainID,
		"--keyring-backend", keyringBackend,
		"--home", directory,
	).CombinedOutput()
	s.t.Log(string(out))
	if err != nil {
		return nil, err
	}

	out, err = exec.Command("sourcehubd", "genesis", "collect-gentxs", "--home", directory).CombinedOutput()
	s.t.Log(string(out))
	if err != nil {
		return nil, err
	}

	// We need to lock across all the test processes as we assign ports to the source hub instance as this
	// process involves finding free ports, dropping them, and then assigning them to the source hub node.
	//
	// We have to do this because source hub (cosmos) annoyingly does not support automatic port assignment
	// (appart from the p2p port which we just managage here for consistency).
	//
	// We need to lock before getting the ports, otherwise they may try and use the port we use for locking.
	// We can only unlock after the source hub node has started and begun listening on the assigned ports.
	unlock, err := crossLock(55555)
	if err != nil {
		return nil, err
	}
	defer unlock()

	gRpcPort, releaseGrpcPort, err := getFreePort()
	if err != nil {
		return nil, err
	}

	rpcPort, releaseRpcPort, err := getFreePort()
	if err != nil {
		return nil, err
	}

	p2pPort, releaseP2pPort, err := getFreePort()
	if err != nil {
		return nil, err
	}

	pprofPort, releasePprofPort, err := getFreePort()
	if err != nil {
		return nil, err
	}

	gRpcAddress := fmt.Sprintf("127.0.0.1:%v", gRpcPort)
	rpcAddress := fmt.Sprintf("tcp://127.0.0.1:%v", rpcPort)
	p2pAddress := fmt.Sprintf("tcp://127.0.0.1:%v", p2pPort)
	pprofAddress := fmt.Sprintf("127.0.0.1:%v", pprofPort)

	releaseGrpcPort()
	releaseRpcPort()
	releaseP2pPort()
	releasePprofPort()

	sourceHubCmd := exec.Command(
		"sourcehubd",
		"start",
		"--minimum-gas-prices", "0stake",
		"--home", directory,
		"--grpc.address", gRpcAddress,
		"--rpc.laddr", rpcAddress,
		"--p2p.laddr", p2pAddress,
		"--rpc.pprof_laddr", pprofAddress,
	)
	var bf testBuffer
	bf.Lines = make(chan string, 100)
	sourceHubCmd.Stdout = &bf
	sourceHubCmd.Stderr = &bf

	err = sourceHubCmd.Start()
	if err != nil {
		return nil, err
	}

	timeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)

cmdReaderLoop:
	for {
		select {
		case <-timeout.Done():
			break cmdReaderLoop
		case line := <-bf.Lines:
			if strings.Contains(line, "starting gRPC server...") {
				// The Comet RPC server is spun up before the gRPC one, so we
				// can safely unlock here.
				unlock()
			}
			// This is guarenteed to be logged after the gRPC server has been spun up
			// so we can be sure that the lock has been unlocked.
			if strings.Contains(line, "committed state") {
				break cmdReaderLoop
			}
		}
	}

	cancel()
	// Void the buffer so that it doesn't fill up and block the process under test
	bf.Void()

	s.t.Cleanup(
		func() {
			err := sourceHubCmd.Process.Kill()
			require.NoError(s.t, err)
		},
	)

	signer, err := keyring.NewTxSignerFromKeyringKey(kr, validatorName)
	if err != nil {
		return nil, err
	}

	return []node.ACPOpt{
		node.WithTxnSigner(immutable.Some[sdk.TxSigner](signer)),
		node.WithSourceHubChainID(chainID),
		node.WithSourceHubGRPCAddress(gRpcAddress),
		node.WithSourceHubCometRPCAddress(rpcAddress),
	}, nil
}

func getFreePort() (int, func(), error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, nil, err
	}

	return l.Addr().(*net.TCPAddr).Port,
		func() {
			// there are no errors that this returns that we actually care about
			_ = l.Close()
		},
		nil
}

// crossLock forms a cross process lock by attempting to listen to the given port.
//
// This function will only return once the port is free. A function to unbind from the
// port is returned - this unlock function may be called multiple times without issue.
func crossLock(port uint16) (func(), error) {
	l, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%v", port))
	if err != nil {
		if strings.Contains(err.Error(), "address already in use") {
			time.Sleep(5 * time.Millisecond)
			return crossLock(port)
		}
		return nil, err
	}

	return func() {
			// there are no errors that this returns that we actually care about
			_ = l.Close()
		},
		nil
}

func getIdentity(s *state, nodeIndex int, index immutable.Option[int]) immutable.Option[acpIdentity.Identity] {
	if !index.HasValue() {
		return immutable.None[acpIdentity.Identity]()
	}

	if len(s.identities) <= nodeIndex {
		identities := make([][]acpIdentity.Identity, nodeIndex+1)
		copy(identities, s.identities)
		s.identities = identities
	}
	nodeIdentities := s.identities[nodeIndex]

	if len(nodeIdentities) <= index.Value() {
		var audience immutable.Option[string]
		switch client := s.nodes[nodeIndex].(type) {
		case *http.Wrapper:
			audience = immutable.Some(strings.TrimPrefix(client.Host(), "http://"))
		case *cli.Wrapper:
			audience = immutable.Some(strings.TrimPrefix(client.Host(), "http://"))
		}

		// Generate the keys using the index as the seed so that multiple
		// runs yield the same private key.  This is important for stuff like
		// the change detector.
		source := rand.NewSource(int64(index.Value()))
		r := rand.New(source)

		privateKey, err := secp256k1.GeneratePrivateKeyFromRand(r)
		require.NoError(s.t, err)

		identity, err := acpIdentity.FromPrivateKey(
			privateKey,
			authTokenExpiration,
			audience,
			immutable.Some(s.sourcehubAddress),
		)
		require.NoError(s.t, err)

		nodeIdentities = append(nodeIdentities, identity)
		s.identities[nodeIndex] = nodeIdentities

		return immutable.Some(identity)
	} else {
		return immutable.Some(nodeIdentities[index.Value()])
	}
}

// testBuffer is a very simple, thread-safe (--race flag friendly), io.Writer
// implementation that allows us to easily access the out/err outputs of CLI commands.
//
// Calling void will result in all writes being discarded.
type testBuffer struct {
	Lines chan string
	void  atomic.Bool
}

var _ io.Writer = (*testBuffer)(nil)

func (b *testBuffer) Write(p []byte) (n int, err error) {
	if !b.void.Load() {
		b.Lines <- string(p)
	}
	return len(p), nil
}

func (b *testBuffer) Void() {
	b.void.Swap(true)
}
