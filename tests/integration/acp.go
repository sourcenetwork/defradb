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
	"slices"
	"strings"
	"sync/atomic"
	"time"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	toml "github.com/pelletier/go-toml"
	"github.com/sourcenetwork/immutable"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/keyring"
	"github.com/sourcenetwork/defradb/node"
	"github.com/sourcenetwork/defradb/tests/clients/cli"
	"github.com/sourcenetwork/defradb/tests/clients/http"
)

type DocumentACPType string

const (
	documentACPTypeEnvName = "DEFRA_DOCUMENT_ACP_TYPE"
)

const (
	SourceHubDocumentACPType DocumentACPType = "source-hub"
	LocalDocumentACPType     DocumentACPType = "local"
)

const (
	// authTokenExpiration is the expiration time for auth tokens.
	authTokenExpiration = time.Minute * 1
)

var (
	documentACPType DocumentACPType
)

// KMSType is the type of KMS to use.
type KMSType string

const (
	// NoneKMSType is the none KMS type. It is used to indicate that no KMS should be used.
	NoneKMSType KMSType = "none"
	// PubSubKMSType is the PubSub KMS type.
	PubSubKMSType KMSType = "pubsub"
)

func getKMSTypes() []KMSType {
	return []KMSType{PubSubKMSType}
}

func init() {
	documentACPType = DocumentACPType(os.Getenv(documentACPTypeEnvName))
	if documentACPType == "" {
		documentACPType = LocalDocumentACPType
	}
}

