// Package alsonow
// Copyright 2025 alsonow. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.
package alsonow

import (
	"fmt"
	"strings"
	"testing"
)

func TestRouter_normalizePath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "/"},
		{"/", "/"},
		{"home", "/home"},
		{"/home", "/home"},
		{"/home/", "/home"},
		{"  /home/about/  ", "/home/about"},
		{"/home//about///contact", "/home/about/contact"},
		{"home//about///contact///////", "/home/about/contact"},
		{"////", "/"},
		{"  /api//v1/  ", "/api/v1"},
		{"/users/123", "/users/123"},
		{"//home//////////////", "/home"},
		{"/////////////////", "/"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizePath(tt.input)

			if got != "/" {
				segments := strings.Split(got[1:], "/")
				for _, segment := range segments {
					if segment == "" {
						fmt.Println("over", got, segment)
					}
				}
			}

			if got != tt.expected {
				t.Errorf("normalizePath(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
