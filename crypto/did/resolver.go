package did

import (
	diddoc "github.com/hyperledger/aries-framework-go/pkg/doc/did"
	vdrapi "github.com/hyperledger/aries-framework-go/pkg/framework/aries/api/vdr"
	"github.com/hyperledger/aries-framework-go/pkg/vdr"
	"github.com/hyperledger/aries-framework-go/pkg/vdr/key"
)

var (
	// DefaultRegistry is the basic instance of the DID VDR Registry, which
	// supports resolving the following DID methods:
	// - did:key
	DefaultRegistry = NewRegistry(vdr.WithVDR(key.New()))
)

// Resolver is a interface to scope the VDR interface
// only to the Resolve method, which resolves
// a did document
type Resolver interface {
	Resolve(did string, opts ...vdrapi.DIDMethodOption) (*diddoc.DocResolution, error)
}

// Registry is a basic DID resolver
// @todo: Add a cache
type Registry struct {
	Resolver
}

// NewRegistry returns a new instance of the did.Registry which is a basic implementation
// of the did VDR.
func NewRegistry(opts ...vdr.Option) *Registry {
	baseVDR := vdr.New(opts...)
	return &Registry{Resolver: baseVDR}
}
