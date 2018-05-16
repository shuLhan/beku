// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//
// Package beku provide library for managing Go packages in GOPATH.
//
package beku

import (
	"bytes"
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

//
// Env contains the environment of Go including GOROOT source directory,
// GOPATH source directory, list of packages in GOPATH, list of standard
// packages, and list of missing packages.
//
type Env struct {
	srcDir      string
	rootSrcDir  string
	defDBFile   string
	pkgs        []*Package
	pkgsMissing []string
	pkgsStd     []string
	db          *ini.Ini
	dbFile      string
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
	Debug = debugMode(debug)

	env = &Env{
		srcDir:     build.Default.GOPATH + "/" + dirSrc,
		rootSrcDir: build.Default.GOROOT + "/" + dirSrc,
		defDBFile:  build.Default.GOPATH + dirDB + "/" + DefDBName,
	}

	err = env.scanStdPackages(env.rootSrcDir)
	if err != nil {
		return
	}

	return
}

//
// GetPackage will return installed package on system.
// If no package found, it will return nil.
//
func (env *Env) GetPackage(importPath, remoteURL string) *Package {
	for x := 0; x < len(env.pkgs); x++ {
		if importPath == env.pkgs[x].ImportPath {
			return env.pkgs[x]
		}

		if remoteURL == env.pkgs[x].RemoteURL {
			return env.pkgs[x]
		}
	}
	return nil
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
		err = fmt.Errorf("scanStdPackages: %s", err)
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
	if Debug >= DebugL2 {
		log.Println("Scanning", rootPath)
	}

	fis, err := ioutil.ReadDir(rootPath)
	if err != nil {
		err = fmt.Errorf("scanPackages: %s", err)
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
// newPackage will append the directory at path as a package only if
// its contain version information.
//
func (env *Env) newPackage(fullPath string, vcsMode VCSMode) (err error) {
	pkgName := strings.TrimPrefix(fullPath, env.srcDir+"/")

	pkg := NewPackage(pkgName, pkgName, vcsMode)

	if Debug >= DebugL2 {
		log.Println("Scanning package:", pkg.ImportPath)
	}

	err = pkg.Scan()
	if err != nil {
		if err == ErrVersion {
			err = nil
			return
		}
		return fmt.Errorf("%s: %s", pkgName, err)
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
		env.dbFile = env.defDBFile
	} else {
		env.dbFile = file
	}

	if Debug >= DebugL1 {
		log.Println("Env.Load:", env.dbFile)
	}

	env.db, err = ini.Open(env.dbFile)
	if err != nil {
		return
	}

	sections := env.db.GetSections(sectionPackage)
	for _, sec := range sections {
		if len(sec.Sub) == 0 {
			log.Println(errDBPackageName, sec.LineNum, env.dbFile)
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
		if len(env.dbFile) == 0 {
			file = env.defDBFile
		} else {
			file = env.dbFile
		}
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

func (env *Env) update(curPkg, newPkg *Package) (ok bool, err error) {
	err = curPkg.Fetch()
	if err != nil {
		return
	}

	if len(newPkg.Version) == 0 {
		newPkg.Version = curPkg.VersionNext
		newPkg.isTag = curPkg.isTag
	}

	if Debug >= DebugL1 {
		log.Println("Sync:\n", newPkg)
	}

	if curPkg.IsEqual(newPkg) {
		fmt.Printf("Nothing to update.\n")
		return
	}

	fmt.Printf("Updating package from,\n%s\nto,\n%s\n", curPkg, newPkg)

	ok = confirm(os.Stdin, msgUpdateView, false)
	if ok {
		err = curPkg.CompareVersion(newPkg)
		if err != nil {
			return
		}
	}

	ok = confirm(os.Stdin, msgUpdateProceed, true)
	if !ok {
		return
	}

	err = curPkg.Update(newPkg)

	return
}

func (env *Env) install(pkg *Package) (ok bool, err error) {
	return
}

//
// updateMissing will remove missing package if it's already provided by new
// package and add it as one of package dependencies.
//
func (env *Env) updateMissing(newPkg *Package) {
	for x := 0; x < len(env.pkgs); x++ {
		env.pkgs[x].UpdateMissingDeps(newPkg)
	}

	var newMissings []string

	for x := 0; x < len(env.pkgsMissing); x++ {
		if strings.HasPrefix(env.pkgsMissing[x], newPkg.ImportPath) {
			continue
		}

		newMissings = append(newMissings, env.pkgsMissing[x])
	}

	env.pkgsMissing = newMissings
}

//
// Sync will download and install a package including their dependencies. If
// the importPath is defined, it will be downloaded into that directory.
//
// (1) First, we check if pkgName contains version.
// (2) And then we check if package already installed, by comparing with
// database.
// (2.1) If package already installed, do an update.
// (2.2) If package is not installed, clone the repository into `importPath`,
// and checkout the latest tag or the latest commit.
//
func (env *Env) Sync(pkgName, importPath string) (err error) {
	err = ErrPackageName

	if len(pkgName) == 0 {
		return
	}

	var (
		ok      bool
		version string
	)

	// (1)
	pkgName, version = parsePkgVersion(pkgName)
	if len(pkgName) == 0 {
		return
	}

	if len(importPath) == 0 {
		importPath = pkgName
	}

	newPkg := NewPackage(pkgName, importPath, VCSModeGit)

	if len(version) > 0 {
		newPkg.Version = version
		newPkg.isTag = IsTagVersion(version)
	}

	// (2)
	curPkg := env.GetPackage(newPkg.ImportPath, newPkg.RemoteURL)
	if curPkg != nil {
		ok, err = env.update(curPkg, newPkg)
	} else {
		ok, err = env.install(newPkg)
	}
	if err != nil {
		return
	}
	if !ok {
		return
	}

	err = env.postSync(curPkg, newPkg)

	return
}

//
// (1) Update missing packages.
// (2) Re-scan package dependencies.
// (3) Run `go install` only if no missing package.
//
func (env *Env) postSync(curPkg, newPkg *Package) (err error) {
	// (1)
	env.updateMissing(newPkg)

	// (2)
	err = curPkg.ScanDeps(env)
	if err != nil {
		return
	}

	// (3)
	if len(curPkg.DepsMissing) == 0 {
		err = curPkg.RunGoInstall(true)
		if err != nil {
			return
		}
	}

	curPkg.ImportPath = newPkg.ImportPath
	curPkg.RemoteName = newPkg.RemoteName
	curPkg.RemoteURL = newPkg.RemoteURL
	curPkg.Version = newPkg.Version
	curPkg.isTag = newPkg.isTag

	if Debug >= DebugL1 {
		log.Printf("Package installed:\n%s", curPkg)
	}

	return
}
