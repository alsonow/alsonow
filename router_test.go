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

func TestRouter_ClearAndLoopDelete(t *testing.T) {
	t.Run("clear and loop delete", func(t *testing.T) {
		m := map[string]int{
			"a": 1,
			"b": 2,
			"c": 3,
		}

		clear(m)
		if len(m) != 0 {
			t.Errorf("expected map to be empty after clear, but got %d elements", len(m))
		}

		m = map[string]int{
			"a": 1,
			"b": 2,
			"c": 3,
		}

		for k := range m {
			delete(m, k)
		}
		if len(m) != 0 {
			t.Errorf("expected map to be empty after loop delete, but got %d elements", len(m))
		}
	})

	t.Run("nil map", func(t *testing.T) {
		var m map[string]string
		clear(m)
	})
}
