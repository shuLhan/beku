// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package beku

import (
	"errors"
	"os"
)

const (
	// DefDBName define default database name, where the dependencies will
	// be saved and loaded.
	DefDBName = "beku.db"
)

const (
	defPATH = "/bin:/usr/bin:/usr/sbin:/usr/local/bin:/usr/local/sbin"

	dirDB       = "var/beku"
	dirBin      = "bin"
	dirPkg      = "pkg"
	dirSrc      = "src"
	dirTestdata = "testdata"
	dirVendor   = "vendor"

	envPATH = "PATH"

	msgCleanDir      = "Clean destination directory?"
	msgContinue      = "Continue?"
	msgUpdateProceed = "Proceed with update?"
	msgUpdateView    = "View commit logs?"

	prefixTag = 'v'

	sepImport        = "/"
	sepImportVersion = '@'
	sepVersion       = '.'
)

const (
	sectionBeku    = "beku"
	sectionPackage = "package"

	keyExclude = "exclude"

	keyDeps         = "deps"
	keyDepsMissing  = "missing"
	keyRemoteName   = "remote-name"
	keyRemoteURL    = "remote-url"
	keyRemoteBranch = "remote-branch"
	keyRequiredBy   = "required-by"
	keyVCSMode      = "vcs"
	keyVersion      = "version"

	gitDefBranch     = "master"
	gitDefRemoteName = "origin"
	gitDir           = ".git"
)

// List of error messages.
var (
	ErrGOROOT = errors.New("GOROOT is not defined")

	// ErrVersion define an error when directory have VCS metadata (e.g.
	// `.git` directory) but did not have any tag or commit.
	ErrVersion = errors.New("no tag or commit found")

	// ErrRemote define an error when package remote URL is empty or
	// invalid.
	ErrRemote = errors.New("empty or invalid remote URL found")

	// ErrPackageName define an error if package name is empty or invalid.
	ErrPackageName = errors.New("empty or invalid package name")

	errDirNotEmpty = "directory %s is not empty"
	errExcluded    = "package '%s' is in excluded list\n"
	errVCS         = "unknown VCS mode %s"
)

var (
	defStdout = os.Stdout //nolint: gochecknoglobals
	defStderr = os.Stderr //nolint: gochecknoglobals
)
