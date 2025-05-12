// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cwrap

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	cbindings "github.com/sourcenetwork/defradb/cbindings/logic"
	"github.com/sourcenetwork/defradb/client"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/event"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/lens/host-go/config/model"
)

var wrapperCount int32 = 0
var _ client.TxnStore = (*CWrapper)(nil)
var _ client.P2P = (*CWrapper)(nil)

type CWrapper struct {
	nodeNum int
}

func NewCWrapper(ctx context.Context, enableNAC bool) *CWrapper {
	identityPrivateKey := identityFromContext(ctx)
	nodeNum := atomic.AddInt32(&wrapperCount, 1) - 1
	setupTests(int(nodeNum), identityPrivateKey, enableNAC)
	return &CWrapper{nodeNum: int(nodeNum)}
}

func (w *CWrapper) PeerInfo() peer.AddrInfo {
	result := cbindings.P2PInfo(w.nodeNum)

	if result.Status != 0 {
		return peer.AddrInfo{}
	}

	addrInfo, err := unmarshalResult[peer.AddrInfo](result.Value)
	if err != nil {
		return peer.AddrInfo{}
	}
	return addrInfo
}

func (w *CWrapper) SetReplicator(ctx context.Context, info peer.AddrInfo, collections ...string) error {
	txnID := txnIDFromContext(ctx)
	peerStr := info.String()
	colStr := strings.Join(collections, ",")

	result := cbindings.P2PsetReplicator(w.nodeNum, colStr, peerStr, txnID)

	if result.Status != 0 {
		return errors.New(result.Error)
	}
	return nil
}

func (w *CWrapper) DeleteReplicator(ctx context.Context, info peer.AddrInfo, collections ...string) error {
	txnID := txnIDFromContext(ctx)
	peerStr := info.String()
	colStr := strings.Join(collections, ",")

	result := cbindings.P2PdeleteReplicator(w.nodeNum, colStr, peerStr, txnID)

	if result.Status != 0 {
		return errors.New(result.Error)
	}
	return nil
}

func (w *CWrapper) GetAllReplicators(ctx context.Context) ([]client.Replicator, error) {
	result := cbindings.P2PgetAllReplicators(w.nodeNum)

	if result.Status != 0 {
		return nil, errors.New(result.Error)
	}

	replicators, err := unmarshalResult[[]client.Replicator](result.Value)
	if err != nil {
		return nil, err
	}
	return replicators, nil
}

func (w *CWrapper) AddP2PCollections(ctx context.Context, collectionIDs ...string) error {
	txnID := txnIDFromContext(ctx)
	colStr := strings.Join(collectionIDs, ",")

	result := cbindings.P2PcollectionAdd(w.nodeNum, colStr, txnID)

	if result.Status != 0 {
		return errors.New(result.Error)
	}
	return nil
}

func (w *CWrapper) RemoveP2PCollections(ctx context.Context, collectionIDs ...string) error {
	txnID := txnIDFromContext(ctx)
	colStr := strings.Join(collectionIDs, ",")

	result := cbindings.P2PcollectionRemove(w.nodeNum, colStr, txnID)

	if result.Status != 0 {
		return errors.New(result.Error)
	}
	return nil
}

func (w *CWrapper) GetAllP2PCollections(ctx context.Context) ([]string, error) {
	txnID := txnIDFromContext(ctx)
	result := cbindings.P2PcollectionGetAll(w.nodeNum, txnID)

	if result.Status != 0 {
		return nil, errors.New(result.Error)
	}

	collections, err := unmarshalResult[[]string](result.Value)
	if err != nil {
		return nil, err
	}
	return collections, nil
}

func (w *CWrapper) AddP2PDocuments(ctx context.Context, docIDs ...string) error {
	txnID := txnIDFromContext(ctx)
	docStr := strings.Join(docIDs, ",")

	result := cbindings.P2PdocumentAdd(w.nodeNum, docStr, txnID)

	if result.Status != 0 {
		return errors.New(result.Error)
	}
	return nil
}

