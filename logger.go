// Package alsonow
// Copyright 2025 alsonow. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.
package alsonow

import (
	"log"
	"net/http"
	"strings"
	"time"
)

func Logger() HandlerFunc {
	return func(c *Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)
		method := c.Req.Method
		path := c.Req.URL.Path

		clientIP := ClientIP(c.Req)
		userAgent := c.Req.UserAgent()

		log.Printf("[ACCESS] %s | %v | %s | %s %s | %s",
			time.Now().Format("2006/01/02 15:04:05"),
			duration,
			clientIP,
			method,
			path,
			userAgent,
		)
	}
}

func ClientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		if idx := strings.Index(ip, ","); idx > 0 {
			return strings.TrimSpace(ip[:idx])
		}
		return strings.TrimSpace(ip)
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return strings.TrimSpace(ip)
	}
	return r.RemoteAddr[:strings.LastIndex(r.RemoteAddr, ":")]
}
