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

	Group(prefix string, middlewares ...HandlerFunc) *Group
	Use(middlewares ...HandlerFunc)
}

// node represents a radix tree node.
// https://en.wikipedia.org/wiki/Radix_tree
type node struct {
	children   map[string]*node
	paramChild *node
	handlers   []HandlerFunc
	isEnd      bool
	paramName  string
}

// routerImpl router implementation
type routerImpl struct {
	trees             map[string]*node // method -> root node
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
		trees: make(map[string]*node),
	}
	r.pool.New = func() any {
		return &Context{
			params: make(map[string]string, 4),
			data:   make(map[string]any, 10),
		}
	}
	return r
}

func normalizePath(path string) string {
	if path == "" {
		return "/"
	}
	path = strings.TrimSpace(path)
	if path != "/" {
		path = strings.TrimSuffix(path, "/")
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	for strings.Contains(path, "//") {
		path = strings.ReplaceAll(path, "//", "/")
	}
	return path
}

func (r *routerImpl) getTree(method string) *node {
	if r.trees[method] == nil {
		r.trees[method] = &node{
			children: make(map[string]*node),
		}
	}
	return r.trees[method]
}

func (r *routerImpl) insert(method, path string, handlers []HandlerFunc) {
	path = normalizePath(path)
	root := r.getTree(method)

	if path == "/" {
		root.isEnd = true
		root.handlers = handlers
		return
	}

	segments := strings.Split(path[1:], "/")
	cur := root

	for i, segment := range segments {
		if segment == "" {
			continue
		}

		isParam := segment[0] == ':'
		var child *node

		if isParam {
			if cur.paramChild == nil {
				cur.paramChild = &node{
					children:  make(map[string]*node),
					paramName: segment[1:],
				}
			}
			child = cur.paramChild
		} else {
			if cur.children == nil {
				cur.children = make(map[string]*node)
			}
			if _, ok := cur.children[segment]; !ok {
				cur.children[segment] = &node{
					children: make(map[string]*node),
				}
			}
			child = cur.children[segment]
		}

		cur = child

		if i == len(segments)-1 {
			cur.isEnd = true
			cur.handlers = handlers
		}
	}
}

func (r *routerImpl) search(method, path string) ([]HandlerFunc, map[string]string) {
	path = normalizePath(path)
	root := r.trees[method]
	if root == nil {
		return nil, nil
	}

	if path == "/" {
		if root.isEnd {
			return root.handlers, nil
		}
		return nil, nil
	}

	segments := strings.Split(path[1:], "/")
	params := make(map[string]string)
	cur := root

	for i, segment := range segments {
		if segment == "" {
			continue
		}

		found := false

		if cur.children != nil {
			if child, ok := cur.children[segment]; ok {
				cur = child
				found = true
			}
		}

		if !found && cur.paramChild != nil {
			cur = cur.paramChild
			if cur.paramName != "" {
				params[cur.paramName] = segment
			}
			found = true
		}

		if !found {
			return nil, nil
		}

		if i == len(segments)-1 && cur.isEnd {
			return cur.handlers, params
		}
	}

	return nil, nil
}

func (r *routerImpl) addRoute(method, path string, groupMiddlewares, handlers []HandlerFunc) {
	final := make([]HandlerFunc, 0, len(r.globalMiddlewares)+len(groupMiddlewares)+len(handlers))
	final = append(final, r.globalMiddlewares...)
	final = append(final, groupMiddlewares...)
	final = append(final, handlers...)

	r.insert(method, path, final)
}

func (r *routerImpl) GET(path string, h ...HandlerFunc)  { r.addRoute(http.MethodGet, path, nil, h) }
func (r *routerImpl) POST(path string, h ...HandlerFunc) { r.addRoute(http.MethodPost, path, nil, h) }
func (r *routerImpl) PUT(path string, h ...HandlerFunc)  { r.addRoute(http.MethodPut, path, nil, h) }
func (r *routerImpl) DELETE(path string, h ...HandlerFunc) {
	r.addRoute(http.MethodDelete, path, nil, h)
}
func (r *routerImpl) PATCH(path string, h ...HandlerFunc) { r.addRoute(http.MethodPatch, path, nil, h) }
func (r *routerImpl) OPTIONS(path string, h ...HandlerFunc) {
	r.addRoute(http.MethodOptions, path, nil, h)
}
func (r *routerImpl) HEAD(path string, h ...HandlerFunc) { r.addRoute(http.MethodHead, path, nil, h) }

func (r *routerImpl) Use(m ...HandlerFunc) {
	r.globalMiddlewares = append(r.globalMiddlewares, m...)
}

func (r *routerImpl) Group(prefix string, m ...HandlerFunc) *Group {
	combined := make([]HandlerFunc, 0, len(r.globalMiddlewares)+len(m))
	combined = append(combined, r.globalMiddlewares...)
	combined = append(combined, m...)

	return &Group{
		prefix:      normalizePath(prefix),
		middlewares: combined,
		router:      r,
	}
}

func (g *Group) add(method, path string, h ...HandlerFunc) {
	fullPath := g.prefix
	if path = normalizePath(path); path != "/" {
		if !strings.HasSuffix(fullPath, "/") {
			fullPath += "/"
		}
		fullPath += strings.TrimPrefix(path, "/")
	}
	g.router.addRoute(method, fullPath, g.middlewares, h)
}

func (g *Group) GET(path string, h ...HandlerFunc)     { g.add(http.MethodGet, path, h...) }
func (g *Group) POST(path string, h ...HandlerFunc)    { g.add(http.MethodPost, path, h...) }
func (g *Group) PUT(path string, h ...HandlerFunc)     { g.add(http.MethodPut, path, h...) }
func (g *Group) DELETE(path string, h ...HandlerFunc)  { g.add(http.MethodDelete, path, h...) }
func (g *Group) PATCH(path string, h ...HandlerFunc)   { g.add(http.MethodPatch, path, h...) }
func (g *Group) OPTIONS(path string, h ...HandlerFunc) { g.add(http.MethodOptions, path, h...) }
func (g *Group) HEAD(path string, h ...HandlerFunc)    { g.add(http.MethodHead, path, h...) }

func (g *Group) Group(sub string, m ...HandlerFunc) *Group {
	newPrefix := g.prefix
	if !strings.HasSuffix(newPrefix, "/") {
		newPrefix += "/"
	}
	newPrefix += strings.TrimPrefix(normalizePath(sub), "/")

	combined := make([]HandlerFunc, 0, len(g.middlewares)+len(m))
	combined = append(combined, g.middlewares...)
	combined = append(combined, m...)

	return &Group{
		prefix:      newPrefix,
		middlewares: combined,
		router:      g.router,
	}
}

func (r *routerImpl) acquireCtx(w http.ResponseWriter, req *http.Request, h []HandlerFunc) *Context {
	ctx := r.pool.Get().(*Context)
	ctx.Writer = w
	ctx.Req = req
	ctx.handlers = h
	ctx.index = -1
	ctx.aborted = false

	for k := range ctx.params {
		delete(ctx.params, k)
	}
	for k := range ctx.data {
		delete(ctx.data, k)
	}
	return ctx
}

func (r *routerImpl) releaseCtx(ctx *Context) {
	ctx.handlers = nil
	ctx.Writer = nil
	ctx.Req = nil
	r.pool.Put(ctx)
}

func (r *routerImpl) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	handlers, params := r.search(req.Method, req.URL.Path)
	if handlers == nil {
		http.NotFound(w, req)
		return
	}

	ctx := r.acquireCtx(w, req, handlers)
	for k, v := range params {
		ctx.params[k] = v
	}

	ctx.Next()
	r.releaseCtx(ctx)
}
