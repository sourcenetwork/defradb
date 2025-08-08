// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package client

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/lens/host-go/config/model"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/crypto"
)

type CollectionName = string

// TxnStore is the primary public programmatic access point to the local DefraDB instance.
//
// It should be constructed via the [db] package, via the [db.NewDB] function.
type TxnStore interface {
	Store

	// NewTxn returns a new transaction on the root store that may be managed externally.
	//
	// It may be used with other functions in the client package. It is not threadsafe.
	NewTxn(ctx context.Context, readOnly bool) (Txn, error)

	// NewConcurrentTxn returns a new transaction on the root store that may be managed externally.
	//
	// It may be used with other functions in the client package. It is threadsafe and multiple threads/Go routines
	// can safely operate on it concurrently.
	NewConcurrentTxn(ctx context.Context, readOnly bool) (Txn, error)
}

type Store interface {
	// PrintDump logs the entire contents of the rootstore (all the data managed by this DefraDB instance).
	//
	// It is likely unwise to call this on a large database instance.
	PrintDump(ctx context.Context) error

	// AddDACPolicy adds policy to document acp system, if available.
	//
	// If policy was successfully added then a policyID is returned,
	// otherwise if acp system was not available then returns the following error:
	// [client.ErrPolicyAddFailureNoACP]
	//
	// Detects the format of the policy automatically by assuming YAML format if JSON
	// validation fails.
	//
	// Note: A policy can not be added without the creatorID (identity).
	AddDACPolicy(ctx context.Context, policy string) (AddPolicyResult, error)

	// AddDACActorRelationship creates a relationship between document and the target actor.
	//
	// If failure occurs, the result will return an error. Upon success the boolean value will
	// be true if the relationship already existed (no-op), and false if a new relationship was made.
	//
	// Note:
	// - The request actor must either be the owner or manager of the document.
	// - If the target actor arg is "*", then the relationship applies to all actors implicitly.
	AddDACActorRelationship(
		ctx context.Context,
		collectionName string,
		docID string,
		relation string,
		targetActor string,
	) (AddActorRelationshipResult, error)

	// DeleteDACActorRelationship deletes a relationship between document and the target actor.
	//
	// If failure occurs, the result will return an error. Upon success the boolean value will
	// be true if the relationship record was found and deleted. Upon success the boolean value
	// will be false if the relationship record was not found (no-op).
	//
	// Note:
	// - The request actor must either be the owner or manager of the document.
	// - If the target actor arg is "*", then the implicitly added relationship with all actors is
	//   removed, however this does not revoke access from actors that had explicit relationships.
	DeleteDACActorRelationship(
		ctx context.Context,
		collectionName string,
		docID string,
		relation string,
		targetActor string,
	) (DeleteActorRelationshipResult, error)

	// AddNACActorRelationship creates a relationship to grant node access to the target actor.
	//
	// If failure occurs, the result will return an error. Upon success the boolean value will
	// be true if the relationship already existed (no-op), and false if a new relationship was made.
	//
	// Note:
	// - The request actor must either be the owner or manager of the document.
	// - If the target actor arg is "*", then the relationship applies to all actors implicitly.
	AddNACActorRelationship(
		ctx context.Context,
		relation string,
		targetActor string,
	) (AddActorRelationshipResult, error)

	// DeleteNACActorRelationship deletes a relationship to revoke node access from target actor.
	//
	// If failure occurs, the result will return an error. Upon success the boolean value will
	// be true if the relationship record was found and deleted. Upon success the boolean value
	// will be false if the relationship record was not found (no-op).
	//
	// Note:
	// - The request actor must either be the owner or manager of the document.
	// - If the target actor arg is "*", then the implicitly added relationship with all actors is
	//   removed, however this does not revoke access from actors that had explicit relationships.
	DeleteNACActorRelationship(
		ctx context.Context,
		relation string,
		targetActor string,
	) (DeleteActorRelationshipResult, error)

	// ReEnableNAC will re-enable node acp that was temporarily disabled (and configured). This will
	// recover previously saved nac state with all the relationships formed.
	//
	// If node acp is already enabled, then returns an error reflecting that it is already enabled.
	//
	// If node acp is not already configured or the previous state was purged then this will return an error,
	// as the user must use the node's start command to configure/enable the node acp the first time.
	//
	// Returns an [client.ErrNotAuthorizedToPerformOperation] error if the requesting identity is not
	// authorized to perform this operation.
	ReEnableNAC(ctx context.Context) error

	// DisableNAC will disable node acp for the users temporarily. This will keep the current node acp
	// state saved so that if it is re-enabled in the future, then we can recover all the relationships formed.
	//
	// If node acp is already disabled, then returns an error reflecting that it is already disabled.
	//
	// If node acp is not already configured or the previous state was purged then this will return an error.
	//
	// Returns an [client.ErrNotAuthorizedToPerformOperation] error if the requesting identity is not
	// authorized to perform this operation.
	DisableNAC(ctx context.Context) error

	// GetNACStatus returns the node acp status that tells us if node access was ever configured,
	// or if node acp is currently enabled or temporarily disabled.
	GetNACStatus(ctx context.Context) (NACStatusResult, error)

	// GetNodeIdentity returns the identity of the node.
	GetNodeIdentity(ctx context.Context) (immutable.Option[identity.PublicRawIdentity], error)

	// VerifySignature verifies the signatures of a block using a public key.
	// Returns an error if any signature verification fails.
	VerifySignature(ctx context.Context, blockCid string, pubKey crypto.PublicKey) error

	// AddSchema takes the provided GQL schema in SDL format, and applies it to the [Store],
	// creating the necessary collections, request types, etc.
	//
	// All schema types provided must not exist prior to calling this, and they may not reference existing
	// types previously defined.
	AddSchema(ctx context.Context, sdl string) ([]CollectionVersion, error)

	// PatchSchema takes the given JSON patch string and applies it to the set of SchemaDescriptions
	// present in the database.
	//
	// If true is provided, the new schema versions will be made active and previous versions deactivated, otherwise
	// [SetActiveSchemaVersion] should be called to do so.
	//
	// It will also update the GQL types used by the query system. It will error and not apply any of the
	// requested, valid updates should the net result of the patch result in an invalid state.  The
	// individual operations defined in the patch do not need to result in a valid state, only the net result
	// of the full patch.
	//
	// The collections (including the schema version ID) will only be updated if any changes have actually
	// been made, if the net result of the patch matches the current persisted description then no changes
	// will be applied.
	//
	// Field [FieldKind] values may be provided in either their raw integer form, or as string as per
	// [FieldKindStringToEnumMapping].
	//
	// A lens configuration may also be provided, it will be added to all collections using the schema.
	PatchSchema(ctx context.Context, patch string, migration immutable.Option[model.Lens], setDefault bool) error

	// PatchCollection takes the given JSON patch string and applies it to the set of CollectionVersions
	// present in the database.
	//
	// It will also update the GQL types used by the query system. It will error and not apply any of the
	// requested, valid updates should the net result of the patch result in an invalid state.  The
	// individual operations defined in the patch do not need to result in a valid state, only the net result
	// of the full patch.
	//
	// Currently only the collection name can be modified.
	PatchCollection(ctx context.Context, patch string) error

	// SetActiveSchemaVersion activates all collection versions with the given schema version, and deactivates all
	// those without it (if they share the same schema root).
	//
	// This will affect all operations interacting with the schema where a schema version is not explicitly
	// provided.  This includes GQL queries and Collection operations.
	//
	// It will return an error if the provided schema version ID does not exist.
	SetActiveSchemaVersion(ctx context.Context, version string) error

	// AddView creates a new Defra View.
	//
	// It takes a GQL query string, for example:
	//
	// Author {
	//	 name
	//	 books {
	//	   name
	//	 }
	// }
	//
	//
	// A GQL SDL that matches its output type must also be provided.  There can only be one `type` declaration,
	// any nested objects must be declared as embedded/schema-only types using the `interface` keyword.
	// Relations must only be specified on the parent side of the relationship.  For example:
	//
	// type AuthorView {
	//   name: String
	//   books: [BookView]
	// }
	// interface BookView {
	//   name: String
	// }
	//
	// It will return the collection definitions of the types defined in the SDL if successful, otherwise an error
	// will be returned.  This function does not execute the given query.
	//
	// Optionally, a lens transform configuration may also be provided - it will execute after the query has run.
	// The transform is not limited to just transforming the input documents, it may also yield new ones, or filter out
	// those passed in from the underlying query.
	AddView(
		ctx context.Context,
		gqlQuery string,
		sdl string,
		transform immutable.Option[model.Lens],
	) ([]CollectionDefinition, error)

	// RefreshViews refreshes the caches of all views matching the given options.  If no options are set, all views
	// will be refreshed.
	//
	// The cached result is dependent on the ACP settings of the source data and the permissions of the user making
	// the call.  At the moment only one cache can be active at a time, so please pay attention to access rights
	// when making this call.
	RefreshViews(ctx context.Context, options CollectionFetchOptions) error

	// SetMigration sets the migration for all collections using the given source-destination schema version IDs.
	//
	// There may only be one migration per collection version.  If another migration was registered it will be
	// overwritten by this migration.
	//
	// Neither of the schema version IDs specified in the configuration need to exist at the time of calling.
	// This is to allow the migration of documents of schema versions unknown to the local node received by the
	// P2P system.
	//
	// Migrations will only run if there is a complete path from the document schema version to the latest local
	// schema version.
	SetMigration(ctx context.Context, config LensConfig) error

	// LensRegistry returns the LensRegistry in use by this database instance.
	//
	// It exposes several useful thread-safe migration related functions.
	LensRegistry() LensRegistry

	// GetCollectionByName attempts to retrieve a collection matching the given name.
	//
	// If no matching collection is found an error will be returned.
	//
	// If a transaction was explicitly provided to this [Store] via [DB].[WithTxn], any function calls
	// made via the returned [Collection] will respect that transaction.
	GetCollectionByName(ctx context.Context, name CollectionName) (Collection, error)

	// GetCollections returns all collections and their descriptions matching the given options
	// that currently exist within this [Store].
	//
	// Inactive collections are not returned by default unless a specific schema version ID
	// is provided.
	//
	// If a transaction was explicitly provided to this [Store] via [DB].[WithTxn], any function calls
	// made via the returned [Collection]s will respect that transaction.
	GetCollections(ctx context.Context, options CollectionFetchOptions) ([]Collection, error)

	// GetSchemaByVersionID returns the schema description for the schema version of the
	// ID provided.
	//
	// Will return an error if it is not found.
	GetSchemaByVersionID(ctx context.Context, versionID string) (SchemaDescription, error)

	// GetSchemas returns all schema versions that currently exist within
	// this [Store].
	GetSchemas(ctx context.Context, options SchemaFetchOptions) ([]SchemaDescription, error)

	// GetAllIndexes returns all the indexes that currently exist within this [Store].
	GetAllIndexes(ctx context.Context) (map[CollectionName][]IndexDescription, error)

	// ExecRequest executes the given GQL request against the [Store].
	ExecRequest(ctx context.Context, request string, opts ...RequestOption) *RequestResult

	// BasicImport imports a json dataset.
	// filepath must be accessible to the node.
	BasicImport(ctx context.Context, filepath string) error

	// BasicExport exports the current data or subset of data to file in json format.
	BasicExport(ctx context.Context, config *BackupConfig) error
}

