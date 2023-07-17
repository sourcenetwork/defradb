//go:build playground

package playground

import (
	"embed"
)

//go:embed dist
var Dist embed.FS
