package delete

import (
	"fmt"
	"strings"
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/defradb/internal/keys"
	testAction "github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

type storageCheckAction struct {
	s          *state.State
	AssertKeys func([]string) error
}

func (a *storageCheckAction) SetState(s *state.State) {
	a.s = s
}
func (a *storageCheckAction) Execute() {
	nodeState := a.s.Nodes[0]
	c := nodeState.Client

	if storeClient, ok := c.(interface{ Rootstore() corekv.TxnStore }); ok {
		store := storeClient.Rootstore()
		var storedKeys []string
		iter, err := store.Iterator(a.s.Ctx, corekv.IterOptions{})
		if err != nil {
			a.s.T.Fatal(err)
			return
		}
		defer iter.Close()

		for {
			hasNext, err := iter.Next()
			if err != nil {
				a.s.T.Fatal(err)
				return
			}
			if !hasNext {
				break
			}
			storedKeys = append(storedKeys, string(iter.Key()))
		}

		if a.AssertKeys != nil {
			if err := a.AssertKeys(storedKeys); err != nil {
				a.s.T.Fatal(err)
			}
		}
	} else {
		a.s.T.Fatal("client does not expose Rootstore, cannot verify storage")
	}
}

func assertNoDataKeys() *storageCheckAction {
	return &storageCheckAction{
		AssertKeys: func(storedKeys []string) error {
			for _, k := range storedKeys {
				// Only check Data keys (prefixed with 'd')
				if !strings.HasPrefix(k, "d") {
					continue
				}

				// Strip 'd' prefix for decoding
				dsKey, err := keys.DecodeDataStoreKey([]byte(k[1:]))
				if err != nil {
					return fmt.Errorf("storage leak: failed to decode data key %s: %w", k, err)
				}

				// Allow Tombstones (DeletedKey) and Priority keys
				if dsKey.InstanceType == keys.DeletedKey || dsKey.InstanceType == keys.PriorityKey {
					continue
				}

				// Any other active key (Value, Field, Primary) is a leak
				return fmt.Errorf("storage leak: found active key %s", k)
			}
			return nil
		},
	}
}

func TestStorageFreedUponDelete(t *testing.T) {
	schema := `
		type User {
			name: String
		}
	`

	test := testUtils.TestCase{
		SupportedClientTypes:   immutable.Some([]state.ClientType{state.GoClientType}),
		SupportedDatabaseTypes: immutable.Some([]state.DatabaseType{testUtils.BadgerIMType, testUtils.BadgerFileType, testUtils.DefraIMType}),
		Actions: []any{
			&testAction.AddSchema{
				Schema: schema,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "Alice"}`,
			},
			testUtils.DeleteDoc{
				CollectionID: 0,
				DocID:        0,
			},
			assertNoDataKeys(),
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
