// Package alsonow
// Copyright 2025 alsonow. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.
package alsonow

import "net/http"

type Router interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}
