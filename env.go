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
	dirBin      string
	dirPkg      string
	dirRootSrc  string
	dirSrc      string
	pkgs        []*Package
	pkgsMissing []string
	pkgsStd     []string
	db          *ini.Ini
	dbDefFile   string
	dbFile      string
	countNew    int
	countUpdate int
	fmtMaxPath  int
	dirty       bool
}

//
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
		dirSrc:     filepath.Join(build.Default.GOPATH, dirSrc),
		dirRootSrc: filepath.Join(build.Default.GOROOT, dirSrc),
		dirBin:     filepath.Join(build.Default.GOPATH, dirBin),
		dirPkg: filepath.Join(build.Default.GOPATH, dirPkg,
			build.Default.GOOS+"_"+build.Default.GOARCH),
		dbDefFile: filepath.Join(build.Default.GOPATH, dirDB, DefDBName),
	}

	err = env.scanStdPackages(env.dirRootSrc)
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
// Scan will gather all package information in user system to start `beku`-ing.
//
func (env *Env) Scan() (err error) {
	err = env.scanPackages(env.dirSrc)
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
		fullPath := filepath.Join(srcPath, dirName)

		// (1)
		if IsIgnoredDir(dirName) {
			continue
		}

		stdPkg := strings.TrimPrefix(fullPath, env.dirRootSrc+"/")
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
		fmt.Println(">>> Scanning", rootPath)
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
		fullPath := filepath.Join(rootPath, dirName)
		dirGit := filepath.Join(fullPath, gitDir)

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
	pkgName := strings.TrimPrefix(fullPath, env.dirSrc+"/")

	pkg := NewPackage(pkgName, pkgName, vcsMode)

	if Debug >= DebugL2 {
		fmt.Println(">>> Scanning package:", pkg.ImportPath)
	}

	err = pkg.Scan()
	if err != nil {
		if err == ErrVersion {
			err = nil
			return
		}
		return fmt.Errorf("%s: %s", pkgName, err)
	}

	curPkg := env.GetPackage(pkg.ImportPath, pkg.RemoteURL)
	if curPkg == nil {
		env.pkgs = append(env.pkgs, pkg)
		env.countNew++

		if len(pkg.ImportPath) > env.fmtMaxPath {
			env.fmtMaxPath = len(pkg.ImportPath)
		}
	} else {
		if curPkg.Version != pkg.Version {
			curPkg.VersionNext = pkg.Version
			curPkg.state = packageStateChange
			env.countUpdate++
		}
	}

	return nil
}

func (env *Env) addPackage(pkg *Package) {
	for x := 0; x < len(pkg.DepsMissing); x++ {
		env.addPackageMissing(pkg.DepsMissing[x])
	}

	env.pkgs = append(env.pkgs, pkg)

	if len(pkg.ImportPath) > env.fmtMaxPath {
		env.fmtMaxPath = len(pkg.ImportPath)
	}
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
		env.dbFile = env.dbDefFile
	} else {
		env.dbFile = file
	}

	if Debug >= DebugL1 {
		fmt.Println(">>> Env.Load:", env.dbFile)
	}

	env.db, err = ini.Open(env.dbFile)
	if err != nil {
		return
	}

	sections := env.db.GetSections(sectionPackage)
	for _, sec := range sections {
		if len(sec.Sub) == 0 {
			fmt.Fprintln(os.Stderr, errDBPackageName, sec.LineNum, env.dbFile)
			continue
		}

		pkg := &Package{
			ImportPath: sec.Sub,
			FullPath:   filepath.Join(env.dirSrc, sec.Sub),
			state:      packageStateLoad,
		}

		pkg.load(sec)

		env.addPackage(pkg)
	}

	return
}

//
// Query the package database. If package is not empty, it will only show the
// information about that package.
//
func (env *Env) Query(pkgs []string) {
	format := fmt.Sprintf("%%-%ds  %%s\n", env.fmtMaxPath)

	for x := 0; x < len(env.pkgs); x++ {
		if len(pkgs) == 0 {
			fmt.Fprintf(defStdout, format, env.pkgs[x].ImportPath,
				env.pkgs[x].Version)
			continue
		}
		for y := 0; y < len(pkgs); y++ {
			if env.pkgs[x].ImportPath == pkgs[y] {
				fmt.Fprintf(defStdout, format,
					env.pkgs[x].ImportPath,
					env.pkgs[x].Version)
			}
		}
	}
}

