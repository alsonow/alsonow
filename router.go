// Package alsonow
// Copyright 2025 alsonow. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.
package alsonow

import (
	"net/http"
	"strings"
)

type Router interface {
	http.Handler
	GET(path string, handlers ...HandlerFunc)
	POST(path string, handlers ...HandlerFunc)
	PUT(path string, handlers ...HandlerFunc)
	DELETE(path string, handlers ...HandlerFunc)
	PATCH(path string, handlers ...HandlerFunc)
	OPTIONS(path string, handlers ...HandlerFunc)
	HEAD(path string, handlers ...HandlerFunc)
	CONNECT(path string, handlers ...HandlerFunc)
	TRACE(path string, handlers ...HandlerFunc)
	Group(prefix string, handlers ...HandlerFunc) *Group
	Use(middleware ...HandlerFunc)
}

type routerImpl struct {
	routes      map[string]map[string][]HandlerFunc // method -> path -> handlers
	middlewares []HandlerFunc
}

type Group struct {
	prefix string
	router *routerImpl
}

func newRouter() *routerImpl {
	return &routerImpl{
		routes: make(map[string]map[string][]HandlerFunc),
	}
}

func (r *routerImpl) addRoute(method, path string, handlers []HandlerFunc) {
	fullHandlers := append([]HandlerFunc{}, r.middlewares...)
	fullHandlers = append(fullHandlers, handlers...)

	if _, ok := r.routes[method]; !ok {
		r.routes[method] = make(map[string][]HandlerFunc)
	}
	r.routes[method][path] = fullHandlers
}

func (r *routerImpl) GET(path string, handlers ...HandlerFunc) {
	r.addRoute(http.MethodGet, path, handlers)
}

func (r *routerImpl) POST(path string, handlers ...HandlerFunc) {
	r.addRoute(http.MethodPost, path, handlers)
}

func (r *routerImpl) PUT(path string, handlers ...HandlerFunc) {
	r.addRoute(http.MethodPut, path, handlers)
}

func (r *routerImpl) DELETE(path string, handlers ...HandlerFunc) {
	r.addRoute(http.MethodDelete, path, handlers)
}

func (r *routerImpl) PATCH(path string, handlers ...HandlerFunc) {
	r.addRoute(http.MethodPatch, path, handlers)
}

func (r *routerImpl) OPTIONS(path string, handlers ...HandlerFunc) {
	r.addRoute(http.MethodOptions, path, handlers)
}

func (r *routerImpl) HEAD(path string, handlers ...HandlerFunc) {
	r.addRoute(http.MethodHead, path, handlers)
}

func (r *routerImpl) CONNECT(path string, handlers ...HandlerFunc) {
	r.addRoute(http.MethodConnect, path, handlers)
}

func (r *routerImpl) TRACE(path string, handlers ...HandlerFunc) {
	r.addRoute(http.MethodTrace, path, handlers)
}

func (r *routerImpl) Use(middleware ...HandlerFunc) {
	r.middlewares = append(r.middlewares, middleware...)
}

func (r *routerImpl) Group(prefix string, handlers ...HandlerFunc) *Group {
	g := &Group{
		prefix: strings.TrimSuffix(prefix, "/"),
		router: r,
	}
	g.router.middlewares = append(r.middlewares, handlers...)
	return g
}

func (g *Group) GET(path string, handlers ...HandlerFunc) {
	g.router.addRoute(http.MethodGet, g.prefix+path, handlers)
}

func (g *Group) POST(path string, handlers ...HandlerFunc) {
	g.router.addRoute(http.MethodPost, g.prefix+path, handlers)
}

func (g *Group) PUT(path string, handlers ...HandlerFunc) {
	g.router.addRoute(http.MethodPut, g.prefix+path, handlers)
}

func (g *Group) DELETE(path string, handlers ...HandlerFunc) {
	g.router.addRoute(http.MethodDelete, g.prefix+path, handlers)
}

func (g *Group) PATCH(path string, handlers ...HandlerFunc) {
	g.router.addRoute(http.MethodPatch, g.prefix+path, handlers)
}

func (g *Group) OPTIONS(path string, handlers ...HandlerFunc) {
	g.router.addRoute(http.MethodOptions, g.prefix+path, handlers)
}

func (g *Group) HEAD(path string, handlers ...HandlerFunc) {
	g.router.addRoute(http.MethodHead, g.prefix+path, handlers)
}

func (g *Group) CONNECT(path string, handlers ...HandlerFunc) {
	g.router.addRoute(http.MethodConnect, g.prefix+path, handlers)
}

func (g *Group) TRACE(path string, handlers ...HandlerFunc) {
	g.router.addRoute(http.MethodTrace, g.prefix+path, handlers)
}

func (g *Group) Group(prefix string, handlers ...HandlerFunc) *Group {
	return &Group{
		prefix: g.prefix + strings.TrimSuffix(prefix, "/"),
		router: g.router,
	}
}

func (r *routerImpl) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	method := req.Method

	if handlers, ok := r.routes[method][path]; ok {
		ctx := r.newContext(w, req, handlers)
		ctx.Next()
		return
	}

	// :id
	for p, handlers := range r.routes[method] {
		if matchParams, ok := matchPath(p, path); ok {
			ctx := r.newContext(w, req, handlers)
			ctx.Params = matchParams
			ctx.Next()
			return
		}
	}

	http.NotFound(w, req)
}

func matchPath(pattern, path string) (map[string]string, bool) {
	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")
	pathParts := strings.Split(strings.Trim(path, "/"), "/")

	if len(patternParts) != len(pathParts) {
		return nil, false
	}

	params := make(map[string]string)
	for i, part := range patternParts {
		if strings.HasPrefix(part, ":") || (strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}")) {
			name := strings.TrimPrefix(strings.TrimSuffix(part, "}"), ":")
			params[name] = pathParts[i]
		} else if part != pathParts[i] {
			return nil, false
		}
	}
	return params, true
}

func (r *routerImpl) newContext(w http.ResponseWriter, req *http.Request, handlers []HandlerFunc) *Context {
	query := make(map[string]string)
	for k, v := range req.URL.Query() {
		if len(v) > 0 {
			query[k] = v[0]
		}
	}
	return &Context{
		Writer:   w,
		Req:      req,
		Params:   make(map[string]string),
		Query:    query,
		handlers: handlers,
		index:    -1,
	}
}
