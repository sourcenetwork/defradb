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

	ds "github.com/ipfs/go-datastore"
	"github.com/lens-vm/lens/host-go/config/model"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/event"
)

type CollectionName = string

// DB is the primary public programmatic access point to the local DefraDB instance.
//
// It should be constructed via the [db] package, via the [db.NewDB] function.
type DB interface {
	// Store contains DefraDB functions protected by an internal, short-lived, transaction, allowing safe
	// access to common database read and write operations.
	Store

	// NewTxn returns a new transaction on the root store that may be managed externally.
	//
	// It may be used with other functions in the client package. It is not threadsafe.
	NewTxn(context.Context, bool) (datastore.Txn, error)

	// NewConcurrentTxn returns a new transaction on the root store that may be managed externally.
	//
	// It may be used with other functions in the client package. It is threadsafe and multiple threads/Go routines
	// can safely operate on it concurrently.
	NewConcurrentTxn(context.Context, bool) (datastore.Txn, error)

	// Rootstore returns the underlying root store, within which all data managed by DefraDB is held.
	Rootstore() datastore.Rootstore

	// Blockstore returns the blockstore, within which all blocks (commits) managed by DefraDB are held.
	//
	// It sits within the rootstore returned by [Root].
	Blockstore() datastore.Blockstore

	// Encstore returns the store, that contains all known encryption keys for documents and their fields.
	//
	// It sits within the rootstore returned by [Root].
	Encstore() datastore.Blockstore

	// Peerstore returns the peerstore where known host information is stored.
	//
	// It sits within the rootstore returned by [Root].
	Peerstore() datastore.DSBatching

	// Headstore returns the headstore where the current heads of the database are stored.
	//
	// It is read-only and sits within the rootstore returned by [Root].
	Headstore() ds.Read

	// Close closes the database instance and releases any resources held.
	//
	// The behaviour of other functions in this package after this function has been called is undefined
	// unless explicitly stated on the function in question.
	//
	// It does not explicitly clear any data from persisted storage, and a new [DB] instance may typically
	// be created after calling this to resume operations on the prior data - this is however dependant on
	// the behaviour of the rootstore provided on database instance creation, as this function will Close
	// the provided rootstore.
	Close()

	// Events returns the database event queue.
	//
	// It may be used to monitor database events - a new event will be yielded for each mutation.
	// Note: it does not copy the queue, just the reference to it.
	Events() *event.Bus

	// MaxTxnRetries returns the number of retries that this DefraDB instance has been configured to
	// make in the event of a transaction conflict in certain scenarios.
	//
	// Currently this is only used within the P2P system and will not affect operations initiated by users.
	MaxTxnRetries() int

	// PrintDump logs the entire contents of the rootstore (all the data managed by this DefraDB instance).
	//
	// It is likely unwise to call this on a large database instance.
	PrintDump(ctx context.Context) error

	// AddPolicy adds policy to acp, if acp is available.
	//
	// If policy was successfully added to acp then a policyID is returned,
	// otherwise if acp was not available then returns the following error:
	// [client.ErrPolicyAddFailureNoACP]
	//
	// Detects the format of the policy automatically by assuming YAML format if JSON
	// validation fails.
	//
	// Note: A policy can not be added without the creatorID (identity).
	AddPolicy(ctx context.Context, policy string) (AddPolicyResult, error)

	// AddDocActorRelationship creates a relationship between document and the target actor.
	//
	// If failure occurs, the result will return an error. Upon success the boolean value will
	// be true if the relationship already existed (no-op), and false if a new relationship was made.
	//
	// Note: The request actor must either be the owner or manager of the document.
	AddDocActorRelationship(
		ctx context.Context,
		collectionName string,
		docID string,
		relation string,
		targetActor string,
	) (AddDocActorRelationshipResult, error)
}

// Store contains the core DefraDB read-write operations.
type Store interface {
	// Backup holds the backup related methods that must be implemented by the database.
	Backup

	// P2P contains functions related to the P2P system.
	//
	// These functions are only useful if there is a configured network peer.
	P2P

	// AddSchema takes the provided GQL schema in SDL format, and applies it to the [Store],
	// creating the necessary collections, request types, etc.
	//
	// All schema types provided must not exist prior to calling this, and they may not reference existing
	// types previously defined.
	AddSchema(context.Context, string) ([]CollectionDescription, error)

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
	PatchSchema(context.Context, string, immutable.Option[model.Lens], bool) error

	// PatchCollection takes the given JSON patch string and applies it to the set of CollectionDescriptions
	// present in the database.
	//
	// It will also update the GQL types used by the query system. It will error and not apply any of the
	// requested, valid updates should the net result of the patch result in an invalid state.  The
	// individual operations defined in the patch do not need to result in a valid state, only the net result
	// of the full patch.
	//
	// Currently only the collection name can be modified.
	PatchCollection(context.Context, string) error

	// SetActiveSchemaVersion activates all collection versions with the given schema version, and deactivates all
	// those without it (if they share the same schema root).
	//
	// This will affect all operations interacting with the schema where a schema version is not explicitly
	// provided.  This includes GQL queries and Collection operations.
	//
	// It will return an error if the provided schema version ID does not exist.
	SetActiveSchemaVersion(context.Context, string) error

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
	RefreshViews(context.Context, CollectionFetchOptions) error

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
	SetMigration(context.Context, LensConfig) error

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
	GetCollectionByName(context.Context, CollectionName) (Collection, error)

	// GetCollections returns all collections and their descriptions matching the given options
	// that currently exist within this [Store].
	//
	// Inactive collections are not returned by default unless a specific schema version ID
	// is provided.
	//
	// If a transaction was explicitly provided to this [Store] via [DB].[WithTxn], any function calls
	// made via the returned [Collection]s will respect that transaction.
	GetCollections(context.Context, CollectionFetchOptions) ([]Collection, error)

	// GetSchemaByVersionID returns the schema description for the schema version of the
	// ID provided.
	//
	// Will return an error if it is not found.
	GetSchemaByVersionID(context.Context, string) (SchemaDescription, error)

	// GetSchemas returns all schema versions that currently exist within
	// this [Store].
	GetSchemas(context.Context, SchemaFetchOptions) ([]SchemaDescription, error)

	// GetAllIndexes returns all the indexes that currently exist within this [Store].
	GetAllIndexes(context.Context) (map[CollectionName][]IndexDescription, error)

	// ExecRequest executes the given GQL request against the [Store].
	ExecRequest(ctx context.Context, request string, opts ...RequestOption) *RequestResult
}

// GQLOptions contains optional arguments for GQL requests.
type GQLOptions struct {
	// OperationName is the name of the operation to exec.
	OperationName string
	// Variables is a map of names to varible values.
	Variables map[string]any
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
	// If provided, only collections with this schema version id will be returned.
	SchemaVersionID immutable.Option[string]

	// If provided, only collections with schemas of this root will be returned.
	SchemaRoot immutable.Option[string]

	// If provided, only collections with this root will be returned.
	Root immutable.Option[uint32]

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
