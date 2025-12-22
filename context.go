// Package alsonow
// Copyright 2025 alsonow. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.
package alsonow

import (
	"fmt"
	"net/http"
)

type Context struct {
	Writer   http.ResponseWriter
	Req      *http.Request
	Params   map[string]string
	Query    map[string]string
	handlers []HandlerFunc
	index    int
}

func (c *Context) Next() {
	c.index++
	if c.index >= len(c.handlers) {
		return
	}
	c.handlers[c.index](c)
}

func (c *Context) String(status int, format string, args ...any) {
	c.Writer.Header().Set("Content-Type", "text/plain")
	c.Writer.WriteHeader(status)
	fmt.Fprintf(c.Writer, format, args...)
}
