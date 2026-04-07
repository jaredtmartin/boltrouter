package boltrouter

import (
	"log"
	"maps"
	"net/http"

	"github.com/jaredtmartin/bolt-go"
)

type Layout func(http.ResponseWriter, *http.Request, ...bolt.Element) bolt.Element
type Handler func(http.ResponseWriter, *http.Request) *ResponseType
type PathType map[string]Handler
type BranchType map[string]*PathType
type ErrorPage func(err Error) bolt.Element
type ResponseType struct {
	content bolt.Element
	err     Error
}

func Response() *ResponseType {
	return &ResponseType{
		content: nil,
		err:     nil,
	}
}
func Content(content bolt.Element) *ResponseType {
	return &ResponseType{
		content: content,
		err:     nil,
	}
}
func (r *ResponseType) Content(content bolt.Element) *ResponseType {
	r.content = content
	return r
}
func (r *ResponseType) Error(message string, details ...string) *ResponseType {
	r.err = NewError(message, details...)
	return r
}

type Router struct {
	layout    Layout
	routes    BranchType
	errorPage ErrorPage
	Mux       *http.ServeMux
	verbose   bool
}

func Branch() BranchType {
	return make(BranchType)
}
func NewRouter(mux *http.ServeMux, layout Layout, errorPage ErrorPage) *Router {
	return &Router{
		layout:    layout,
		routes:    Branch(),
		errorPage: errorPage,
		Mux:       mux,
	}
}
func (router *Router) Route(routes BranchType) *Router {
	for path, route := range routes {
		router.Path(path).Map(*route)
	}
	return router
}
func (router *Router) Handle(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	router.Mux.HandleFunc(pattern, handler)
}
func (router *Router) Verbose(verbose bool) {
	router.verbose = verbose
}
func (router *Router) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	router.Mux.HandleFunc(pattern, handler)
}
func (router *Router) Path(path string) *PathType {
	if router.routes[path] == nil {
		router.routes[path] = &PathType{}
	}
	router.Mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		pathHandler(w, r, router, *router.routes[path])
	})
	if router.verbose {
		log.Println("Added route:", urlFromPath(path))
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

func pathHandler(w http.ResponseWriter, r *http.Request, router *Router, methods PathType) {
	if handler, ok := (methods)[r.Method]; ok && handler != nil {
		response := handler(w, r)
		if response.err != nil {
			if r.Header.Get("HX-Request") != "" {
				http.Error(w, response.err.Error(), http.StatusInternalServerError)
				return
			}
			router.layout(w, r, router.errorPage(response.err)).Send(w)
			return
		}
		router.layout(w, r, response.content).Send(w)
		return
	}
	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}
