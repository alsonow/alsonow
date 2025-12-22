// Package alsonow
// Copyright 2025 alsonow. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.
package alsonow

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type HandlerFunc func(*Context)

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
	c.Writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	c.Writer.WriteHeader(status)
	fmt.Fprintf(c.Writer, format, args...)
}

func (c *Context) JSON(status int, v any) {
	c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	c.Writer.WriteHeader(status)
	if err := json.NewEncoder(c.Writer).Encode(v); err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
	}
}