func (w *CWrapper) RemoveP2PDocuments(ctx context.Context, docIDs ...string) error {
	txnID := txnIDFromContext(ctx)
	docStr := strings.Join(docIDs, ",")

	result := cbindings.P2PdocumentRemove(w.nodeNum, docStr, txnID)

	if result.Status != 0 {
		return errors.New(result.Error)
	}
	return nil
}

func (w *CWrapper) GetAllP2PDocuments(ctx context.Context) ([]string, error) {
	txnID := txnIDFromContext(ctx)
	result := cbindings.P2PdocumentGetAll(w.nodeNum, txnID)

	if result.Status != 0 {
		return nil, errors.New(result.Error)
	}

	docs, err := unmarshalResult[[]string](result.Value)
	if err != nil {
		return nil, err
	}
	return docs, nil
}

func (w *CWrapper) SyncDocuments(
	ctx context.Context,
	collectionName string,
	docIDs []string,
) error {
	txnID := txnIDFromContext(ctx)
	docs := strings.Join(docIDs, ",")
	deadline, hasDeadline := ctx.Deadline()
	timerStr := ""
	if hasDeadline {
		timerStr = time.Until(deadline).String()
	}
	result := cbindings.P2PdocumentSync(w.nodeNum, collectionName, docs, txnID, timerStr)
	if result.Status != 0 {
		return errors.New(result.Error)
	}
	return nil
}

func (w *CWrapper) BasicImport(ctx context.Context, filepath string) error {
	panic("not implemented")
}

func (w *CWrapper) BasicExport(ctx context.Context, config *client.BackupConfig) error {
	panic("not implemented")
}

func (w *CWrapper) AddSchema(ctx context.Context, schema string) ([]client.CollectionVersion, error) {
	txnID := txnIDFromContext(ctx)
	result := cbindings.AddSchema(w.nodeNum, schema, txnID)

	if result.Status != 0 {
		return nil, errors.New(result.Error)
	}

	collectionVersions, err := unmarshalResult[[]client.CollectionVersion](result.Value)
	if err != nil {
		return nil, err
	}
	return collectionVersions, nil
}

func (w *CWrapper) AddDACPolicy(
	ctx context.Context,
	policy string,
) (client.AddPolicyResult, error) {
	txnID := txnIDFromContext(ctx)
	identity := identityFromContext(ctx)

	result := cbindings.ACPAddDACPolicy(w.nodeNum, identity, policy, txnID)

	if result.Status != 0 {
		return client.AddPolicyResult{}, errors.New(result.Error)
	}

	addPolicyRes, err := unmarshalResult[client.AddPolicyResult](result.Value)
	if err != nil {
		return client.AddPolicyResult{}, err
	}
	return addPolicyRes, nil
}

func (w *CWrapper) AddDACActorRelationship(
	ctx context.Context,
	collectionName string,
	docID string,
	relation string,
	targetActor string,
) (client.AddActorRelationshipResult, error) {
	txnID := txnIDFromContext(ctx)
	identity := identityFromContext(ctx)

	result := cbindings.ACPAddDACActorRelationship(
		w.nodeNum,
		identity,
		collectionName,
		docID,
		relation,
		targetActor,
		txnID,
	)

	if result.Status != 0 {
		return client.AddActorRelationshipResult{}, errors.New(result.Error)
	}

	// Unmarshall the output from JSON to client.AddActorRelationshipResult
	addRelationshipRes, err := unmarshalResult[client.AddActorRelationshipResult](result.Value)
	if err != nil {
		return client.AddActorRelationshipResult{}, err
	}
	return addRelationshipRes, nil
}

func (w *CWrapper) DeleteDACActorRelationship(
	ctx context.Context,
	collectionName string,
	docID string,
	relation string,
	targetActor string,
) (client.DeleteActorRelationshipResult, error) {
	txnID := txnIDFromContext(ctx)
	identity := identityFromContext(ctx)

	result := cbindings.ACPDeleteDACActorRelationship(
		w.nodeNum,
		identity,
		collectionName,
		docID,
		relation,
		targetActor,
		txnID,
	)

	if result.Status != 0 {
		return client.DeleteActorRelationshipResult{}, errors.New(result.Error)
	}

	deleteRelationshipRes, err := unmarshalResult[client.DeleteActorRelationshipResult](result.Value)
	if err != nil {
		return client.DeleteActorRelationshipResult{}, err
	}
	return deleteRelationshipRes, nil
}

