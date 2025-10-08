// Copyright 2024 Democratized Data Foundation
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
	"reflect"

	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/lens/host-go/config/model"

	"github.com/sourcenetwork/defradb/client/request"
)

// OrphanCollectionID represents an orphan Collection.
//
// Some actions may result in CollectionVersions being defined in an oprhaned state,
// such as registering Lens migrations for version(s) that do not yet exist locally.
//
// Orphaned collections cannot be queried.
const OrphanCollectionID string = "OrphanCollectionID"

// CollectionVersion describes a Collection and all its associated metadata.
type CollectionVersion struct {
	// Name contains the name of the collection.
	Name string

	// The immutable VersionID of this collection version.
	VersionID string

	// The immutable ID of this collection, consistent across all versions.
	CollectionID string

	// CollectionSet contains the information required to identify a collection as part of
	// a larger set.
	//
	// These are global, deterministic properties that, like CollectionID and VersionID, are common across all
	// Defra nodes hosting the collection.
	//
	// Collections only form a collection set if, at the time of their creation, they form a circular set of relations -
	// for example if the Book collection contains a primary relation to Author, and Author contains a primary relation
	// to Book.
	//
	// If this CollectionVersion is not part of a collection set, this property will be None.
	CollectionSet immutable.Option[CollectionSetDescription]

	// Query may hold a query, along with a Lens transform to source data from.
	//
	// If a value is provided, this Collection may not be directly written too,
	// and it may not (yet) have its documents synced across the P2P network.
	Query immutable.Option[QuerySource]

	// PreviousVersion may hold the path details to the previous collection version.
	//
	// If it is None, this is either the first version, or this is an orphaned version
	// created by setting a migration from a collection version not yet known locally.
	PreviousVersion immutable.Option[CollectionSource]

	// Fields contains the fields local to the node within this Collection.
	//
	// Most fields defined here will also be present on the [SchemaDescription]. A notable
	// exception to this are the fields of the (optional) secondary side of a relation
	// which are local only, and will not be present on the [SchemaDescription].
	Fields []CollectionFieldDescription

	// Indexes contains the secondary indexes that this Collection has.
	Indexes []IndexDescription

	// EncryptedIndexes contains the encrypted indexes that this Collection has.
	EncryptedIndexes []EncryptedIndexDescription

	// Policy contains the policy information on this collection.
	//
	// It is possible for a collection to not have a policy, a collection
	// without a policy has no access control.
	//
	// Note: The policy information must be validated using acp right after
	// parsing is done, to avoid storing an invalid policyID or policy resource
	// that may not even exist on acp.
	Policy immutable.Option[PolicyDescription]

	// IsActive defines whether this version of the collection is active or not.
	//
	// The active version will be used when accessed via various functions/endpoints,
	// such as GQL.
	//
	// Only one version can be active at a time.
	IsActive bool

	// IsMaterialized defines whether the items in this collection are cached or not.
	//
	// If it is true, they will be, if false, the data returned on query will be calculated
	// at query-time from source.
	//
	// At the moment this can only be set to `false` if this collection sources its data from
	// another collection/query (is a View).
	IsMaterialized bool

	// IsBranchable defines whether the history of this collection is tracked as a single,
	// verifiable entity.
	//
	// If set to `true` any change to the contents of this set will be linked to a collection
	// level commit via the document(s) composite commit.
	//
	// This enables multiple nodes to verify that they have the same state/history.
	//
	// The history may be queried like a document history can be queried, for example via 'commits'
	// GQL queries.
	//
	// Currently this property is immutable and can only be set on collection creation, however
	// that will change in the future.
	IsBranchable bool

	// IsEmbeddedOnly defines whether this collection exists only as a child object embedded within
	// another collection or not.
	//
	// If true, it will not be directly queriable.
	IsEmbeddedOnly bool

	// IsPlaceholder defines whether or not this collection version is an empty placeholder waiting
	// to be defined in this Defra node.
	//
	// This can happen if a migration between version ids is defined locally before the version (for
	// example, via PatchCollection).
	IsPlaceholder bool

	// VectorEmbeddings contains the configuration for generating embedding vectors.
	//
	// This is only usable with array fields.
	//
	// When configured, embeddings may call 3rd party APIs inline with document mutations.
	// This may cause increase latency in the completion of the mutation requests.
	// This is necessary to ensure that the generated docID is representative of the
	// content of the document.
	VectorEmbeddings []VectorEmbeddingDescription
}

