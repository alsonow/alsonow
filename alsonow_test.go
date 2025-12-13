// Package alsonow
// Copyright 2025 alsonow. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.
package alsonow

import (
	"os"
	"testing"
	"time"
)

func TestAlsoNowRun(t *testing.T) {
	_ = os.Setenv("ALSONOW_ADDR", "0.0.0.0:2025")
	an := New()
	go an.Run()
	time.Sleep(3 * time.Second)
	an.Stop()
}
