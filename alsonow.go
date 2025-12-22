// Package alsonow
// Copyright 2025 alsonow. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.
package alsonow

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type AlsoNow struct {
	Router
	stop         chan struct{}
	serverConfig *http.Server
}

func New() *AlsoNow {
	router := &routerImpl{
		routes: make(map[string]map[string][]HandlerFunc),
	}

	an := &AlsoNow{
		Router: router,
		stop:   make(chan struct{}),
		serverConfig: &http.Server{
			ReadHeaderTimeout: 5 * time.Second,
		},
	}
	return an
}

func (an *AlsoNow) WithServerConfig(cfg *http.Server) *AlsoNow {
	if cfg != nil {
		an.serverConfig = cfg
	}
	return an
}

func (an *AlsoNow) Run() {
	fmt.Println("🌠 Also now.")
	host := "0.0.0.0"
	port := "1221"

	if addr := os.Getenv("ALSONOW_ADDR"); addr != "" {
		h, p, err := net.SplitHostPort(addr)
		if err != nil {
			log.Printf("Invalid ALSONOW_ADDR %q, using default: %v", addr, err)
		} else {
			host, port = h, p
		}
	}

	addr := fmt.Sprintf("%s:%s", host, port)
	fmt.Printf("Listening on http://%s:%s\n", host, port)

	server := an.serverConfig
	server.Handler = an
	server.Addr = addr

	listenErr := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			listenErr <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-an.stop:
		log.Println("Received Stop() signal")
	case <-quit:
		log.Println("Received system interrupt (SIGINT/SIGTERM)")
	case err := <-listenErr:
		log.Fatalf("Server listen error: %v", err)
	}

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced shutdown: %v", err)
	}

	log.Println("Server stopped.")
}

func (an *AlsoNow) Stop() {
	close(an.stop)
}
