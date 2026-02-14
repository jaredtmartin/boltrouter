package boltrouter

import (
	"log"
	"maps"
	"net/http"

	"github.com/jaredtmartin/bolt-go"
)

type Layout func(http.ResponseWriter, *http.Request, ...bolt.Element) bolt.Element
type Handler func(http.ResponseWriter, *http.Request) (bolt.Element, error)
type PathType map[string]Handler
type ErrorPage func(error) bolt.Element

type Router struct {
	layout    Layout
	routes    map[string]*PathType
	errorPage ErrorPage
	mux       *http.ServeMux
	verbose   bool
}

func NewRouter(mux *http.ServeMux, layout Layout, errorPage ErrorPage) *Router {
	return &Router{
		layout:    layout,
		routes:    make(map[string]*PathType),
		errorPage: errorPage,
		mux:       mux,
	}
}

func (router *Router) Route(routes func(r *Router)) *Router {
	routes(router)
	return router
}
func (r *Router) Handle(w http.ResponseWriter, r2 *http.Request) {
	http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("public"))))
}
func (r *Router) Verbose(verbose bool) {
	r.verbose = verbose
}
func (router *Router) Path(path string) *PathType {
	if router.routes[path] == nil {
		router.routes[path] = &PathType{}
	}
	router.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		pathHandler(w, r, router, router.routes[path])
	})
	if router.verbose {
		log.Println("Added route:", path)
	}
	return router.routes[path]
}

func (r *PathType) Handle(method string, handler Handler) *PathType {
	(*r)[method] = handler
	return r
}

func (r *PathType) Map(route PathType) *PathType {
	maps.Copy(*r, route)
	return r
}

func (r *PathType) Get(handler Handler) *PathType {
	(*r)[http.MethodGet] = handler
	return r
}
func (r *PathType) Post(handler Handler) *PathType {
	(*r)[http.MethodPost] = handler
	return r
}
func (r *PathType) Delete(handler Handler) *PathType {
	(*r)[http.MethodDelete] = handler
	return r
}
func (r *PathType) Put(handler Handler) *PathType {
	(*r)[http.MethodPut] = handler
	return r
}
func (r *PathType) Patch(handler Handler) *PathType {
	(*r)[http.MethodPatch] = handler
	return r
}

func pathHandler(w http.ResponseWriter, r *http.Request, router *Router, methods *PathType) {
	if handler, ok := (*methods)[r.Method]; ok && handler != nil {
		element, err := handler(w, r)
		if err != nil {
			if r.Header.Get("HX-Request") != "" {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			router.layout(w, r, router.errorPage(err)).Send(w)
			return
		}
		router.layout(w, r, element).Send(w)
		return
	}
	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}
