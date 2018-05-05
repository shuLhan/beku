// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//
// Package beku provide library for managing Go packages in GOPATH.
//
package beku

import (
	"bytes"
	"errors"
	"fmt"
	"go/build"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

const (
	envDEBUG = "BEKU_DEBUG"
	gitDir   = ".git"
	srcDir   = "src"
)

// List of error messages.
var (
	ErrGOPATH = errors.New("GOPATH is not defined")
	ErrGOROOT = errors.New("GOROOT is not defined")
)

//
// Env contains the environment of Go including GOROOT source directory,
// GOPATH source directory, list of packages in GOPATH, and list of standard
// packages, and list of missing packages.
//
type Env struct {
	srcDir      string
	rootSrcDir  string
	pkgs        []*Package
	pkgsMissing []string
	pkgsStd     []string
	Debug       debugMode
}

//
// LoadEnv will gather all information in user system to start `beku`-ing.
//
// (0) `beku` required that `$GOPATH` environment variable must exist.
// (1) It will load all standard packages (packages in `$GOROOT/src`)
// (2) It will load all packages in `$GOPATH/src`
// (3) Scan package dependencies and link them
//
func LoadEnv() (*Env, error) {
	// (0)
	if len(build.Default.GOPATH) == 0 {
		return nil, ErrGOPATH
	}
	if len(build.Default.GOROOT) == 0 {
		return nil, ErrGOROOT
	}

	debug, _ := strconv.Atoi(os.Getenv(envDEBUG))

	env := &Env{
		srcDir:     build.Default.GOPATH + "/" + srcDir,
		rootSrcDir: build.Default.GOROOT + "/" + srcDir,
		Debug:      debugMode(debug),
	}

	err := env.scanStdPackages(env.rootSrcDir)
	if err != nil {
		return nil, err
	}

	err = env.scanPackages(env.srcDir)
	if err != nil {
		return nil, err
	}

	for x := 0; x < len(env.pkgs); x++ {
		err = env.pkgs[x].ScanDeps(env)
		if err != nil {
			return nil, err
		}
	}

	return env, nil
}

//
// scanStdPackages will traverse each directory in GOROOT `src` recursively
// until no subdirectory found. All path to subdirectories will be saved on
// Environment `pkgsStd`.
//
// (0) skip file
// (1) skip ignored directory
//
func (env *Env) scanStdPackages(srcPath string) error {
	fis, err := ioutil.ReadDir(srcPath)
	if err != nil {
		return err
	}

	for _, fi := range fis {
		// (0)
		if !fi.IsDir() {
			continue
		}

		dirName := fi.Name()
		fullPath := srcPath + "/" + dirName

		// (1)
		if IsIgnoredDir(dirName) {
			continue
		}

		stdPkg := strings.TrimPrefix(fullPath, env.rootSrcDir+"/")
		env.pkgsStd = append(env.pkgsStd, stdPkg)
	}

	return nil
}

//
// scanPackages will traverse each directory in GOPATH `src` recursively until
// it's found VCS metadata, e.g. `.git` directory.
//
// (0) skip file
// (1) skip ignored directory
// (2) skip directory without `.git`
//
func (env *Env) scanPackages(rootPath string) (err error) {
	if env.Debug >= DebugL2 {
		log.Println("Scanning", rootPath)
	}

	fis, err := ioutil.ReadDir(rootPath)
	if err != nil {
		return
	}

	var nextRoot []string

	for _, fi := range fis {
		// (0)
		if !fi.IsDir() {
			continue
		}

		dirName := fi.Name()
		fullPath := rootPath + "/" + dirName
		dirGit := fullPath + "/" + gitDir

		// (1)
		if IsIgnoredDir(dirName) {
			continue
		}

		// (2)
		_, err = os.Stat(dirGit)
		if err != nil {
			nextRoot = append(nextRoot, fullPath)
			err = nil
			continue
		}

		err = env.newPackage(fullPath, VCSModeGit)
		if err != nil {
			return
		}
	}

	for x := 0; x < len(nextRoot); x++ {
		err = env.scanPackages(nextRoot[x])
		if err != nil {
			return
		}
	}

	return
}

//
// newPackage will append the directory at `fullPath` as a package only if its
// contain version information.
//
func (env *Env) newPackage(fullPath string, vcs VCSMode) (err error) {
	importPath := strings.TrimPrefix(fullPath, env.srcDir+"/")

	pkg, err := NewPackage(env, importPath, fullPath, vcs)
	if err != nil {
		if err == ErrVersion {
			err = nil
			return
		}
		return fmt.Errorf("%s: %s", importPath, err)
	}

	env.pkgs = append(env.pkgs, pkg)

	return nil
}

//
// addPackageMissing will add import path to list of missing package only if
// not exist yet.
//
func (env *Env) addPackageMissing(importPath string) {
	for x := 0; x < len(env.pkgsMissing); x++ {
		if importPath == env.pkgsMissing[x] {
			return
		}
	}

	env.pkgsMissing = append(env.pkgsMissing, importPath)
}

//
// String return formatted output of the environment instance.
//
func (env *Env) String() string {
	var buf bytes.Buffer

	fmt.Fprintf(&buf, `{
   srcDir: %s
  pkgsStd: %s
`, env.srcDir, env.pkgsStd)

	for x := 0; x < len(env.pkgs); x++ {
		fmt.Fprintf(&buf, " pkg %4d: %s\n", x, env.pkgs[x])
	}

	fmt.Fprintf(&buf, "pkgs missing: %s", env.pkgsMissing)

	fmt.Fprintf(&buf, "}")

	return buf.String()
}
