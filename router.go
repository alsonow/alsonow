// Package alsonow
// Copyright 2025 alsonow. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.
package alsonow

import "net/http"

type HandlerFunc func(*Context)

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
}

type routerImpl struct {
	routes map[string]map[string][]HandlerFunc // method -> path -> handlers
}

func (r *routerImpl) addRoute(method, path string, handlers []HandlerFunc) {
	if _, ok := r.routes[method]; !ok {
		r.routes[method] = make(map[string][]HandlerFunc)
	}
	r.routes[method][path] = handlers
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

func (r *routerImpl) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	method := req.Method

	handlers, ok := r.routes[method][path]
	if !ok {
		http.NotFound(w, req)
		return
	}

	ctx := &Context{
		Writer:   w,
		Req:      req,
		Params:   make(map[string]string),
		Query:    make(map[string]string),
		handlers: handlers,
		index:    -1,
	}

	for k, v := range req.URL.Query() {
		if len(v) > 0 {
			ctx.Query[k] = v[0]
		}
	}

	ctx.Next()
}
