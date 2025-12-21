// Package alsonow
// Copyright 2025 alsonow. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.
package alsonow

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type AlsoNow struct {
	Router
	stop chan struct{}
}

func New() *AlsoNow {
	an := &AlsoNow{
		stop: make(chan struct{}),
	}
	return an
}

func (an *AlsoNow) Run() {
	fmt.Println("🌠 Also now.")
	host := "0.0.0.0"
	port := "1221"

	if addr := os.Getenv("ALSONOW_ADDR"); addr != "" {
		fields := strings.SplitN(addr, ":", 2)
		host = fields[0]
		port = fields[1]
	}

	addr := host + ":" + port
	fmt.Println("Serving on http://localhost:" + port)

	server := &http.Server{
		Addr:              addr,
		Handler:           an,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Faild to start server, err:%v ", err)
		}
	}()

	<-an.stop

	// Block and wait for a signal, then shut down the server after receiving it.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // Block until a termination signal(SIGINT or SIGTERM) is received.

	if err := server.Shutdown(context.Background()); err != nil {
		log.Fatalf("Faild to shutdown server, err:%v ", err)
	}

	log.Println("Server stopped.")
}

func (an *AlsoNow) Stop() {
	select {
	case an.stop <- struct{}{}:
	default:
	}
}
