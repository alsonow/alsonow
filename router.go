// Package alsonow
// Copyright 2025 alsonow. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.
package alsonow

import (
	"fmt"
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
	// trees method -> root node
	trees       map[string]*node
	middlewares []HandlerFunc
	pool        sync.Pool
}

type Group struct {
	prefix      string
	middlewares []HandlerFunc
	parent      *Group
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

	// Run the loop first.
	for strings.Contains(path, "//") {
		path = strings.ReplaceAll(path, "//", "/")
	}

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	if path != "/" {
		path = strings.TrimSuffix(path, "/")
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

func (r *routerImpl) insert(method, path string, combined []HandlerFunc) {
	path = normalizePath(path)
	root := r.getTree(method)

	if path == "/" {
		root.isEnd = true
		root.handlers = combined
		return
	}

	fmt.Println(path)
	segments := strings.Split(path[1:], "/")
	cur := root

	for _, segment := range segments {
		isParam := segment[0] == ':'
		var child *node

		if isParam {
			paramName := segment[1:]
			if cur.paramChild != nil {
				if cur.paramChild.paramName != paramName {
					panic(fmt.Sprintf(
						"cannot register '%s': parameter name ':%s' conflicts with existing ':%s' in previously registered path",
						path, paramName, cur.paramChild.paramName,
					))
				}
			} else {
				cur.paramChild = &node{
					paramName: paramName,
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
	}

	// At this point, len(segments) must be greater than 0
	cur.isEnd = true
	cur.handlers = combined
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

	for _, segment := range segments {
		if cur.children != nil {
			if child, ok := cur.children[segment]; ok {
				cur = child
				continue
			}
		}

		if cur.paramChild != nil {
			cur = cur.paramChild
			params[cur.paramName] = segment
			continue
		}

		return nil, nil
	}

	if cur.isEnd {
		return cur.handlers, params
	}

	return nil, nil
}

func (r *routerImpl) addRoute(method, path string, middlewares, handlers []HandlerFunc) {
	combined := make([]HandlerFunc, 0, len(middlewares)+len(handlers))
	combined = append(combined, middlewares...)
	combined = append(combined, handlers...)

	r.insert(method, path, combined)
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
	r.middlewares = append(r.middlewares, m...)
}

func (r *routerImpl) Group(prefix string, m ...HandlerFunc) *Group {
	return &Group{
		prefix:      normalizePath(prefix),
		middlewares: m,
		router:      r,
	}
}

func (g *Group) collectMiddlewares() []HandlerFunc {
	var mids []HandlerFunc
	current := g
	for current != nil {
		mids = append(mids, current.middlewares...)
		current = current.parent
	}

	mids = append(mids, g.router.middlewares...)
	return mids
}

func (g *Group) add(method, path string, h ...HandlerFunc) {
	fullPath := g.prefix
	if path = normalizePath(path); path != "/" {
		if !strings.HasSuffix(fullPath, "/") {
			fullPath += "/"
		}
		fullPath += strings.TrimPrefix(path, "/")
	}

	middlewares := g.collectMiddlewares()
	g.router.addRoute(method, fullPath, middlewares, h)
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

	return &Group{
		prefix:      newPrefix,
		middlewares: m,
		parent:      g,
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

	// go1.21+
	clear(ctx.params)
	clear(ctx.data)

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
