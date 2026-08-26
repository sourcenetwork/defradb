//go:build rust_ffi

package rustffi

/*
#cgo CFLAGS: -I${SRCDIR}
#cgo LDFLAGS: -L${SRCDIR} -ldefra_ffi -Wl,-rpath,${SRCDIR}

#include "defra.h"
#include <stdlib.h>
*/
import "C"

import (
	"encoding/hex"
	"fmt"

	"github.com/sourcenetwork/defradb/acp/identity"
)

// RegisterIdentityWithRust registers an identity's signing config with Rust's global identity store.
// This allows Go-created identities to be used for block signing in Rust.
func RegisterIdentityWithRust(ident identity.Identity) error {
	if ident == nil {
		return nil
	}

	fullIdent, ok := ident.(identity.FullIdentity)
	if !ok {
		return nil
	}

	did := ident.DID()
	privateKeyHex := hex.EncodeToString(fullIdent.PrivateKey().Raw())
	publicKeyHex := hex.EncodeToString(ident.PublicKey().Raw())
	keyType := string(fullIdent.PrivateKey().Type())

	var a cargs
	defer a.free()

	result := C.RegisterIdentity(a.s(did), a.s(privateKeyHex), a.s(publicKeyHex), a.s(keyType))

	if result.status != 0 {
		err := C.GoString(result.error)
		C.defra_free_string(result.error)
		return fmt.Errorf("ffi: RegisterIdentity failed: %s", err)
	}

	if result.value != nil {
		C.defra_free_string(result.value)
	}

	return nil
}
