//go:build !js || !wasm

package vgrouter

import (
	"fmt"
	"log"
	"strings"

	"github.com/vugu/vugu/js"
)

// Router handles URL routing.
type Router struct {
	useFragment bool
	pathPrefix  string

	popStateFunc js.Func

	eventEnv EventEnv

	rlist           []routeEntry
	notFoundHandler RouteHandler

	// bindRoutePath string // the route (with :param stuff in it) that matches the bind params, so we can reconstruct it
	bindRouteMPath mpath
	bindParamMap   map[string]BindParam
}

// ListenForPopState registers an event listener so the user navigating with
// forward/back/history or fragment changes will be detected and handled by this router.
// Any call to SetUseFragment or SetPathPrefix should occur before calling
// ListenForPopState.
//
// Only works in wasm environment and if called outside it will have no effect and return error.
func (r *Router) ListenForPopState() error {
	return r.addPopStateListener(func(this js.Value, args []js.Value) interface{} {
		// TODO: see if we need something better for error handling

		// log.Printf("addPopStateListener callack")

		u, err := r.readBrowserURL()
		// log.Printf("addPopStateListener callack: u=%#v, err=%v", u, err)
		if err != nil {
			log.Printf("ListenForPopState: error from readBrowserURL: %v", err)
			return nil
		}

		p := u.Path
		if !strings.HasPrefix(p, r.pathPrefix) {
			log.Printf("ListenForPopState: prefix error: %v",
				ErrMissingPrefix{Path: p, Message: fmt.Sprintf("path %q does not begin with prefix %q", p, r.pathPrefix)})
			return nil
		}

		tp := strings.TrimPrefix(p, r.pathPrefix)
		q := u.Query()

		// log.Printf("addPopStateListener calling process: tp=%q, q=%#v", tp, q)

		r.eventEnv.Lock()
		defer r.eventEnv.UnlockRender()
		r.process(tp, q)

		return nil
	})
}

// BrowserAvail returns true if in browser mode.
func (r *Router) BrowserAvail() bool {
	// this is really just so otehr packages don't have to import `js` just to figure out if they should do extra browser setup
	return js.Global().Truthy()
}
