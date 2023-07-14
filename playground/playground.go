//go:build playground

package playground

import (
	"embed"
)

//go:generate npm install
//go:generate npm run build
//go:embed dist
var Dist embed.FS
