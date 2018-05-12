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
	"path/filepath"
	"strconv"
	"strings"

	"github.com/shuLhan/share/lib/ini"
)

const (
	// DefDBName define default database name, where the dependencies will
	// be saved and loaded.
	DefDBName = "gopath.deps"
)

const (
	envDEBUG = "BEKU_DEBUG"
	gitDir   = ".git"
	srcDir   = "src"
	defDBDir = "/var/beku"
)

// List of error messages.
var (
	ErrGOPATH = errors.New("GOPATH is not defined")
	ErrGOROOT = errors.New("GOROOT is not defined")

	errDBPackageName = "missing package name, line %d at %s"
)

var (
	sectionPackage = "package"
)

//
// Env contains the environment of Go including GOROOT source directory,
// GOPATH source directory, list of packages in GOPATH, list of standard
// packages, and list of missing packages.
//
type Env struct {
	srcDir      string
	rootSrcDir  string
	defDB       string
	pkgs        []*Package
	pkgsMissing []string
	pkgsStd     []string
	db          *ini.Ini
	Debug       debugMode
}

// NewEnvironment will gather all information in user system.
// `beku` required that `$GOPATH` environment variable must exist.
//
func NewEnvironment() (env *Env, err error) {
	if len(build.Default.GOPATH) == 0 {
		return nil, ErrGOPATH
	}
	if len(build.Default.GOROOT) == 0 {
		return nil, ErrGOROOT
	}

	debug, _ := strconv.Atoi(os.Getenv(envDEBUG))

	env = &Env{
		srcDir:     build.Default.GOPATH + "/" + srcDir,
		rootSrcDir: build.Default.GOROOT + "/" + srcDir,
		defDB:      build.Default.GOPATH + defDBDir + "/" + DefDBName,
		Debug:      debugMode(debug),
	}

	err = env.scanStdPackages(env.rootSrcDir)
	if err != nil {
		return
	}

	return
}

//
// Scan will gather all information in user system to start `beku`-ing.
//
// (1) It will load all standard packages (packages in `$GOROOT/src`)
// (2) It will load all packages in `$GOPATH/src`
// (3) Scan package dependencies and link them
//
func (env *Env) Scan() (err error) {
	err = env.scanPackages(env.srcDir)
	if err != nil {
		return
	}

	for x := 0; x < len(env.pkgs); x++ {
		err = env.pkgs[x].ScanDeps(env)
		if err != nil {
			return
		}
	}

	return
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

func (env *Env) addPackage(pkg *Package) {
	for x := 0; x < len(pkg.DepsMissing); x++ {
		env.addPackageMissing(pkg.DepsMissing[x])
	}

	env.pkgs = append(env.pkgs, pkg)
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
// Load will read saved dependencies from file.
//
func (env *Env) Load(file string) (err error) {
	if len(file) == 0 {
		file = env.defDB
	}

	if env.Debug >= DebugL1 {
		log.Println("Env.Load:", file)
	}

	env.db, err = ini.Open(file)
	if err != nil {
		return
	}

	sections := env.db.GetSections(sectionPackage)
	for _, sec := range sections {
		if len(sec.Sub) == 0 {
			log.Println(errDBPackageName, sec.LineNum, file)
			continue
		}

		pkg := &Package{
			ImportPath: sec.Sub,
			FullPath:   env.srcDir + "/" + sec.Sub,
		}

		pkg.load(sec)

		env.addPackage(pkg)
	}

	return
}

//
// Save the dependencies to `file`.
//
func (env *Env) Save(file string) (err error) {
	if len(file) == 0 {
		file = env.defDB
	}

	dir := filepath.Dir(file)

	err = os.MkdirAll(dir, 0700)
	if err != nil {
		return
	}

	env.db = &ini.Ini{}

	for _, pkg := range env.pkgs {
		sec := ini.NewSection(sectionPackage, pkg.ImportPath)

		switch pkg.vcs {
		case VCSModeGit:
			sec.Set(keyVCSMode, valVCSModeGit)
		}

		sec.Set(keyRemoteName, pkg.RemoteName)
		sec.Set(keyRemoteURL, pkg.RemoteURL)
		sec.Set(keyVersion, pkg.Version)

		for _, dep := range pkg.Deps {
			sec.Add(keyDeps, dep)
		}
		for _, req := range pkg.RequiredBy {
			sec.Add(keyRequiredBy, req)
		}
		for _, mis := range pkg.DepsMissing {
			sec.Add(keyDepsMissing, mis)
		}

		sec.AddNewLine()

		env.db.AddSection(sec)
	}

	err = env.db.Save(file)

	return
}

//
// String return formatted output of the environment instance.
//
func (env *Env) String() string {
	var buf bytes.Buffer

	fmt.Fprintf(&buf, `
         GOPATH src: %s
  Standard Packages: %s
`, env.srcDir, env.pkgsStd)

	for x := 0; x < len(env.pkgs); x++ {
		fmt.Fprintf(&buf, "%s", env.pkgs[x])
	}

	fmt.Fprintf(&buf, "\n[package \"_missing_\"]\n")

	for x := range env.pkgsMissing {
		fmt.Fprintln(&buf, "ImportPath =", env.pkgsMissing[x])
	}

	return buf.String()
}
