// Package alsonow
// Copyright 2025 alsonow. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.
package alsonow

import (
	"net/http"
	"strings"
	"sync"
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
	staticRoutes      map[string]map[string][]HandlerFunc // method -> path -> handlers
	paramRoutes       map[string]map[string][]HandlerFunc // method -> pattern -> handlers
	globalMiddlewares []HandlerFunc
	pool              sync.Pool
}

type Group struct {
	prefix      string
	middlewares []HandlerFunc
	router      *routerImpl
}

func newRouter() Router {
	r := &routerImpl{
		staticRoutes: make(map[string]map[string][]HandlerFunc),
		paramRoutes:  make(map[string]map[string][]HandlerFunc),
	}
	r.pool.New = func() any {
		return &Context{
			params: make(map[string]string),
			keys:   make(map[string]any),
		}
	}
	return r
}

func normalizePath(path string) string {
	if path == "" {
		return "/"
	}
	path = strings.TrimSpace(path)
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	path = strings.TrimSuffix(path, "/")
	if path == "" {
		return "/"
	}
	return path
}

func (r *routerImpl) addRoute(method, rawPath string, routeMiddlewares, handlers []HandlerFunc) {
	path := normalizePath(rawPath)

	finalHandlers := make([]HandlerFunc, 0, len(r.globalMiddlewares)+len(routeMiddlewares)+len(handlers))
	finalHandlers = append(finalHandlers, r.globalMiddlewares...)
	finalHandlers = append(finalHandlers, routeMiddlewares...)
	finalHandlers = append(finalHandlers, handlers...)

	if strings.ContainsAny(path[1:], ":") {
		if r.paramRoutes[method] == nil {
			r.paramRoutes[method] = make(map[string][]HandlerFunc)
		}
		r.paramRoutes[method][path] = finalHandlers
	} else {
		if r.staticRoutes[method] == nil {
			r.staticRoutes[method] = make(map[string][]HandlerFunc)
		}
		r.staticRoutes[method][path] = finalHandlers
	}
}

func (r *routerImpl) GET(path string, handlers ...HandlerFunc) {
	r.addRoute(http.MethodGet, path, nil, handlers)
}

func (r *routerImpl) POST(path string, handlers ...HandlerFunc) {
	r.addRoute(http.MethodPost, path, nil, handlers)
}

func (r *routerImpl) PUT(path string, handlers ...HandlerFunc) {
	r.addRoute(http.MethodPut, path, nil, handlers)
}

func (r *routerImpl) DELETE(path string, handlers ...HandlerFunc) {
	r.addRoute(http.MethodDelete, path, nil, handlers)
}

func (r *routerImpl) PATCH(path string, handlers ...HandlerFunc) {
	r.addRoute(http.MethodPatch, path, nil, handlers)
}

func (r *routerImpl) OPTIONS(path string, handlers ...HandlerFunc) {
	r.addRoute(http.MethodOptions, path, nil, handlers)
}

func (r *routerImpl) HEAD(path string, handlers ...HandlerFunc) {
	r.addRoute(http.MethodHead, path, nil, handlers)
}

func (r *routerImpl) CONNECT(path string, handlers ...HandlerFunc) {
	r.addRoute(http.MethodConnect, path, nil, handlers)
}

func (r *routerImpl) TRACE(path string, handlers ...HandlerFunc) {
	r.addRoute(http.MethodTrace, path, nil, handlers)
}

func (r *routerImpl) Use(middlewares ...HandlerFunc) {
	r.globalMiddlewares = append(r.globalMiddlewares, middlewares...)
}

func (r *routerImpl) Group(prefix string, middlewares ...HandlerFunc) *Group {
	return &Group{
		prefix:      normalizePath(prefix),
		middlewares: append([]HandlerFunc{}, append(r.globalMiddlewares, middlewares...)...),
		router:      r,
	}
}

