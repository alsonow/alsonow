// Package alsonow
// Copyright 2025 alsonow. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.
package alsonow

import (
	"net"
	"net/http"
	"strings"
)

// ClientIP returns the client's real IP address from the request.
// It considers X-Forwarded-For, X-Real-IP, and RemoteAddr headers.
func ClientIP(r *http.Request) string {
	// Check the X-Forwarded-For header
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		// Handle multiple IPs in the X-Forwarded-For header.
		if idx := strings.Index(ip, ","); idx > 0 {
			ip = ip[:idx]
		}
		ip = strings.TrimSpace(ip)
		if isValidIP(ip) {
			return ip
		}
	}

	// Check the X-Real-IP header
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		ip = strings.TrimSpace(ip)
		if isValidIP(ip) {
			return ip
		}
	}

	// Fallback to RemoteAddr if no other headers are found.
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "unknown"
	}

	if isValidIP(ip) {
		return ip
	}

	return "unknown"
}

// isValidIP validates if the given string is a valid IP address (either IPv4 or IPv6).
func isValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}