// Txn is a Store instance that has been wrapped by a transaction.
//
// It privides access to all the Store methods and ensures that they are
// executed under the transaction.
type Txn interface {
	Store

	// ID returns the unique immutable identifier for this transaction.
	ID() uint64

	// Commit finalizes a transaction, attempting to commit it to the Datastore.
	// May return an error if the transaction has gone stale. The presence of an
	// error is an indication that the data was not committed to the Datastore.
	Commit(ctx context.Context) error

	// Discard throws away changes recorded in a transaction without committing
	// them to the underlying Datastore. Any calls made to Discard after Commit
	// has been successfully called will have no effect on the transaction and
	// state of the Datastore, making it safe to defer.
	Discard(ctx context.Context)
}

// GQLOptions contains optional arguments for GQL requests.
type GQLOptions struct {
	// OperationName is the name of the operation to exec.
	OperationName string `json:"operationName"`
	// Variables is a map of names to varible values.
	Variables map[string]any `json:"variables"`
}

// RequestOption sets an optional request setting.
type RequestOption func(*GQLOptions)

// WithOperationName sets the operation name for a GQL request.
func WithOperationName(operationName string) RequestOption {
	return func(o *GQLOptions) {
		o.OperationName = operationName
	}
}

// WithVariables sets the variables for a GQL request.
func WithVariables(variables map[string]any) RequestOption {
	return func(o *GQLOptions) {
		o.Variables = variables
	}
}