func (w *CWrapper) GetNACStatus(ctx context.Context) (client.NACStatusResult, error) {
	txnID := txnIDFromContext(ctx)
	identity := identityFromContext(ctx)
	result := cbindings.ACPGetNACStatus(w.nodeNum, identity, txnID)
	if result.Status != 0 {
		return client.NACStatusResult{}, errors.New(result.Error)
	}
	return unmarshalResult[client.NACStatusResult](result.Value)
}

func (w *CWrapper) ReEnableNAC(ctx context.Context) error {
	txnID := txnIDFromContext(ctx)
	identity := identityFromContext(ctx)
	result := cbindings.ACPReEnableNAC(w.nodeNum, identity, txnID)
	if result.Status != 0 {
		return errors.New(result.Error)
	}
	return nil
}

func (w *CWrapper) DisableNAC(ctx context.Context) error {
	txnID := txnIDFromContext(ctx)
	identity := identityFromContext(ctx)
	result := cbindings.ACPDisableNAC(w.nodeNum, identity, txnID)
	if result.Status != 0 {
		return errors.New(result.Error)
	}
	return nil
}

func (w *CWrapper) AddNACActorRelationship(
	ctx context.Context,
	relation string,
	targetActor string,
) (client.AddActorRelationshipResult, error) {
	txnID := txnIDFromContext(ctx)
	identity := identityFromContext(ctx)
	result := cbindings.ACPAddNACActorRelationship(w.nodeNum, identity, relation, targetActor, txnID)
	if result.Status != 0 {
		return client.AddActorRelationshipResult{}, errors.New(result.Error)
	}
	return unmarshalResult[client.AddActorRelationshipResult](result.Value)
}

func (w *CWrapper) DeleteNACActorRelationship(
	ctx context.Context,
	relation string,
	targetActor string,
) (client.DeleteActorRelationshipResult, error) {
	txnID := txnIDFromContext(ctx)
	identity := identityFromContext(ctx)
	result := cbindings.ACPDeleteNACActorRelationship(w.nodeNum, identity, relation, targetActor, txnID)
	if result.Status != 0 {
		return client.DeleteActorRelationshipResult{}, errors.New(result.Error)
	}
	return unmarshalResult[client.DeleteActorRelationshipResult](result.Value)
}

func (w *CWrapper) PatchCollection(
	ctx context.Context,
	patch string,
	migration immutable.Option[model.Lens],
) error {
	var opts cbindings.GoCOptions
	opts.TxID = txnIDFromContext(ctx)
	opts.Identity = identityFromContext(ctx)
	opts.Version = ""
	opts.CollectionID = ""
	opts.Name = ""
	opts.GetInactive = 0

	cMigration, migrationErr := optionToString(migration)
	if migrationErr != nil {
		return migrationErr
	}

	result := cbindings.CollectionPatch(w.nodeNum, patch, cMigration, opts)

	if result.Status != 0 {
		return errors.New(result.Error)
	}
	return nil
}

func (w *CWrapper) SetActiveCollectionVersion(ctx context.Context, schemaVersionID string) error {
	txnID := txnIDFromContext(ctx)
	result := cbindings.SetActiveCollection(w.nodeNum, schemaVersionID, txnID)
	if result.Status != 0 {
		return errors.New(result.Error)
	}
	return nil
}

func (w *CWrapper) AddView(
	ctx context.Context,
	query string,
	sdl string,
	transform immutable.Option[model.Lens],
) ([]client.CollectionDefinition, error) {
	txnID := txnIDFromContext(ctx)
	cTransform, err := stringFromLensOption(transform)

	if err != nil {
		return []client.CollectionDefinition{}, err
	}

	result := cbindings.ViewAdd(w.nodeNum, query, sdl, cTransform, txnID)

	if result.Status != 0 {
		return []client.CollectionDefinition{}, errors.New(result.Error)
	}

	colDefRes, err := unmarshalResult[[]client.CollectionDefinition](result.Value)
	if err != nil {
		return []client.CollectionDefinition{}, err
	}
	return colDefRes, nil
}

