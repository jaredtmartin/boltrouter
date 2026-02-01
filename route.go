package route

import (
	"maps"
	"net/http"
)

type Handler func(http.ResponseWriter, *http.Request)
type RouteType map[string]Handler

// type Router map[string]Route

var routes = make(map[string]*RouteType)

func (r *RouteType) Handle(method string, handler Handler) *RouteType {
	(*r)[method] = handler
	return r
}

func (r *RouteType) Map(route RouteType) *RouteType {
	maps.Copy(*r, route)
	return r
}

func (r *RouteType) Get(handler Handler) *RouteType {
	(*r)[http.MethodGet] = handler
	return r
}
func (r *RouteType) Post(handler Handler) *RouteType {
	(*r)[http.MethodPost] = handler
	return r
}
func (r *RouteType) Delete(handler Handler) *RouteType {
	(*r)[http.MethodDelete] = handler
	return r
}
func (r *RouteType) Put(handler Handler) *RouteType {
	(*r)[http.MethodPut] = handler
	return r
}
func (r *RouteType) Patch(handler Handler) *RouteType {
	(*r)[http.MethodPatch] = handler
	return r
}

func Path(mux *http.ServeMux, path string) *RouteType {
	if routes[path] == nil {
		routes[path] = &RouteType{}
	}
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		pathHandler(w, r, *routes[path])
	})
	return routes[path]
}

func pathHandler(w http.ResponseWriter, r *http.Request, methods RouteType) {
	if handler, ok := methods[r.Method]; ok && handler != nil {
		handler(w, r)
		return
	}
	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}

// func Get(mux *http.ServeMux, path string, handler Handler) {
// 	Route(mux, path).Handle("GET", handler)
// }
