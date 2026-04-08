package boltrouter

import (
	"fmt"
	"log"
	"maps"
	"net/http"
	"strings"

	"github.com/jaredtmartin/bolt-go"
)

type Layout func(http.ResponseWriter, *http.Request, ...bolt.Element) bolt.Element
type Handler func(http.ResponseWriter, *http.Request) ResponseType
type PathType map[string]Handler
type BranchType map[string]*PathType
type ErrorPageType func(err ResponseType) bolt.Element
type ResponseType interface {
	Err() error
	ErrPublic() string
	ErrDetail() string
	GetContent() []bolt.Element
}
type ResponseStruct struct {
	content []bolt.Element
	headers map[string]string
	err     error
}

func (r *ResponseStruct) Error(err error) *ResponseStruct {
	r.err = err
	return r
}
func (r *ResponseStruct) Content(content ...bolt.Element) *ResponseStruct {
	r.content = content
	return r
}
func (r *ResponseStruct) Success(msg string) *ResponseStruct {
	r.headers["HX-Success"] = msg
	return r
}
func (r *ResponseStruct) Header(key, value string) *ResponseStruct {
	r.headers[key] = value
	return r
}
func (r *ResponseStruct) Warning(msg string) *ResponseStruct {
	r.headers["HX-Warning"] = msg
	return r
}
func (r *ResponseStruct) Info(msg string) *ResponseStruct {
	r.headers["HX-Info"] = msg
	return r
}
func (r *ResponseStruct) Redirect(msg string) *ResponseStruct {
	r.headers["HX-Redirect"] = msg
	return r
}
func (r *ResponseStruct) PushUrl(url string) *ResponseStruct {
	r.headers["HX-Push-Url"] = url
	return r
}
func (r *ResponseStruct) ReplaceUrl(url string) *ResponseStruct {
	r.headers["HX-Replace-Url"] = url
	return r
}
func (r *ResponseStruct) Err() error {
	return r.err
}
func (r *ResponseStruct) ErrPublic() string {
	if r.err == nil {
		return ""
	}
	parts := strings.Split(r.err.Error(), ":")
	if len(parts) < 1 {
		return ""
	}
	return strings.TrimSpace(parts[0])
}
func (r *ResponseStruct) ErrDetail() string {
	if r.err == nil {
		return ""
	}
	parts := strings.Split(r.err.Error(), ":")
	if len(parts) < 2 {
		return ""
	}
	// join all the parts after the first one with :
	return strings.TrimSpace(strings.Join(parts[1:], ": "))
}
func (r *ResponseStruct) GetContent() []bolt.Element {
	return r.content
}

// Error(err).WrapErr("This dog has wandered off.")
func (r *ResponseStruct) WrapErr(msg string) *ResponseStruct {
	r.err = fmt.Errorf("%s: %w", msg, r.err)
	return r
}

func Response() *ResponseStruct {
	return &ResponseStruct{
		headers: make(map[string]string),
	}
}
func Content(content ...bolt.Element) ResponseType {
	return Response().Content(content...)
}
func Error(err error) ResponseType {
	res := Response().Error(err)
	return res
}
func Success(msg string) ResponseType {
	return Response().Success(msg)
}

type Router struct {
	layout    Layout
	routes    BranchType
	errorPage ErrorPageType
	Mux       *http.ServeMux
	verbose   bool
}

func Branch() BranchType {
	return make(BranchType)
}
func NewRouter(mux *http.ServeMux, layout Layout, errorPage ErrorPageType) *Router {
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
		if response.Err() != nil {
			if r.Header.Get("HX-Request") != "" {
				http.Error(w, response.ErrPublic(), http.StatusInternalServerError)
				return
			}
			router.layout(w, r, router.errorPage(response)).Send(w)
			return
		}
		router.layout(w, r, bolt.Fragment(response.GetContent()...)).Send(w)
		return
	}
	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}