func (w *CWrapper) RefreshViews(ctx context.Context, opts client.CollectionFetchOptions) error {
	txnID := txnIDFromContext(ctx)
	versionID := stringFromImmutableOptionString(opts.VersionID)
	collectionID := stringFromImmutableOptionString(opts.CollectionID)
	name := stringFromImmutableOptionString(opts.Name)
	var cGetInactive bool = false
	if opts.IncludeInactive.HasValue() {
		if opts.IncludeInactive.Value() {
			cGetInactive = true
		}
	}

	result := cbindings.ViewRefresh(w.nodeNum, name, collectionID, versionID, cGetInactive, txnID)

	if result.Status != 0 {
		return errors.New(result.Error)
	}
	return nil
}

func (w *CWrapper) SetMigration(ctx context.Context, config client.LensConfig) error {
	txnID := txnIDFromContext(ctx)
	src := config.SourceSchemaVersionID
	dst := config.DestinationSchemaVersionID
	lensConfig, err := json.Marshal(config.Lens)
	if err != nil {
		return err
	}
	lens := string(lensConfig)

	result := cbindings.LensSet(w.nodeNum, src, dst, lens, txnID)

	if result.Status != 0 {
		return errors.New(result.Error)
	}
	return nil
}

func (w *CWrapper) LensRegistry() client.LensRegistry {
	return &LensRegistry{}
}

func (w *CWrapper) GetCollectionByName(ctx context.Context, name client.CollectionName) (client.Collection, error) {
	cols, err := w.GetCollections(ctx, client.CollectionFetchOptions{Name: immutable.Some(name)})
	if err != nil {
		return nil, err
	}

	if len(cols) == 0 {
		return nil, fmt.Errorf("collection with name %q not found", name)
	}

	// cols will always have length == 1 here
	return cols[0], nil
}

func (w *CWrapper) GetCollections(
	ctx context.Context,
	options client.CollectionFetchOptions,
) ([]client.Collection, error) {
	txnID := txnIDFromContext(ctx)
	identity := identityFromContext(ctx)

	var name string
	if options.Name.HasValue() {
		name = options.Name.Value()
	} else {
		name = ""
	}

	var version string
	if options.VersionID.HasValue() {
		version = options.VersionID.Value()
	} else {
		version = ""
	}

	var collectionID string
	if options.CollectionID.HasValue() {
		collectionID = options.CollectionID.Value()
	} else {
		collectionID = ""
	}

	var includeInactive int = 0
	if options.IncludeInactive.HasValue() {
		if options.IncludeInactive.Value() {
			includeInactive = 1
		}
	}

	var opts cbindings.GoCOptions
	opts.TxID = txnID
	opts.Version = version
	opts.CollectionID = collectionID
	opts.Name = name
	opts.Identity = identity
	opts.GetInactive = includeInactive

	result := cbindings.CollectionDescribe(w.nodeNum, opts)

	if result.Status != 0 {
		return []client.Collection{}, errors.New(result.Error)
	}

	defs, err := unmarshalResult[[]client.CollectionDefinition](result.Value)
	if err != nil {
		return nil, err
	}

	cols := make([]client.Collection, len(defs))
	for i, def := range defs {
		cols[i] = &Collection{def: def, nodeNum: w.nodeNum}
	}
	return cols, nil
}

func (w *CWrapper) GetAllIndexes(ctx context.Context) (map[client.CollectionName][]client.IndexDescription, error) {
	txnID := txnIDFromContext(ctx)
	colName := ""
	result := cbindings.IndexList(w.nodeNum, colName, txnID)

	if result.Status != 0 {
		return nil, errors.New(result.Error)
	}

	res, err := unmarshalResult[map[client.CollectionName][]client.IndexDescription](result.Value)
	if err != nil {
		return nil, errors.New(result.Error)
	}

	return res, nil
}

