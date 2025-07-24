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
	"time"

	cbindings "github.com/sourcenetwork/defradb/cbindings/logic"
	"github.com/sourcenetwork/defradb/client"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/event"

	"github.com/lens-vm/lens/host-go/config/model"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sourcenetwork/immutable"
)

var _ client.TxnStore = (*CWrapper)(nil)
var _ client.P2P = (*CWrapper)(nil)

type CWrapper struct{}

func NewCWrapper() *CWrapper {
	setupTests()
	return &CWrapper{}
}

func (w *CWrapper) PeerInfo() peer.AddrInfo {
	result := cbindings.P2PInfo()

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

	result := cbindings.P2PsetReplicator(colStr, peerStr, txnID)

	if result.Status != 0 {
		return errors.New(result.Error)
	}
	return nil
}

func (w *CWrapper) DeleteReplicator(ctx context.Context, info peer.AddrInfo, collections ...string) error {
	txnID := txnIDFromContext(ctx)
	peerStr := info.String()
	colStr := strings.Join(collections, ",")

	result := cbindings.P2PdeleteReplicator(colStr, peerStr, txnID)

	if result.Status != 0 {
		return errors.New(result.Error)
	}
	return nil
}

func (w *CWrapper) GetAllReplicators(ctx context.Context) ([]client.Replicator, error) {
	result := cbindings.P2PgetAllReplicators()

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

	result := cbindings.P2PcollectionAdd(colStr, txnID)

	if result.Status != 0 {
		return errors.New(result.Error)
	}
	return nil
}

func (w *CWrapper) RemoveP2PCollections(ctx context.Context, collectionIDs ...string) error {
	txnID := txnIDFromContext(ctx)
	colStr := strings.Join(collectionIDs, ",")

	result := cbindings.P2PcollectionRemove(colStr, txnID)

	if result.Status != 0 {
		return errors.New(result.Error)
	}
	return nil
}

func (w *CWrapper) GetAllP2PCollections(ctx context.Context) ([]string, error) {
	txnID := txnIDFromContext(ctx)
	result := cbindings.P2PcollectionGetAll(txnID)

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

	result := cbindings.P2PdocumentAdd(docStr, txnID)

	if result.Status != 0 {
		return errors.New(result.Error)
	}
	return nil
}

func (w *CWrapper) RemoveP2PDocuments(ctx context.Context, docIDs ...string) error {
	txnID := txnIDFromContext(ctx)
	docStr := strings.Join(docIDs, ",")

	result := cbindings.P2PdocumentRemove(docStr, txnID)

	if result.Status != 0 {
		return errors.New(result.Error)
	}
	return nil
}

func (w *CWrapper) GetAllP2PDocuments(ctx context.Context) ([]string, error) {
	txnID := txnIDFromContext(ctx)
	result := cbindings.P2PdocumentGetAll(txnID)

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
	result := cbindings.P2PdocumentSync(collectionName, docs, txnID, timerStr)
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
	result := cbindings.AddSchema(schema, txnID)

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

	result := cbindings.ACPAddPolicy(identity, policy, txnID)

	if result.Status != 0 {
		return client.AddPolicyResult{}, errors.New(result.Error)
	}

	// Unmarshall the output from JSON to client.AddPolicyResult
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

	result := cbindings.ACPAddRelationship(identity, collectionName, docID, relation, targetActor, txnID)

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

	result := cbindings.ACPDeleteRelationship(identity, collectionName, docID, relation, targetActor, txnID)

	if result.Status != 0 {
		return client.DeleteActorRelationshipResult{}, errors.New(result.Error)
	}

	// Unmarshall the output from JSON to client.DeleteActorRelationshipResult
	deleteRelationshipRes, err := unmarshalResult[client.DeleteActorRelationshipResult](result.Value)
	if err != nil {
		return client.DeleteActorRelationshipResult{}, err
	}
	return deleteRelationshipRes, nil
}

func (w *CWrapper) PatchSchema(
	ctx context.Context,
	patch string,
	migration immutable.Option[model.Lens],
	setAsDefaultVersion bool,
) error {
	txnID := txnIDFromContext(ctx)
	cMigration, migrationErr := optionToString(migration)

	if migrationErr != nil {
		return migrationErr
	}

	result := cbindings.PatchSchema(patch, cMigration, setAsDefaultVersion, txnID)

	if result.Status != 0 {
		return errors.New(result.Error)
	}
	return nil
}

