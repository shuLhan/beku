// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"log"

	"github.com/shuLhan/beku"
)

const (
	logPrefix = "beku - "
)

func main() {
	log.SetPrefix(logPrefix)

	env, err := beku.LoadEnv()
	if err != nil {
		log.Fatal(err)
	}

	if env.Debug >= beku.DebugL1 {
		log.Printf("Environment: %s\n", env)
	}
}
