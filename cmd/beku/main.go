// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//
// Beku is a command line program to manage packages in $GOPATH. Beku provide
// syntax like `pacman`.
//
// Beku read and write the package database into a file named "gopath.deps".
//
// At first execution, beku will try to open the package database in current
// directory. If no file found, it will try to open
// "$GOPATH/var/beku/gopath.deps". When both locations does not provide
// package database, beku will scan entire "$GOPATH/src" and write the
// package database into "$GOPATH/var/beku/gopath.deps".
//
// ## Query Operation
//
//	-Q, --query [pkg ...]
//
// Query the package database.
//
//
// ## Sync Operation
//
//	-S, --sync <pkg[@version]>
//
// Synchronizes package. Given a package import path, beku will try to clone
// the package into GOPATH source directory and set the package version to
// latest the tag. If no tag found, it will use the latest commit on master
// branch. A specific version can be set using "@version" suffix.
//
// If package already exist, it will reset the HEAD to the version that is set
// on database file.
//
// Sync operation will not install missing dependencies.
//
// ### Options
//
//	[--into <destination>]
//
// This option will install the package import path into custom directory.
// It is useful if you have the fork of the main package but want to install
// it to the legacy directory.
//
// ### Examples
//
//	beku -S golang.org/x/text
//
// Download package `golang.org/x/text` into `$GOPATH/src/golang.org/x/text`,
// and set their version to the latest commit on branch master.
//
//	beku -S github.com/golang/text --into golang.org/x/text
//
// Download package `github.com/golang/text` into
// `$GOPATH/src/golang.org/x/text`, and set their version to the latest commit
// on branch master.
//
//	beku -S golang.org/x/text@v0.3.0
//
// Download package `golang.org/x/text` into `$GOPATH/src/golang.org/x/text`
// and checkout the tag `v0.3.0` as the working version.
//
//	beku -S golang.org/x/text@5c1cf69
//
// Download package `golang.org/x/text` into `$GOPATH/src/golang.org/x/text`
// and checkout the commit `5c1cf69` as the working version.
//
//
// # Known Limitations
//
// * Only work with package hosted with Git on HTTPS or SSH.
//
// * Tested only on package hosted on Github.
//
// # References
//
// [1] https://www.archlinux.org/pacman/
//
package main

import (
	"log"
)

var (
	cmd *command
)

func main() {
	err := newCommand()
	if err != nil {
		log.Fatal(err)
	}

	switch cmd.op {
	case opQuery:
		cmd.env.Query(cmd.queryPkg)
	case opSync:
		err = cmd.env.Sync(cmd.syncPkg, "")
	case opSync | opSyncInto:
		err = cmd.env.Sync(cmd.syncPkg, cmd.syncInto)
	}

	if err != nil {
		log.Fatal(err)
	}

	cmd.env.Save("")
}
