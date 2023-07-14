//go:build playground

package http

import (
	"io/fs"
	"net/http"

	"github.com/sourcenetwork/defradb/playground"
)

func init() {
	sub, err := fs.Sub(playground.Dist, "dist")
	if err != nil {
		panic(err)
	}
	router.Handle("/*", http.FileServer(http.FS(sub)))
}
