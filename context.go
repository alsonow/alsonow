// Package alsonow
// Copyright 2025 alsonow. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.
package alsonow

import (
	"context"
	"net/http"
	"net/url"
)

// HandlerFunc defines the handler used by router.
type HandlerFunc func(*Context)

// Context holds request-scope data and provides helper methods.
type Context struct {
	Writer   http.ResponseWriter
	Req      *http.Request
	params   map[string]string
	data     map[string]any // Stores custom data for the request.
	index    int
	handlers []HandlerFunc
	aborted  bool
}

// Context returns the request's context
func (c *Context) Context() context.Context {
	return c.Req.Context()
}

// Header returns the value of a request header.
func (c *Context) Header(key string) string {
	return c.Req.Header.Get(key)
}

// SetHeader sets a response header.
func (c *Context) SetHeader(key, value string) {
	c.Writer.Header().Set(key, value)
}

// Status sets the HTTP status code (does not write headers yet).
func (c *Context) Status(code int) {
	c.Writer.WriteHeader(code)
}

// SetCookie sets a cookie in the response.
func (c *Context) SetCookie(cookie *http.Cookie) {
	c.Writer.Header().Add("Set-Cookie", cookie.String())
}

// Cookie gets the value of a named cookie from the request.
// Returns empty string and error if not found.
func (c *Context) Cookie(name string) (string, error) {
	cookie, err := c.Req.Cookie(name)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// DeleteCookie removes a cookie by setting it to expired.
func (c *Context) DeleteCookie(name string) {
	c.SetCookie(&http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// Host to get the host of the request
func (c *Context) Host() string {
	return c.Req.Host
}

// URL to get the full URL (scheme://host/path?query)
func (c *Context) URL() string {
	return c.Scheme() + "://" + c.Host() + c.Req.URL.Path
}

// Scheme to get the scheme of the request
func (c *Context) Scheme() string {
	if c.Req.TLS != nil {
		return "https"
	}
	return "http"
}

// Path to get the full normalized path of the request
func (c *Context) Path() string {
	return c.Req.URL.Path
}

// Method to get the HTTP method of the request
func (c *Context) Method() string {
	return c.Req.Method
}

// Param returns the value of a named route parameter.
func (c *Context) Param(key string) string {
	if c.params == nil {
		return ""
	}
	return c.params[key]
}

// Params returns the Context params.
func (c *Context) Params() map[string]string {
	return c.params
}

// QueryParam returns the first value for the named query parameter.
func (c *Context) QueryParam(key string) string {
	return c.Req.URL.Query().Get(key)
}

// QueryAll returns the full parsed query values.
func (c *Context) QueryAll() url.Values {
	return c.Req.URL.Query()
}

// Set stores a value in the request context.
func (c *Context) Set(key string, value any) {
	if c.data == nil {
		c.data = make(map[string]any)
	}
	c.data[key] = value
}

// Get retrieves a value from the request context.
func (c *Context) Get(key string) (any, bool) {
	v, ok := c.data[key]
	return v, ok
}

// Next invokes the next handler in the chain.
func (c *Context) Next() {
	// If already aborted or request context is done, stop processing
	if c.aborted {
		return
	}

	select {
	case <-c.Req.Context().Done():
		return
	default:
	}

	c.index++

	for c.index < len(c.handlers) {
		if c.aborted {
			return
		}

		c.handlers[c.index](c)
		c.index++
	}

	// Concise, elegant...
	// If you want to a take another look during the next code review~ (￣▽￣*)ゞ
	// for ; c.index < len(c.handlers); c.index++ {
	// 	if c.aborted {
	// 		return
	// 	}
	// 	c.handlers[c.index](c)
	// }
}

// Abort stops execution of remaining handlers.
func (c *Context) Abort() {
	c.aborted = true
}

// IsAborted reports whether the handler chain has been aborted.
func (c *Context) IsAborted() bool {
	return c.aborted
}