// GQLResult represents the immediate results of a GQL request.
//
// It does not handle subscription channels. This object and its children are json serializable.
type GQLResult struct {
	// Errors contains any errors generated whilst attempting to execute the request.
	//
	// If there are values in this slice the request will likely not have run to completion
	// and [Data] will be nil.
	Errors []error `json:"errors,omitempty"`

	// Data contains the resultant data produced by the GQL request.
	//
	// It will be nil if any errors were raised during execution.
	Data any `json:"data"`
}

// gqlError represents an error that was encountered during a GQL request.
//
// This is only used for marshalling to keep our responses spec compliant.
type gqlError struct {
	// Message contains a description of the error.
	Message string `json:"message"`
}

// gqlResult is used to marshal and unmarshal GQLResults.
//
// The serialized data should always match the graphQL spec.
type gqlResult struct {
	// Errors contains the formatted result errors
	Errors []gqlError `json:"errors,omitempty"`
	// Data contains the result data
	Data any `json:"data"`
}

func (res *GQLResult) UnmarshalJSON(data []byte) error {
	dec := json.NewDecoder(bytes.NewBuffer(data))
	dec.UseNumber()
	var out gqlResult
	if err := dec.Decode(&out); err != nil {
		return err
	}
	res.Data = out.Data
	res.Errors = make([]error, len(out.Errors))
	for i, e := range out.Errors {
		res.Errors[i] = ReviveError(e.Message)
	}
	return nil
}