func (w *CWrapper) PatchCollection(
	ctx context.Context,
	patch string,
) error {
	var opts cbindings.GoCOptions
	opts.TxID = txnIDFromContext(ctx)
	opts.Identity = identityFromContext(ctx)
	opts.Version = ""
	opts.CollectionID = ""
	opts.Name = ""
	opts.GetInactive = 0

	result := cbindings.CollectionPatch(patch, opts)

	if result.Status != 0 {
		return errors.New(result.Error)
	}
	return nil
}

func (w *CWrapper) SetActiveSchemaVersion(ctx context.Context, schemaVersionID string) error {
	txnID := txnIDFromContext(ctx)
	result := cbindings.SetActiveSchema(schemaVersionID, txnID)
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

	result := cbindings.ViewAdd(query, sdl, cTransform, txnID)

	if result.Status != 0 {
		return []client.CollectionDefinition{}, errors.New(result.Error)
	}

	// Unmarshall the output from JSON to []client.CollectionDefinition
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

	result := cbindings.ViewRefresh(name, collectionID, versionID, cGetInactive, txnID)

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

	result := cbindings.LensSet(src, dst, lens, txnID)

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

	result := cbindings.CollectionDescribe(opts)

	if result.Status != 0 {
		return []client.Collection{}, errors.New(result.Error)
	}

	defs, err := unmarshalResult[[]client.CollectionDefinition](result.Value)
	if err != nil {
		return nil, err
	}

	return collectionsFromDefinitions(defs)
}

func (w *CWrapper) GetSchemaByVersionID(ctx context.Context, versionID string) (client.SchemaDescription, error) {
	schemas, err := w.GetSchemas(ctx, client.SchemaFetchOptions{ID: immutable.Some(versionID)})
	if err != nil {
		return client.SchemaDescription{}, err
	}
	return schemas[0], nil
}

func (w *CWrapper) GetSchemas(
	ctx context.Context,
	options client.SchemaFetchOptions,
) ([]client.SchemaDescription, error) {
	txnID := txnIDFromContext(ctx)
	root := stringFromImmutableOptionString(options.Root)
	version := stringFromImmutableOptionString(options.ID)
	name := stringFromImmutableOptionString(options.Name)

	result := cbindings.DescribeSchema(name, root, version, txnID)

	if result.Status != 0 {
		return []client.SchemaDescription{}, errors.New(result.Error)
	}

	res, err := unmarshalResult[[]client.SchemaDescription](result.Value)
	if err != nil {
		return []client.SchemaDescription{}, errors.New(result.Error)
	}
	return res, nil
}

func (w *CWrapper) GetAllIndexes(ctx context.Context) (map[client.CollectionName][]client.IndexDescription, error) {
	txnID := txnIDFromContext(ctx)
	colName := ""
	result := cbindings.IndexList(colName, txnID)

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
	operation, variables := extractCStringsFromRequestOptions(opts)
	result := cbindings.ExecuteQuery(query, identity, txnID, operation, variables)

	if result.Status == 2 {
		id := result.Value
		newchan := WrapSubscriptionAsChannel(id)
		return &client.RequestResult{
			Subscription: newchan,
		}
	}

	// Unmarshal the result into a *client.RequestResult
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

	result := cbindings.TransactionCreate(concurrent, readOnly)

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

	retTxn := getTxnFromHandle(data.ID)
	retTxnCast := retTxn.(client.Txn) //nolint:forcetypeassert
	return retTxnCast, nil
}

func (w *CWrapper) NewConcurrentTxn(ctx context.Context, readOnly bool) (client.Txn, error) {
	var concurrent bool = true

	result := cbindings.TransactionCreate(concurrent, readOnly)

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

	retTxn := getTxnFromHandle(data.ID)
	retTxnCast := retTxn.(client.Txn) //nolint:forcetypeassert
	return retTxnCast, nil
}

func (w *CWrapper) Close() {
	cbindings.NodeStop()
}

func (w *CWrapper) Events() event.Bus {
	return cbindings.GetGlobalNode().DB.Events()
}

func (w *CWrapper) MaxTxnRetries() int {
	return cbindings.GetGlobalNode().DB.MaxTxnRetries()
}

func (w *CWrapper) PrintDump(ctx context.Context) error {
	panic("not implemented")
}

func (w *CWrapper) Connect(ctx context.Context, addr peer.AddrInfo) error {
	panic("not implemented")
}

func (w *CWrapper) GetNodeIdentity(ctx context.Context) (immutable.Option[identity.PublicRawIdentity], error) {
	result := cbindings.NodeIdentity()

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

	result := cbindings.BlockVerifySignature(keyType, pubKeyStr, blockCid)

	if result.Status != 0 {
		return errors.New(result.Error)
	}
	return nil
}
