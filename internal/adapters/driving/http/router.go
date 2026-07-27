package http

import (
	"net/http"
)

type Route struct {
	Method     string
	Path       string
	Handler    http.HandlerFunc
	Middleware []string
	Name       string
}

type Router struct {
	routes []Route
	groups []RouteGroup
}

type RouteGroup struct {
	Prefix     string
	Middleware []string
	Routes     []Route
}

func NewRouter() *Router {
	return &Router{}
}

func (r *Router) GET(path string, handler http.HandlerFunc, middleware ...string) *Route {
	route := Route{Method: http.MethodGet, Path: path, Handler: handler, Middleware: middleware}
	r.routes = append(r.routes, route)
	return &r.routes[len(r.routes)-1]
}

func (r *Router) POST(path string, handler http.HandlerFunc, middleware ...string) *Route {
	route := Route{Method: http.MethodPost, Path: path, Handler: handler, Middleware: middleware}
	r.routes = append(r.routes, route)
	return &r.routes[len(r.routes)-1]
}

func (r *Router) PUT(path string, handler http.HandlerFunc, middleware ...string) *Route {
	route := Route{Method: http.MethodPut, Path: path, Handler: handler, Middleware: middleware}
	r.routes = append(r.routes, route)
	return &r.routes[len(r.routes)-1]
}

func (r *Router) DELETE(path string, handler http.HandlerFunc, middleware ...string) *Route {
	route := Route{Method: http.MethodDelete, Path: path, Handler: handler, Middleware: middleware}
	r.routes = append(r.routes, route)
	return &r.routes[len(r.routes)-1]
}

func (r *Router) PATCH(path string, handler http.HandlerFunc, middleware ...string) *Route {
	route := Route{Method: http.MethodPatch, Path: path, Handler: handler, Middleware: middleware}
	r.routes = append(r.routes, route)
	return &r.routes[len(r.routes)-1]
}

func (r *Router) Group(prefix string, middleware []string, fn func(*Router)) {
	sub := NewRouter()
	fn(sub)
	for _, route := range sub.routes {
		route.Path = prefix + route.Path
		route.Middleware = append(middleware, route.Middleware...)
		r.routes = append(r.routes, route)
	}
}

func (r *Router) All() []Route {
	return r.routes
}
