// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//
// Beku is a command line program to manage packages in $GOPATH. Beku provide
// syntax like `pacman` [1].
//
// Beku read and write the package database into a file named "gopath.deps".
//
// At first execution, beku will try to open the package database in current
// directory. If no file found, it will try to open
// "$GOPATH/var/beku/gopath.deps". When both locations does not provide
// package database, beku will scan entire "$GOPATH/src" and write the
// package database into "$GOPATH/var/beku/gopath.deps".
//
// [1] https://www.archlinux.org/pacman/
//
package main

import (
	"log"

	"github.com/shuLhan/beku"
)

const (
	logPrefix = "beku - "
)

var (
	env *beku.Env
)

func main() {
	var err error

	log.SetPrefix(logPrefix)

	env, err = beku.NewEnvironment()
	if err != nil {
		log.Fatal(err)
	}

	err = loadDatabase()
	if err != nil {
		log.Println("No database found.")
		log.Println("Initializing database for the first time...")
		firstTime()
	}

	if env.Debug >= beku.DebugL1 {
		log.Printf("Environment: %s", env)
	}
}

func loadDatabase() (err error) {
	err = env.Load(beku.DefDBName)
	if err == nil {
		return
	}

	err = env.Load("")

	return
}

func firstTime() {
	err := env.Scan()
	if err != nil {
		log.Fatal("Scan:", err)
	}

	err = env.Save("")
	if err != nil {
		log.Fatal("Save:", err)
	}

	log.Println("Initialization complete.")
}
