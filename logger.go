// Package alsonow
// Copyright 2025 alsonow. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.
package alsonow

import (
	"log"
	"time"
)

func Logger() HandlerFunc {
	return func(c *Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)

		clientIP := ClientIP(c.Req)
		userAgent := c.Req.UserAgent()

		log.Printf("[ACCESS] %s | %v | %s | %s %s | %s",
			time.Now().Format("2006/01/02 15:04:05"),
			duration,
			clientIP,
			c.Method(),
			c.Path(),
			userAgent,
		)
	}
}
