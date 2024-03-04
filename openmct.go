//go:build openmct
package gotelem

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// this package provides a web router for the statif openmct build.
// it should only be included if the build has been run,
// to do so, run npm install and then npm run build.

//go:embed web/dist
var public embed.FS

func OpenMCTRouter(r chi.Router) {
	// strip the subdirectory
	pfs, _ := fs.Sub(public, "web/dist")

	// default route.
	r.Handle("/*", http.FileServerFS(pfs))
}

func init() {
	RouterMods = append(RouterMods, OpenMCTRouter)
}