func (w *CWrapper) ExecRequest(
	ctx context.Context,
	query string,
	opts ...client.RequestOption,
) *client.RequestResult {
	txnID := txnIDFromContext(ctx)
	identity := identityFromContext(ctx)
	operation, variables, err := extractStringsFromRequestOptions(opts)
	if err != nil {
		return &client.RequestResult{
			GQL: client.GQLResult{
				Errors: []error{err},
			},
		}
	}
	result := cbindings.ExecuteQuery(w.nodeNum, query, identity, txnID, operation, variables)

	if result.Status == 2 {
		id := result.Value
		newchan := wrapSubscriptionAsChannel(ctx, id)
		return &client.RequestResult{
			Subscription: newchan,
		}
	}

	retval := &client.RequestResult{}
	if result.Status != 0 {
		retval.GQL.Errors = append(retval.GQL.Errors, fmt.Errorf("%s", result.Error))
		return retval
	}
	if err := json.Unmarshal([]byte(result.Value), &retval.GQL); err != nil {
		retval.GQL.Errors = append(retval.GQL.Errors, err)
	}
	return retval
}

func (w *CWrapper) NewTxn(ctx context.Context, readOnly bool) (client.Txn, error) {
	var concurrent bool = false

	result := cbindings.TransactionCreate(w.nodeNum, concurrent, readOnly)

	if result.Status != 0 {
		return nil, errors.New(result.Error)
	}

	var data struct {
		ID uint64 `json:"id"`
	}

	err := json.Unmarshal([]byte(result.Value), &data)
	if err != nil {
		return nil, err
	}

	retTxn := getTxnFromHandle(w.nodeNum, data.ID)
	retTxnCast := retTxn.(client.Txn) //nolint:forcetypeassert
	return retTxnCast, nil
}

func (w *CWrapper) NewConcurrentTxn(ctx context.Context, readOnly bool) (client.Txn, error) {
	var concurrent bool = true

	result := cbindings.TransactionCreate(w.nodeNum, concurrent, readOnly)

	if result.Status != 0 {
		return nil, errors.New(result.Error)
	}

	var data struct {
		ID uint64 `json:"id"`
	}

	err := json.Unmarshal([]byte(result.Value), &data)
	if err != nil {
		return nil, err
	}

	retTxn := getTxnFromHandle(w.nodeNum, data.ID)
	retTxnCast := retTxn.(client.Txn) //nolint:forcetypeassert
	return retTxnCast, nil
}

func (w *CWrapper) Close() {
	cbindings.NodeStop(w.nodeNum)
}

func (w *CWrapper) Events() event.Bus {
	return cbindings.GetNode(w.nodeNum).DB.Events()
}

func (w *CWrapper) MaxTxnRetries() int {
	return cbindings.GetNode(w.nodeNum).DB.MaxTxnRetries()
}

func (w *CWrapper) PrintDump(ctx context.Context) error {
	panic("not implemented")
}

func (w *CWrapper) Connect(ctx context.Context, addr peer.AddrInfo) error {
	panic("not implemented")
}

func (w *CWrapper) GetNodeIdentity(ctx context.Context) (immutable.Option[identity.PublicRawIdentity], error) {
	result := cbindings.NodeIdentity(w.nodeNum)

	if result.Status != 0 {
		return immutable.None[identity.PublicRawIdentity](), errors.New(result.Error)
	}

	if result.Value == "Node has no identity assigned to it." {
		return immutable.None[identity.PublicRawIdentity](), nil
	}

	var res identity.PublicRawIdentity
	res, err := unmarshalResult[identity.PublicRawIdentity](result.Value)
	if err != nil {
		return immutable.None[identity.PublicRawIdentity](), err
	}
	return immutable.Some(res), nil
}

func (w *CWrapper) VerifySignature(ctx context.Context, blockCid string, pubKey crypto.PublicKey) error {
	pubKeyStr := pubKey.String()
	keyType := string(pubKey.Type())

	result := cbindings.BlockVerifySignature(w.nodeNum, keyType, pubKeyStr, blockCid)

	if result.Status != 0 {
		return errors.New(result.Error)
	}
	return nil
}