// CollectionSetDescription contains the information required to identify a collection as part of
// a larger set.
//
// These are global, deterministic properties that, like CollectionID and VersionID, are common across all
// Defra nodes hosting the collection.
//
// Collections only form a collection set if, at the time of their creation, they form a circular set of relations -
// for example if the Book collection contains a primary relation to Author, and Author contains a primary relation
// to Book.
type CollectionSetDescription struct {
	// CollectionSetID is the ID of the collection set that this item belongs to.
	CollectionSetID string

	// RelativeID is this item's relative location within the collection set.
	//
	// This is currently based on Name, lexographically ascending, at the time of creation.
	RelativeID int
}

// QuerySource represents a collection data source from a query.
//
// The query will be executed when data from this source is requested, and the query results
// yielded to the consumer.
type QuerySource struct {
	// Query contains the base query of this data source.
	Query request.Select

	// Transform is a optional Lens configuration.  If specified, data drawn from the [Query] will have the
	// transform applied before being returned.
	//
	// The transform is not limited to just transforming the input documents, it may also yield new ones, or filter out
	// those passed in from the underlying query.
	Transform immutable.Option[model.Lens]
}

// CollectionSource represents a collection data source from another collection instance.
//
// Data against all collection instances in a CollectionSource chain will be returned as-if
// from the same dataset when queried.  Lens transforms may be applied between instances.
//
// Typically these are used to link together multiple schema versions into the same dataset.
type CollectionSource struct {
	// SourceCollectionID is the local identifier of the source [CollectionVersion] from which to
	// share data.
	//
	// This is a bi-directional relationship, and documents in the host collection instance will also
	// be available to the source collection instance.
	SourceCollectionID string

	// Transform is a optional Lens configuration.  If specified, data drawn from the source will have the
	// transform applied before being returned by any operation on the host collection instance.
	//
	// If the transform supports an inverse operation, that inverse will be applied when the source collection
	// draws data from this host.
	Transform immutable.Option[model.Lens]
}

// GetFieldByName returns the field for the given field name. If such a field is found it
// will return it and true, if it is not found it will return false.
func (col CollectionVersion) GetFieldByName(fieldName string) (CollectionFieldDescription, bool) {
	for _, field := range col.Fields {
		if field.Name == fieldName {
			return field, true
		}
	}
	return CollectionFieldDescription{}, false
}

// GetFieldByRelation returns the field that supports the relation of the given name.
func (col CollectionVersion) GetFieldByRelation(
	relationName string,
	otherCollectionName string,
	otherFieldName string,
) (CollectionFieldDescription, bool) {
	for _, field := range col.Fields {
		if field.RelationName.Value() == relationName &&
			!(col.Name == otherCollectionName && otherFieldName == field.Name) &&
			field.Kind != FieldKind_DocID {
			return field, true
		}
	}
	return CollectionFieldDescription{}, false
}

// Equal returns true if this and the given [CollectionVersion] are equal.
func (col CollectionVersion) Equal(other CollectionVersion) bool {
	return reflect.DeepEqual(col, other)
}

// VectorEmbeddingDescription hold the relevant information to generate embeddings.
//
// Embeddings are AI/ML specific vector representations of some content.
// In the case of DefraDB, that content is one or multiple fields, optionally added to a template.
type VectorEmbeddingDescription struct {
	// FieldName is the name of the field on the collection that this embedding description applies to.
	FieldName string
	// Fields are the fields in the parent schema that will be used as the basis of the
	// vector generation.
	Fields []string
	// Model is the LLM of the provider to use for generating the embeddings.
	// For example: text-embedding-3-small
	Model string
	// Provider is the API provider to use for generating the embeddings.
	// For example: openai
	Provider string
	// (Optional) Template is the local path of the template to use with the
	// field values to form the content to send to the model.
	//
	// For example, with the following schema,
	// ```
	// type User {
	//   name: String
	//   age: Int
	//   name_about_v: [Float32!] @embedding(fields: ["name", "age"], ...)
	// }
	// ````
	// we can define the following Go template.
	// ```
	// {{ .name }} is {{ .age }} years old.
	// ```
	Template string
	// URL is the url enpoint of the provider's API.
	// For example: https://api.openai.com/v1
	//
	// Not providing a URL will result in the use of the default
	// known URL for the given provider.
	URL string
}

// IsSupportedVectorEmbeddingSourceKind return true if the fields used for embedding generation
// are of supported type.
//
// Currently, the supported types are Float32, Float64, Int and String
func IsSupportedVectorEmbeddingSourceKind(fieldKind FieldKind) bool {
	switch fieldKind {
	case FieldKind_NILLABLE_FLOAT32, FieldKind_NILLABLE_FLOAT64, FieldKind_NILLABLE_INT, FieldKind_NILLABLE_STRING:
		return true
	default:
		return false
	}
}