func (res GQLResult) MarshalJSON() ([]byte, error) {
	out := gqlResult{Data: res.Data}
	out.Errors = make([]gqlError, len(res.Errors))
	for i, e := range res.Errors {
		out.Errors[i] = gqlError{Message: e.Error()}
	}
	return json.Marshal(out)
}

// RequestResult represents the results of a GQL request.
type RequestResult struct {
	// GQL contains the immediate results of the GQL request.
	GQL GQLResult

	// Subscription is an optional channel which returns results
	// from a subscription request.
	Subscription <-chan GQLResult
}

// CollectionFetchOptions represents a set of options used for fetching collections.
type CollectionFetchOptions struct {
	// If provided, only collections with this version id will be returned.
	VersionID immutable.Option[string]

	// If provided, only collections with this CollectionID will be returned.
	CollectionID immutable.Option[string]

	// If provided, only collections with this name will be returned.
	Name immutable.Option[string]

	// If IncludeInactive is true, then inactive collections will also be returned.
	IncludeInactive immutable.Option[bool]
}

// SchemaFetchOptions represents a set of options used for fetching schemas.
type SchemaFetchOptions struct {
	// If provided, only schemas of this root will be returned.
	Root immutable.Option[string]

	// If provided, only schemas with this name will be returned.
	Name immutable.Option[string]

	// If provided, only the schema with this id will be returned.
	ID immutable.Option[string]
}
