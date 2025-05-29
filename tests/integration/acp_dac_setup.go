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

	"github.com/sourcenetwork/defradb/keyring"
	"github.com/sourcenetwork/defradb/node"
	"github.com/sourcenetwork/defradb/tests/clients/cli"
	"github.com/sourcenetwork/defradb/tests/clients/http"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	toml "github.com/pelletier/go-toml"
	"github.com/sourcenetwork/immutable"
	"github.com/stretchr/testify/require"
)

func getNodeAudience(s *state, nodeIndex int) immutable.Option[string] {
	if nodeIndex >= len(s.nodes) {
		return immutable.None[string]()
	}
	switch client := s.nodes[nodeIndex].Client.(type) {
	case *http.Wrapper:
		return immutable.Some(strings.TrimPrefix(client.Host(), "http://"))
	case *cli.Wrapper:
		return immutable.Some(strings.TrimPrefix(client.Host(), "http://"))
	}

	return immutable.None[string]()
}

func setupSourceHub(s *state) ([]node.DocumentACPOpt, error) {
	var isDocumentACPTest bool
	for _, a := range s.testCase.Actions {
		switch a.(type) {
		case
			AddPolicyWithDAC,
			AddActorRelationshipWithDAC,
			DeleteActorRelationshipWithDAC:
			isDocumentACPTest = true
		}
	}

	if !isDocumentACPTest {
		// Spinning up SourceHub instances is a bit slow, so we should be quite aggressive in trimming down the
		// runtime of the test suite when SourceHub ACP is selected.
		s.t.Skipf("test has no document ACP elements when testing with SourceHub ACP")
	}

	const moniker string = "foo"
	const chainID string = "sourcehub-test"
	const validatorName string = "test-validator"
	const keyringBackend string = "test"
	directory := s.t.TempDir()

	kr, err := keyring.OpenFileKeyring(
		directory,
		[]byte("secret"),
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

	args := []string{"init", moniker, "--chain-id", chainID, "--home", directory}
	s.t.Log("$ sourcehubd " + strings.Join(args, " "))
	out, err := exec.Command("sourcehubd", args...).CombinedOutput()
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

	args = []string{
		"keys", "import-hex", validatorName, acpKeyHex,
		"--keyring-backend", keyringBackend,
		"--home", directory,
	}

	s.t.Log("$ sourcehubd " + strings.Join(args, " "))
	out, err = exec.Command("sourcehubd", args...).CombinedOutput()
	s.t.Log(string(out))
	if err != nil {
		return nil, err
	}

	args = []string{
		"keys", "show", validatorName,
		"--address",
		"--keyring-backend", keyringBackend,
		"--home", directory,
	}
	s.t.Log("$ sourcehubd " + strings.Join(args, " "))
	out, err = exec.Command("sourcehubd", args...).CombinedOutput()
	s.t.Log(string(out))
	if err != nil {
		return nil, err
	}

	// The result is suffixed with a newline char so we must trim the whitespace
	validatorAddress := strings.TrimSpace(string(out))
	s.sourcehubAddress = validatorAddress

	args = []string{"genesis", "add-genesis-account", validatorAddress, "1000000000uopen",
		"--keyring-backend", keyringBackend,
		"--home", directory,
	}
	s.t.Log("$ sourcehubd " + strings.Join(args, " "))
	out, err = exec.Command("sourcehubd", args...).CombinedOutput()
	s.t.Log(string(out))
	if err != nil {
		return nil, err
	}

	args = []string{"genesis", "gentx", validatorName, "100000000uopen",
		"--chain-id", chainID,
		"--keyring-backend", keyringBackend,
		"--home", directory}
	s.t.Log("$ sourcehubd " + strings.Join(args, " "))
	out, err = exec.Command("sourcehubd", args...).CombinedOutput()
	s.t.Log(string(out))
	if err != nil {
		return nil, err
	}

	args = []string{"genesis", "collect-gentxs", "--home", directory}
	s.t.Log("$ sourcehubd " + strings.Join(args, " "))
	out, err = exec.Command("sourcehubd", args...).CombinedOutput()
	s.t.Log(string(out))
	if err != nil {
		return nil, err
	}

	// We need to lock across all the test processes as we assign ports to the source hub instance as this
	// process involves finding free ports, dropping them, and then assigning them to the source hub node.
	//
	// We have to do this because source hub (cosmos) annoyingly does not support automatic port assignment
	// (apart from the p2p port which we just manage here for consistency).
	//
	// We need to lock before getting the ports, otherwise they may try and use the port we use for locking.
	// We can only unlock after the source hub node has started and begun listening on the assigned ports.
	unlock, err := crossLock(44444)
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

	args = []string{
		"start",
		"--minimum-gas-prices", "0uopen",
		"--home", directory,
		"--grpc.address", gRpcAddress,
		"--rpc.laddr", rpcAddress,
		"--p2p.laddr", p2pAddress,
		"--rpc.pprof_laddr", pprofAddress,
	}
	s.t.Log("$ sourcehubd " + strings.Join(args, " "))
	sourceHubCmd := exec.Command("sourcehubd", args...)
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
			// This is guaranteed to be logged after the gRPC server has been spun up
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

	return []node.DocumentACPOpt{
		node.WithTxnSigner(immutable.Some[node.TxSigner](signer)),
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

	return l.Addr().(*net.TCPAddr).Port, //nolint:forcetypeassert
		func() {
			// there are no errors that this returns that we actually care about
			_ = l.Close()
		},
		nil
}

// crossLock forms a cross process lock by attempting to listen to the given port.
//
// This function will only return once the port is free or the timeout is reached.
// A function to unbind from the port is returned - this unlock function may be called
// multiple times without issue.
func crossLock(port uint16) (func(), error) {
	timeout := time.After(20 * time.Second)
	for {
		select {
		case <-timeout:
			return nil, fmt.Errorf("timeout reached while trying to acquire cross process lock on port %v", port)
		default:
			l, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%v", port))
			if err != nil {
				if strings.Contains(err.Error(), "address already in use") {
					time.Sleep(5 * time.Millisecond)
					continue
				}
				return nil, err
			}

			return func() {
					// there are no errors that this returns that we actually care about
					_ = l.Close()
				},
				nil
		}
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
