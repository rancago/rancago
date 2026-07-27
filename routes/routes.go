// Package routes provides a declarative route builder for Rancago Framework.
// Supports method chaining and middleware string names for clean route registration.
package routes

import (
	"net/http"
)

// Route represents a single HTTP route.
type Route struct {
	Method     string
	Path       string
	Handler    http.HandlerFunc
	Middleware []string
	Name       string
}

// Router holds a collection of routes and route groups.
type Router struct {
	routes []Route
}

// NewRouter creates a new empty Router.
func NewRouter() *Router { return &Router{} }

// GET registers a GET route.
func (r *Router) GET(path string, handler http.HandlerFunc, middleware ...string) *Route {
	return r.add(http.MethodGet, path, handler, middleware)
}

// POST registers a POST route.
func (r *Router) POST(path string, handler http.HandlerFunc, middleware ...string) *Route {
	return r.add(http.MethodPost, path, handler, middleware)
}

// PUT registers a PUT route.
func (r *Router) PUT(path string, handler http.HandlerFunc, middleware ...string) *Route {
	return r.add(http.MethodPut, path, handler, middleware)
}

// PATCH registers a PATCH route.
func (r *Router) PATCH(path string, handler http.HandlerFunc, middleware ...string) *Route {
	return r.add(http.MethodPatch, path, handler, middleware)
}

// DELETE registers a DELETE route.
func (r *Router) DELETE(path string, handler http.HandlerFunc, middleware ...string) *Route {
	return r.add(http.MethodDelete, path, handler, middleware)
}

// Group adds a prefixed group of routes with shared middleware.
func (r *Router) Group(prefix string, middleware []string, fn func(*Router)) {
	sub := NewRouter()
	fn(sub)
	for _, route := range sub.routes {
		route.Path = prefix + route.Path
		route.Middleware = append(middleware, route.Middleware...)
		r.routes = append(r.routes, route)
	}
}

// As sets the name of the most recently registered route.
func (ro *Route) As(name string) *Route {
	ro.Name = name
	return ro
}

// All returns all registered routes.
func (r *Router) All() []Route { return r.routes }

// Mount registers all routes on an http.ServeMux.
// Middleware string names are informational only - wire up actual middleware in bootstrap.
func (r *Router) Mount(mux *http.ServeMux) {
	for _, route := range r.routes {
		handler := route.Handler
		mux.HandleFunc(route.Path, handler)
	}
}

func (r *Router) add(method, path string, handler http.HandlerFunc, middleware []string) *Route {
	route := Route{Method: method, Path: path, Handler: handler, Middleware: middleware}
	r.routes = append(r.routes, route)
	return &r.routes[len(r.routes)-1]
}

// Web returns a pre-configured Router for web (HTML) routes.
func Web() *Router { return NewRouter() }

// Api returns a pre-configured Router for API routes.
// All API routes are typically prefixed with /api/vN.
func Api() *Router { return NewRouter() }
