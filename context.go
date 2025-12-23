// Package alsonow
// Copyright 2025 alsonow. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.
package alsonow

import (
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
)

type HandlerFunc func(*Context)

type Context struct {
	Writer   http.ResponseWriter
	Req      *http.Request
	params   map[string]string
	handlers []HandlerFunc
	index    int
	aborted  bool
	keys     map[string]any
}

func (c *Context) Abort() {
	c.aborted = true
}

func (c *Context) IsAborted() bool {
	return c.aborted
}

func (c *Context) Next() {
	if c.aborted {
		return
	}
	c.index++
	for c.index < len(c.handlers) {
		c.handlers[c.index](c)
		c.index++
		if c.aborted {
			return
		}
	}
}

func (c *Context) Param(key string) string {
	if c.params == nil {
		return ""
	}
	return c.params[key]
}

func (c *Context) SetParam(key, value string) {
	if c.params == nil {
		c.params = make(map[string]string)
	}
	c.params[key] = value
}

func (c *Context) QueryParam(key string) string {
	return c.Req.URL.Query().Get(key)
}

func (c *Context) QueryParams() map[string]string {
	m := make(map[string]string)
	for k, v := range c.Req.URL.Query() {
		if len(v) > 0 {
			m[k] = v[0]
		}
	}
	return m
}

func (c *Context) QueryAll() url.Values {
	return c.Req.URL.Query()
}

func (c *Context) FormValue(key string) (string, error) {
	if err := c.Req.ParseForm(); err != nil {
		return "", err
	}
	return c.Req.FormValue(key), nil
}

func (c *Context) PostForm() (url.Values, error) {
	if err := c.Req.ParseForm(); err != nil {
		return nil, err
	}
	return c.Req.PostForm, nil
}

func (c *Context) MultipartForm() (*multipart.Form, error) {
	if err := c.Req.ParseMultipartForm(32 << 20); err != nil {
		return nil, err
	}
	return c.Req.MultipartForm, nil
}

func (c *Context) BindJSON(obj any) error {
	if c.Req.Body == nil {
		return fmt.Errorf("request body is nil")
	}
	dec := json.NewDecoder(c.Req.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(obj)
}

func (c *Context) Header(key string) string {
	return c.Req.Header.Get(key)
}

func (c *Context) SetHeader(key, value string) {
	c.Writer.Header().Set(key, value)
}

func (c *Context) Status(code int) {
	c.Writer.WriteHeader(code)
}

func (c *Context) Context() context.Context {
	return c.Req.Context()
}

func (c *Context) Set(key string, value any) {
	c.keys[key] = value
}

func (c *Context) Get(key string) (any, bool) {
	v, ok := c.keys[key]
	return v, ok
}

func (c *Context) HTML(status int, html string) {
	c.SetHeader("Content-Type", "text/html; charset=utf-8")
	c.Status(status)
	fmt.Fprint(c.Writer, html)
}

func (c *Context) String(status int, format string, args ...any) {
	c.SetHeader("Content-Type", "text/plain; charset=utf-8")
	c.Writer.WriteHeader(status)
	fmt.Fprintf(c.Writer, format, args...)
}

func (c *Context) JSON(status int, v any) {
	c.SetHeader("Content-Type", "application/json; charset=utf-8")
	c.Writer.WriteHeader(status)

	enc := json.NewEncoder(c.Writer)
	enc.SetEscapeHTML(true)
	if err := enc.Encode(v); err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
	}
}
