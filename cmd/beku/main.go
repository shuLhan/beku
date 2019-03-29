// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//
// Beku is a command line program to manage packages in user's environment
// (GOPATH or vendor) directory.  Beku provide syntax like `pacman`.
//
// See README in root of this repository for user manual [1].
//
// [1] https://github.com/shuLhan/beku
//
package main

import (
	"fmt"
	"os"
)

const (
	verMajor    = 0
	verMinor    = 6
	verPatch    = 0
	verMetadata = ""
)

func main() {
	cmd, err := newCommand()
	if err != nil {
		if err == errNoDB {
			fmt.Fprintln(os.Stderr, err)
			fmt.Fprintln(os.Stderr, "Run 'beku -S' to initialize database")
		} else {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}

	cmd.env.NoConfirm = cmd.noConfirm

	switch cmd.op {
	case opDatabase | opExclude:
		cmd.env.Exclude(cmd.pkgs)
	case opFreeze:
		err = cmd.env.Freeze()
	case opQuery:
		cmd.env.Query(cmd.pkgs)
	case opRemove:
		err = cmd.env.Remove(cmd.pkgs[0], false)
	case opRemove | opRecursive:
		err = cmd.env.Remove(cmd.pkgs[0], true)
	case opSync:
		err = cmd.sync()
	case opSync | opSyncInto:
		err = cmd.sync()
	case opSync | opUpdate:
		err = cmd.sync()
	default:
		fmt.Fprintln(os.Stderr, errInvalidOptions)
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	err = cmd.env.Save("")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