func (g *Group) add(method, path string, handlers ...HandlerFunc) {
	fullPath := g.prefix
	if len(path) > 0 {
		if path[0] != '/' && !strings.HasSuffix(fullPath, "/") {
			fullPath += "/"
		}
		fullPath += strings.TrimPrefix(path, "/")
	}
	g.router.addRoute(method, fullPath, g.middlewares, handlers)
}

func (g *Group) GET(path string, handlers ...HandlerFunc) { g.add(http.MethodGet, path, handlers...) }

func (g *Group) POST(path string, handlers ...HandlerFunc) { g.add(http.MethodPost, path, handlers...) }

func (g *Group) PUT(path string, handlers ...HandlerFunc) { g.add(http.MethodPut, path, handlers...) }

func (g *Group) DELETE(path string, handlers ...HandlerFunc) {
	g.add(http.MethodDelete, path, handlers...)
}

func (g *Group) PATCH(path string, handlers ...HandlerFunc) {
	g.add(http.MethodPatch, path, handlers...)
}

func (g *Group) OPTIONS(path string, handlers ...HandlerFunc) {
	g.add(http.MethodOptions, path, handlers...)
}

func (g *Group) HEAD(path string, handlers ...HandlerFunc) { g.add(http.MethodHead, path, handlers...) }

func (g *Group) CONNECT(path string, handlers ...HandlerFunc) {
	g.add(http.MethodConnect, path, handlers...)
}

func (g *Group) TRACE(path string, handlers ...HandlerFunc) {
	g.add(http.MethodTrace, path, handlers...)
}

func (g *Group) Group(subPrefix string, middlewares ...HandlerFunc) *Group {
	newPrefix := g.prefix
	if !strings.HasSuffix(newPrefix, "/") {
		newPrefix += "/"
	}
	newPrefix += strings.TrimPrefix(normalizePath(subPrefix), "/")

	combined := make([]HandlerFunc, 0, len(g.middlewares)+len(middlewares))
	combined = append(combined, g.middlewares...)
	combined = append(combined, middlewares...)

	return &Group{
		prefix:      newPrefix,
		middlewares: combined,
		router:      g.router,
	}
}

func (r *routerImpl) acquireContext(w http.ResponseWriter, req *http.Request, handlers []HandlerFunc) *Context {
	ctx := r.pool.Get().(*Context)
	ctx.Writer = w
	ctx.Req = req
	ctx.handlers = handlers
	ctx.index = -1
	ctx.aborted = false

	for k := range ctx.params {
		delete(ctx.params, k)
	}
	for k := range ctx.keys {
		delete(ctx.keys, k)
	}

	return ctx
}

func (r *routerImpl) releaseContext(ctx *Context) {
	ctx.handlers = nil
	ctx.Writer = nil
	ctx.Req = nil
	r.pool.Put(ctx)
}

func matchPath(pattern, path string) (map[string]string, bool) {
	pattern = normalizePath(pattern)
	path = normalizePath(path)

	if pattern == path {
		return nil, true
	}

	pp := strings.Split(pattern[1:], "/")
	ap := strings.Split(path[1:], "/")

	if len(pp) != len(ap) {
		return nil, false
	}

	params := make(map[string]string)
	for i, part := range pp {
		if strings.HasPrefix(part, ":") {
			name := part[1:]
			if name == "" {
				return nil, false
			}
			params[name] = ap[i]
		} else if part != ap[i] {
			return nil, false
		}
	}
	return params, true
}

func (r *routerImpl) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	method := req.Method
	path := normalizePath(req.URL.Path)

	var handlers []HandlerFunc
	var params map[string]string

	if m, ok := r.staticRoutes[method]; ok {
		if h, ok := m[path]; ok {
			handlers = h
		}
	}

	if handlers == nil {
		if m, ok := r.paramRoutes[method]; ok {
			for pattern, h := range m {
				if p, ok := matchPath(pattern, path); ok {
					handlers = h
					params = p
					break
				}
			}
		}
	}

	if handlers == nil {
		http.NotFound(w, req)
		return
	}

	ctx := r.acquireContext(w, req, handlers)
	for k, v := range params {
		ctx.params[k] = v
	}

	ctx.Next()

	r.releaseContext(ctx)
}