//
// Rescan GOPATH for new packages.
//
func (env *Env) Rescan() (err error) {
	err = env.Scan()
	if err != nil {
		return
	}

	format := fmt.Sprintf("%%-%ds  %%-12s  %%-12s\n", env.fmtMaxPath)

	if env.countUpdate > 0 {
		fmt.Printf(">>> The following packages will be updated,\n\n")
		fmt.Printf(format+"\n", "ImportPath", "Old Version", "New Version")

		for _, pkg := range env.pkgs {
			if pkg.state&packageStateChange == 0 {
				continue
			}

			fmt.Printf(format, pkg.ImportPath, pkg.Version, pkg.VersionNext)
		}
	}
	if env.countNew > 0 {
		fmt.Printf("\n>>> New packages,\n\n")
		fmt.Printf(format+"\n", "ImportPath", "Old Version", "New Version")

		for _, pkg := range env.pkgs {
			if pkg.state&packageStateNew == 0 {
				continue
			}

			fmt.Printf(format, pkg.ImportPath, "-", pkg.Version)
		}
	}

	if env.countUpdate == 0 && env.countNew == 0 {
		fmt.Println(">>> Database and GOPATH is in sync.")
		return
	}

	fmt.Println()

	ok := confirm(os.Stdin, msgContinue, false)
	if !ok {
		return
	}

	if env.countUpdate > 0 {
		for x, pkg := range env.pkgs {
			if pkg.state&packageStateChange == 0 {
				continue
			}

			env.pkgs[x].Version = env.pkgs[x].VersionNext
			env.pkgs[x].VersionNext = ""
			env.pkgs[x].state = packageStateDirty
		}
	}
	if env.countNew > 0 {
		for _, pkg := range env.pkgs {
			if pkg.state&packageStateNew == 0 {
				continue
			}
			pkg.state = packageStateDirty
			env.updateMissing(pkg)
		}
	}
	env.dirty = true

	return
}

//
// Remove package from GOPATH. If recursive is true, it will also remove their
// dependencies, as long as they are not required by other package.
//
func (env *Env) Remove(rmPkg string, recursive bool) (err error) {
	pkg := env.GetPackage(rmPkg, "")
	if pkg == nil {
		fmt.Println("Package", rmPkg, "not installed")
		return
	}

	if len(pkg.RequiredBy) > 0 {
		fmt.Fprintln(os.Stderr, `Can't remove package.
This package is required by,
`,
			pkg.RequiredBy)
		return
	}

	var listRemoved []string
	tobeRemoved := make(map[string]bool)

	if recursive {
		env.filterUnusedDeps(pkg, tobeRemoved)
	}

	for k, v := range tobeRemoved {
		if v {
			if k == pkg.ImportPath {
				continue
			}
			listRemoved = append(listRemoved, k)
		}
	}
	listRemoved = append(listRemoved, pkg.ImportPath)

	fmt.Println("The following package will be removed,")
	for _, importPath := range listRemoved {
		fmt.Println(" *", importPath)
	}

	ok := confirm(os.Stdin, msgContinue, false)
	if !ok {
		return
	}

	for _, importPath := range listRemoved {
		err = env.removePackage(importPath)
		if err != nil {
			err = fmt.Errorf("Remove: %s", err)
			return
		}

		pkgImportPath := filepath.Join(env.dirPkg, importPath)

		if Debug >= DebugL1 {
			fmt.Println(">>> Removing", pkgImportPath)
		}

		err = os.RemoveAll(pkgImportPath)
		if err != nil {
			err = fmt.Errorf("Remove: %s", err)
			return
		}

		_ = RmdirEmptyAll(pkgImportPath)
	}

	return
}

