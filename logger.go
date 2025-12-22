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

		log.Printf("%s %s %v",
			c.Req.Method,
			c.Req.URL.Path,
			duration,
		)
	}
}
