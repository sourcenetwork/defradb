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
	acpType = ACPType(os.Getenv(acpTypeEnvName))
	if acpType == "" {
		acpType = LocalACPType
	}
}

// AddPolicy will attempt to add the given policy using DefraDB's ACP system.
type AddPolicy struct {
	// NodeID may hold the ID (index) of the node we want to add policy to.
	//
	// If a value is not provided the policy will be added in all nodes, unless testing with
	// sourcehub ACP, in which case the policy will only be defined once.
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

		expectedErrorRaised := AssertError(s.t, s.testCase.Description, err, action.ExpectedError)
		assertExpectedErrorRaised(s.t, s.testCase.Description, action.ExpectedError, expectedErrorRaised)

		if !expectedErrorRaised {
			require.Equal(s.t, action.ExpectedError, "")
			require.Equal(s.t, action.ExpectedPolicyID, policyResult.PolicyID)
		}

		// The policy should only be added to a SourceHub chain once - there is no need to loop through
		// the nodes.
		if acpType == SourceHubACPType {
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
	// This is a required field. To test the invalid usage of not having this arg, use -1 index.
	TargetIdentity int

	// The requestor identity, i.e. identity of the actor creating the relationship.
	// Note: This identity must either own or have managing access defined in the policy.
	//
	// This is a required field. To test the invalid usage of not having this arg, use -1 index.
	RequestorIdentity int

	// Result returns true if it was a no-op due to existing before, and false if a new relationship was made.
	ExpectedExistence bool

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

func addDocActorRelationshipACP(s *state, action AddDocActorRelationship) {
	processNode := func(nodeID int) {
		node := s.nodes[nodeID]

		collectionName, docID := getCollectionAndDocInfo(s, action.CollectionID, action.DocID, nodeID)
		requestorIdentity := getRequestorIdentity(s, action.RequestorIdentity, nodeID)

		result, err := node.AddDocActorRelationship(
			db.SetContextIdentity(s.ctx, requestorIdentity),
			collectionName,
			docID,
			action.Relation,
			getTargetIdentity(s, action.TargetIdentity, nodeID),
		)

		assertACPResult(s, action.ExpectedError, err, action.ExpectedExistence, result.ExistedAlready, "existed")
	}

	executeACPAction(action.NodeID, processNode, s)
}

// DeleteDocActorRelationship will attempt to delete a relationship between a document and an actor.
type DeleteDocActorRelationship struct {
	// NodeID may hold the ID (index) of the node we want to delete doc actor relationship on.
	//
	// If a value is not provided the relationship will be deleted on all nodes, unless testing with
	// sourcehub ACP, in which case the relationship will only be deleted once.
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
	// This is a required field. To test the invalid usage of not having this arg, use -1 index.
	TargetIdentity int

	// The requestor identity, i.e. identity of the actor deleting the relationship.
	// Note: This identity must either own or have managing access defined in the policy.
	//
	// This is a required field. To test the invalid usage of not having this arg, use -1 index.
	RequestorIdentity int

	// Result returns true if the relationship record was expected to be found and deleted,
	// and returns false if no matching relationship record was found (no-op).
	ExpectedRecordFound bool

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

func deleteDocActorRelationshipACP(s *state, action DeleteDocActorRelationship) {
	processNode := func(nodeID int) {
		node := s.nodes[nodeID]

		collectionName, docID := getCollectionAndDocInfo(s, action.CollectionID, action.DocID, nodeID)
		requestorIdentity := getRequestorIdentity(s, action.RequestorIdentity, nodeID)

		result, err := node.DeleteDocActorRelationship(
			db.SetContextIdentity(s.ctx, requestorIdentity),
			collectionName,
			docID,
			action.Relation,
			getTargetIdentity(s, action.TargetIdentity, nodeID),
		)

		assertACPResult(s, action.ExpectedError, err, action.ExpectedRecordFound, result.RecordFound, "record found")
	}

	executeACPAction(action.NodeID, processNode, s)
}

func executeACPAction(nodeID immutable.Option[int], processNode func(nodeID int), s *state) {
	if nodeID.HasValue() {
		processNode(nodeID.Value())
	} else {
		for i := range getNodes(nodeID, s.nodes) {
			processNode(i)
			if acpType == SourceHubACPType {
				break
			}
		}
	}
}

func getCollectionAndDocInfo(s *state, collectionID, docInd, nodeID int) (string, string) {
	collectionName := ""
	docID := ""
	if collectionID != -1 {
		collection := s.collections[nodeID][collectionID]
		if !collection.Description().Name.HasValue() {
			require.Fail(s.t, "Expected non-empty collection name, but it was empty.", s.testCase.Description)
		}
		collectionName = collection.Description().Name.Value()

		if docInd != -1 {
			docID = s.docIDs[collectionID][docInd].String()
		}
	}
	return collectionName, docID
}

func getTargetIdentity(s *state, targetIdent, nodeID int) string {
	if targetIdent != -1 {
		optionalTargetIdentity := getIdentity(s, nodeID, immutable.Some(targetIdent))
		if !optionalTargetIdentity.HasValue() {
			require.Fail(s.t, "Expected non-empty target identity, but it was empty.", s.testCase.Description)
		}
		return optionalTargetIdentity.Value().DID
	}
	return ""
}

func getRequestorIdentity(s *state, requestorIdent, nodeID int) immutable.Option[acpIdentity.Identity] {
	if requestorIdent != -1 {
		requestorIdentity := getIdentity(s, nodeID, immutable.Some(requestorIdent))
		if !requestorIdentity.HasValue() {
			require.Fail(s.t, "Expected non-empty requestor identity, but it was empty.", s.testCase.Description)
		}
		return requestorIdentity
	}
	return acpIdentity.None
}

func assertACPResult(
	s *state,
	expectedError string,
	actualErr error,
	expectedBool, actualBool bool,
	boolDesc string,
) {
	expectedErrorRaised := AssertError(s.t, s.testCase.Description, actualErr, expectedError)
	assertExpectedErrorRaised(s.t, s.testCase.Description, expectedError, expectedErrorRaised)

	if !expectedErrorRaised {
		require.Equal(s.t, expectedError, "")
		require.Equal(s.t, expectedBool, actualBool, boolDesc)
	}
}

func setupSourceHub(s *state) ([]node.ACPOpt, error) {
	var isACPTest bool
	for _, a := range s.testCase.Actions {
		switch a.(type) {
		case
			AddPolicy,
			AddDocActorRelationship,
			DeleteDocActorRelationship:
			isACPTest = true
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
	// (appart from the p2p port which we just manage here for consistency).
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

// Generate the keys using the index as the seed so that multiple
// runs yield the same private key.  This is important for stuff like
// the change detector.
func generateIdentity(s *state, seedIndex int, nodeIndex int) (acpIdentity.Identity, error) {
	var audience immutable.Option[string]
	switch client := s.nodes[nodeIndex].(type) {
	case *http.Wrapper:
		audience = immutable.Some(strings.TrimPrefix(client.Host(), "http://"))
	case *cli.Wrapper:
		audience = immutable.Some(strings.TrimPrefix(client.Host(), "http://"))
	}

	source := rand.NewSource(int64(seedIndex))
	r := rand.New(source)

	privateKey, err := secp256k1.GeneratePrivateKeyFromRand(r)
	require.NoError(s.t, err)

	identity, err := acpIdentity.FromPrivateKey(
		privateKey,
		authTokenExpiration,
		audience,
		immutable.Some(s.sourcehubAddress),
		// Creating and signing the bearer token is slow, so we skip it if it not
		// required.
		!(acpType == SourceHubACPType || audience.HasValue()),
	)

	return identity, err
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
		identities := make([]acpIdentity.Identity, index.Value()+1)
		// Fill any empty identities up to the index.
		for i := range identities {
			if i < len(nodeIdentities) && nodeIdentities[i] != (acpIdentity.Identity{}) {
				identities[i] = nodeIdentities[i]
				continue
			}
			newIdentity, err := generateIdentity(s, i, nodeIndex)
			require.NoError(s.t, err)
			identities[i] = newIdentity
		}
		s.identities[nodeIndex] = identities
		return immutable.Some(identities[index.Value()])
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