// AddDocPolicy will attempt to add the given policy using DefraDB's Document ACP system.
type AddDocPolicy struct {
	// NodeID may hold the ID (index) of the node we want to add policy to.
	//
	// If a value is not provided the policy will be added in all nodes, unless testing with
	// sourcehub ACP, in which case the policy will only be defined once.
	NodeID immutable.Option[int]

	// The raw policy string.
	Policy string

	// The policy creator identity, i.e. actor creating the policy.
	Identity immutable.Option[Identity]

	// The expected policyID generated based on the Policy loaded in to the ACP system.
	//
	// This is an optional attribute, for situations where a test might want to assert
	// the exact policyID. When this is not provided the test will just assert that
	// the resulting policyID is not empty.
	ExpectedPolicyID immutable.Option[string]

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

// addPolicyDocumentACP will attempt to add the given policy using DefraDB's Document ACP system.
func addPolicyDocumentACP(
	s *state,
	action AddDocPolicy,
) {
	// If we expect an error, then ExpectedPolicyID should never be provided.
	if action.ExpectedError != "" && action.ExpectedPolicyID.HasValue() {
		require.Fail(s.t, "Expected error should not have an expected policyID with it.", s.testCase.Description)
	}

	nodeIDs, nodes := getNodesWithIDs(action.NodeID, s.nodes)
	maxNodeID := slices.Max(nodeIDs)
	// Expand the policyIDs slice once, so we can minimize how many times we need to expaind it.
	// We use the maximum nodeID provided to make sure policyIDs slice can accomodate upto that nodeID.
	if len(s.policyIDs) <= maxNodeID {
		// Expand the slice if required, so that the policyID can be accessed by node index
		policyIDs := make([][]string, maxNodeID+1)
		copy(policyIDs, s.policyIDs)
		s.policyIDs = policyIDs
	}

	for index, node := range nodes {
		nodeID := nodeIDs[index]
		ctx := getContextWithIdentity(s.ctx, s, action.Identity, nodeID)
		policyResult, err := node.AddPolicy(ctx, action.Policy)

		expectedErrorRaised := AssertError(s.t, s.testCase.Description, err, action.ExpectedError)
		assertExpectedErrorRaised(s.t, s.testCase.Description, action.ExpectedError, expectedErrorRaised)

		if !expectedErrorRaised {
			require.Equal(s.t, action.ExpectedError, "")
			if action.ExpectedPolicyID.HasValue() {
				require.Equal(s.t, action.ExpectedPolicyID.Value(), policyResult.PolicyID)
			} else {
				require.NotEqual(s.t, policyResult.PolicyID, "")
			}

			s.policyIDs[nodeID] = append(s.policyIDs[nodeID], policyResult.PolicyID)
		}

		// The policy should only be added to a SourceHub chain once - there is no need to loop through
		// the nodes.
		if documentACPType == SourceHubDocumentACPType {
			// Note: If we break here the state will only preserve the policyIDs result on the
			// first node if acp type is sourcehub, make sure to replicate the policyIDs state
			// on all the nodes, so we don't have to handle all the edge cases later in actions.
			for otherIndexes := index + 1; otherIndexes < len(nodes); otherIndexes++ {
				s.policyIDs[nodeIDs[otherIndexes]] = s.policyIDs[nodeID]
			}
			break
		}
	}
}

// AddDocActorRelationship will attempt to create a new relationship for a document with an actor.
type AddDocActorRelationship struct {
	// NodeID may hold the ID (index) of the node we want to add doc actor relationship on.
	//
	// If a value is not provided the relationship will be added in all nodes, unless testing with
	// sourcehub ACP, in which case the relationship will only be defined once.
	NodeID immutable.Option[int]

	// The collection in which this document we want to add a relationship for exists.
	//
	// This is a required field. To test the invalid usage of not having this arg, use -1 index.
	CollectionID int

	// The index-identifier of the document within the collection.  This is based on
	// the order in which it was created, not the ordering of the document within the
	// database.
	//
	// This is a required field. To test the invalid usage of not having this arg, use -1 index.
	DocID int

	// The name of the relation to set between document and target actor (should be defined in the policy).
	//
	// This is a required field.
	Relation string

	// The target public identity, i.e. the identity of the actor to tie the document's relation with.
	//
	// This is a required field. To test the invalid usage of not having this arg, use NoIdentity() or leave default.
	TargetIdentity immutable.Option[Identity]

	// The requestor identity, i.e. identity of the actor creating the relationship.
	// Note: This identity must either own or have managing access defined in the policy.
	//
	// This is a required field. To test the invalid usage of not having this arg, use NoIdentity() or leave default.
	RequestorIdentity immutable.Option[Identity]

	// Result returns true if it was a no-op due to existing before, and false if a new relationship was made.
	ExpectedExistence bool

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

func addDocActorRelationshipACP(
	s *state,
	action AddDocActorRelationship,
) {
	var docID string
	actionNodeID := action.NodeID
	nodeIDs, nodes := getNodesWithIDs(action.NodeID, s.nodes)
	for index, node := range nodes {
		nodeID := nodeIDs[index]

		var collectionName string
		collectionName, docID = getCollectionAndDocInfo(s, action.CollectionID, action.DocID, nodeID)

		exists, err := node.AddDocActorRelationship(
			getContextWithIdentity(s.ctx, s, action.RequestorIdentity, nodeID),
			collectionName,
			docID,
			action.Relation,
			getIdentityDID(s, action.TargetIdentity),
		)

		expectedErrorRaised := AssertError(s.t, s.testCase.Description, err, action.ExpectedError)
		assertExpectedErrorRaised(s.t, s.testCase.Description, action.ExpectedError, expectedErrorRaised)

		if !expectedErrorRaised {
			require.Equal(s.t, action.ExpectedError, "")
			require.Equal(s.t, action.ExpectedExistence, exists.ExistedAlready)
		}

		// The relationship should only be added to a SourceHub chain once - there is no need to loop through
		// the nodes.
		if documentACPType == SourceHubDocumentACPType {
			actionNodeID = immutable.Some(0)
			break
		}
	}

	if action.ExpectedError == "" && !action.ExpectedExistence {
		expect := map[string]struct{}{
			docID: {},
		}

		waitForUpdateEvents(s, actionNodeID, action.CollectionID, expect, action.TargetIdentity)
	}
}

// DeleteDocActorRelationship will attempt to delete a relationship between a document and an actor.
type DeleteDocActorRelationship struct {
	// NodeID may hold the ID (index) of the node we want to delete doc actor relationship on.
	//
	// If a value is not provided the relationship will be deleted on all nodes, unless testing with
	// sourcehub document ACP, in which case the relationship will only be deleted once.
	NodeID immutable.Option[int]

	// The collection in which the target document we want to delete relationship for exists.
	//
	// This is a required field. To test the invalid usage of not having this arg, use -1 index.
	CollectionID int

	// The index-identifier of the document within the collection.  This is based on
	// the order in which it was created, not the ordering of the document within the
	// database.
	//
	// This is a required field. To test the invalid usage of not having this arg, use -1 index.
	DocID int

	// The name of the relation within the relationship we want to delete (should be defined in the policy).
	//
	// This is a required field.
	Relation string

	// The target public identity, i.e. the identity of the actor with whom the relationship is with.
	//
	// This is a required field. To test the invalid usage of not having this arg, use NoIdentity() or leave default.
	TargetIdentity immutable.Option[Identity]

	// The requestor identity, i.e. identity of the actor deleting the relationship.
	// Note: This identity must either own or have managing access defined in the policy.
	//
	// This is a required field. To test the invalid usage of not having this arg, use NoIdentity() or leave default.
	RequestorIdentity immutable.Option[Identity]

	// Result returns true if the relationship record was expected to be found and deleted,
	// and returns false if no matching relationship record was found (no-op).
	ExpectedRecordFound bool

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

func deleteDocActorRelationshipACP(
	s *state,
	action DeleteDocActorRelationship,
) {
	nodeIDs, nodes := getNodesWithIDs(action.NodeID, s.nodes)
	for index, node := range nodes {
		nodeID := nodeIDs[index]

		collectionName, docID := getCollectionAndDocInfo(s, action.CollectionID, action.DocID, nodeID)

		deleteDocActorRelationshipResult, err := node.DeleteDocActorRelationship(
			getContextWithIdentity(s.ctx, s, action.RequestorIdentity, nodeID),
			collectionName,
			docID,
			action.Relation,
			getIdentityDID(s, action.TargetIdentity),
		)

		expectedErrorRaised := AssertError(s.t, s.testCase.Description, err, action.ExpectedError)
		assertExpectedErrorRaised(s.t, s.testCase.Description, action.ExpectedError, expectedErrorRaised)

		if !expectedErrorRaised {
			require.Equal(s.t, action.ExpectedError, "")
			require.Equal(s.t, action.ExpectedRecordFound, deleteDocActorRelationshipResult.RecordFound)
		}

		// The relationship should only be added to a SourceHub chain once - there is no need to loop through
		// the nodes.
		if documentACPType == SourceHubDocumentACPType {
			break
		}
	}
}

func getCollectionAndDocInfo(s *state, collectionID, docInd, nodeID int) (string, string) {
	collectionName := ""
	docID := ""
	if collectionID != -1 {
		collection := s.nodes[nodeID].collections[collectionID]
		if collection.Version().Name == "" {
			require.Fail(s.t, "Expected non-empty collection name, but it was empty.", s.testCase.Description)
		}
		collectionName = collection.Version().Name

		if docInd != -1 {
			docID = s.docIDs[collectionID][docInd].String()
		}
	}
	return collectionName, docID
}

func setupSourceHub(s *state) ([]node.DocumentACPOpt, error) {
	var isDocumentACPTest bool
	for _, a := range s.testCase.Actions {
		switch a.(type) {
		case
			AddDocPolicy,
			AddDocActorRelationship,
			DeleteDocActorRelationship:
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

	return l.Addr().(*net.TCPAddr).Port,
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
