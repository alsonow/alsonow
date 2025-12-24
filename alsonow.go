// Package alsonow
// Copyright 2025 alsonow. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.
package alsonow

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type AlsoNow struct {
	Router
	server   *http.Server
	stop     chan struct{}
	stopOnce sync.Once
}

// New returns a new AlsoNow instance.
func New() *AlsoNow {
	router := newRouter()
	an := &AlsoNow{
		Router: router,
		stop:   make(chan struct{}),
		server: &http.Server{
			ReadHeaderTimeout: 10 * time.Second,
			ReadTimeout:       30 * time.Second,
			WriteTimeout:      30 * time.Second,
			IdleTimeout:       90 * time.Second,
		},
	}

	an.server.Handler = an
	an.Use(Recover())

	return an
}

func (an *AlsoNow) WithLogger() *AlsoNow {
	an.Use(Logger())
	return an
}

func (an *AlsoNow) WithServer(server *http.Server) *AlsoNow {
	if server != nil {
		if server.Handler == nil {
			server.Handler = an
		}
		an.server = server
	}
	return an
}

func formatListenURL(addr string, isTLS bool) string {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}

	scheme := "http"
	if isTLS {
		scheme = "https"
	}

	if host == "" || host == "0.0.0.0" || host == "::" {
		host = "localhost"
	}

	if (isTLS && port == "443") || (!isTLS && port == "80") {
		return fmt.Sprintf("%s://%s", scheme, host)
	}

	return fmt.Sprintf("%s://%s:%s", scheme, host, port)
}

func (an *AlsoNow) Run(addr ...string) {
	runAddr := ":1221"

	if len(addr) > 0 && addr[0] != "" {
		runAddr = addr[0]
	} else if env := os.Getenv("ALSONOW_ADDR"); env != "" {
		runAddr = env
	}

	an.server.Addr = runAddr
	log.Printf("🌠 AlsoNow starting on %s", formatListenURL(runAddr, false))

	go func() {
		if err := an.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	an.waitStopSignal()
}

func (an *AlsoNow) RunTLS(addr, certFile, keyFile string) {
	if addr == "" {
		addr = ":443"
	}

	an.server.Addr = addr
	an.server.TLSConfig = &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	log.Printf("🌠 AlsoNow starting on %s", formatListenURL(addr, true))

	go func() {
		if err := an.server.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
			log.Fatalf("TLS Server error: %v", err)
		}
	}()

	an.waitStopSignal()
}

func (an *AlsoNow) waitStopSignal() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-an.stop:
		log.Println("Received Stop() call")
	case s := <-sig:
		log.Printf("Received signal: %v, shutting down gracefully...", s)
	}

	log.Println("Shutting down server, will timeout after 30 seconds...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := an.server.Shutdown(ctx); err != nil {
		log.Printf("Forced shutdown: %v", err)
	} else {
		log.Println("Server stopped gracefully.")
	}
}

func (an *AlsoNow) Stop() {
	an.stopOnce.Do(func() {
		close(an.stop)
	})
}