func (env *Env) filterUnusedDeps(pkg *Package, tobeRemoved map[string]bool) {
	var dep *Package
	var nfound int

	_, ok := tobeRemoved[pkg.ImportPath]
	if ok {
		return
	}

	tobeRemoved[pkg.ImportPath] = true
	for x := 0; x < len(pkg.Deps); x++ {
		tobeRemoved[pkg.Deps[x]] = true
	}

	for x := 0; x < len(pkg.Deps); x++ {
		dep = env.GetPackage(pkg.Deps[x], "")

		if len(dep.Deps) > 0 {
			env.filterUnusedDeps(dep, tobeRemoved)
		}

		if len(dep.RequiredBy) == 1 {
			continue
		}

		nfound = 0
		for y := 0; y < len(dep.RequiredBy); y++ {
			found, ok := tobeRemoved[dep.RequiredBy[y]]
			if ok && found {
				nfound++
			}
		}
		if nfound == len(dep.RequiredBy) {
			continue
		}

		tobeRemoved[pkg.Deps[x]] = false
	}
}

//
// removePackage from list environment (including source and installed archive
// or binary). This also remove in other packages "RequiredBy" if exist.
//
func (env *Env) removePackage(importPath string) (err error) {
	pkg := env.GetPackage(importPath, "")
	if pkg == nil {
		return
	}

	err = pkg.Remove()
	if err != nil {
		return
	}

	idx := -1
	for x := 0; x < len(env.pkgs); x++ {
		if env.pkgs[x].ImportPath == importPath {
			idx = x
			continue
		}

		ok := env.pkgs[x].RemoveRequiredBy(importPath)
		if ok {
			env.dirty = true
		}
	}

	if idx < 0 {
		return
	}

	lenpkgs := len(env.pkgs)

	copy(env.pkgs[idx:], env.pkgs[idx+1:])
	env.pkgs[lenpkgs-1] = nil
	env.pkgs = env.pkgs[:lenpkgs-1]

	env.dirty = true

	return
}

//
// Save the dependencies to `file` only if it's dirty flag is true.
//
func (env *Env) Save(file string) (err error) {
	if !env.dirty {
		return
	}

	if len(file) == 0 {
		if len(env.dbFile) == 0 {
			file = env.dbDefFile
		} else {
			file = env.dbFile
		}
	}

	if Debug >= DebugL1 {
		fmt.Println(">>> Saving db", file)
	}

	dir := filepath.Dir(file)

	if Debug >= DebugL1 {
		fmt.Println(">>> Save: MkdirAll:", dir)
	}

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
`, env.dirSrc, env.pkgsStd)

	for x := 0; x < len(env.pkgs); x++ {
		fmt.Fprintf(&buf, "%s", env.pkgs[x].String())
	}

	fmt.Fprintf(&buf, "\n[package \"_missing_\"]\n")

	for x := range env.pkgsMissing {
		fmt.Fprintln(&buf, "ImportPath =", env.pkgsMissing[x])
	}

	return buf.String()
}

//
// install a package.
//
func (env *Env) install(pkg *Package) (ok bool, err error) {
	err = pkg.Install()
	if err != nil {
		_ = pkg.Remove()
		return
	}

	ok = true

	return
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
		fmt.Println(">>> Sync:\n", newPkg)
	}

	if curPkg.IsEqual(newPkg) {
		fmt.Printf("Nothing to update.\n")
		ok = true
		return
	}

	fmt.Printf("Updating package from,\n%s\nto,\n%s\n", curPkg.String(),
		newPkg.String())

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
	if err != nil {
		return
	}

	env.dirty = true

	return
}

//
// updateMissing will remove missing package if it's already provided by new
// package and add it as one of package dependencies.
//
func (env *Env) updateMissing(newPkg *Package) {
	var updated bool

	if Debug >= DebugL1 {
		fmt.Println(">>> Update missing:", newPkg.ImportPath)
	}

	for x := 0; x < len(env.pkgs); x++ {
		updated = env.pkgs[x].UpdateMissingDep(newPkg)
		if updated {
			env.dirty = true
		}
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
	if curPkg == nil {
		curPkg = newPkg
		env.addPackage(newPkg)
		env.dirty = true
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
		err = curPkg.GoInstall(true)
		if err != nil {
			return
		}
	}

	fmt.Println(">>> Package installed:\n", curPkg)

	return
}
