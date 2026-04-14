// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package collection_version

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// This test covers a bug that was found as part of https://github.com/sourcenetwork/defradb/issues/4710
// The bug has been fixed, but the test remains as coverage of this case is important.
func TestCollectionVersionWith_OneOne_OneMany_SelfRef(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Dev_RC_Domain {
						routes: [Dev_RC_RedirectRoute]
						firstRoute: Dev_RC_RedirectRoute @primary @relation(name: "domain_first_route")
					}

					type Dev_RC_RedirectRoute {
						firstForDomain: Dev_RC_Domain @relation(name: "domain_first_route")

						domain: Dev_RC_Domain
						after: Dev_RC_RedirectRoute
					}
				`,
			},
			&action.GetCollections{
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Dev_RC_Domain",
						IsActive:       true,
						IsMaterialized: true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Kind: client.FieldKind_DocID,
								Typ:  client.NONE_CRDT,
							},
							{
								Name:         "_firstRouteID",
								Kind:         client.FieldKind_DocID,
								Typ:          client.LWW_REGISTER,
								IsPrimary:    true,
								RelationName: immutable.Some("domain_first_route"),
							},
							{
								Name: "firstRoute",
								Kind: &client.SelfKind{
									RelativeID: "1",
								},
								Typ:          client.NONE_CRDT,
								IsPrimary:    true,
								RelationName: immutable.Some("domain_first_route"),
							},
							{
								Name: "routes",
								Kind: &client.SelfKind{
									RelativeID: "1",
									Array:      true,
								},
								Typ:          client.NONE_CRDT,
								RelationName: immutable.Some("dev_rc_domain_dev_rc_redirectroute"),
							},
						},
					},
					{
						Name:           "Dev_RC_RedirectRoute",
						IsActive:       true,
						IsMaterialized: true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Kind: client.FieldKind_DocID,
								Typ:  client.NONE_CRDT,
							},
							{
								Name:         "_afterID",
								Kind:         client.FieldKind_DocID,
								Typ:          client.LWW_REGISTER,
								IsPrimary:    true,
								RelationName: immutable.Some("dev_rc_redirectroute_dev_rc_redirectroute"),
							},
							{
								Name:         "_domainID",
								Kind:         client.FieldKind_DocID,
								Typ:          client.LWW_REGISTER,
								IsPrimary:    true,
								RelationName: immutable.Some("dev_rc_domain_dev_rc_redirectroute"),
							},
							{
								Name:         "_firstForDomainID",
								Kind:         client.FieldKind_DocID,
								Typ:          client.LWW_REGISTER,
								RelationName: immutable.Some("domain_first_route"),
							},
							{
								Name: "after",
								Kind: &client.SelfKind{
									RelativeID: "1",
								},
								Typ:          client.NONE_CRDT,
								IsPrimary:    true,
								RelationName: immutable.Some("dev_rc_redirectroute_dev_rc_redirectroute"),
							},
							{
								Name: "domain",
								Kind: &client.SelfKind{
									RelativeID: "0",
								},
								Typ:          client.NONE_CRDT,
								IsPrimary:    true,
								RelationName: immutable.Some("dev_rc_domain_dev_rc_redirectroute"),
							},
							{
								Name: "firstForDomain",
								Kind: &client.SelfKind{
									RelativeID: "0",
								},
								Typ:          client.NONE_CRDT,
								RelationName: immutable.Some("domain_first_route"),
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
