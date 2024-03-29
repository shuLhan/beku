// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package beku provide library for managing Go packages in user's environment
// (GOPATH directory).
package beku

import (
	"bytes"
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/shuLhan/share/lib/debug"
	"github.com/shuLhan/share/lib/ini"
	libio "github.com/shuLhan/share/lib/io"
)

// Env contains the environment of Go including GOROOT source directory,
// package root directory (prefix), list of packages, list of standard
// packages, and list of missing packages.
type Env struct {
	path   string
	prefix string // Equal to GOPATH.

	dirBin       string
	dirGoRootSrc string
	dirPkg       string
	dirSrc       string

	dbDefFile string
	dbFile    string

	pkgs        []*Package
	pkgsExclude []string
	pkgsMissing []string
	pkgsStd     []string
	pkgsUnused  []*Package

	db *ini.Ini

	countNew    int
	countUpdate int
	fmtMaxPath  int

	dirty     bool
	NoConfirm bool
	noDeps    bool
}

// NewEnvironment will gather all information in user system.
func NewEnvironment(noDeps bool) (env *Env, err error) {
	if len(build.Default.GOROOT) == 0 {
		return nil, ErrGOROOT
	}

	env = &Env{
		path:   os.Getenv(envPATH),
		prefix: build.Default.GOPATH,

		dirBin:       filepath.Join(build.Default.GOPATH, dirBin),
		dirGoRootSrc: filepath.Join(build.Default.GOROOT, dirSrc),
		dirPkg:       filepath.Join(build.Default.GOPATH, dirPkg, build.Default.GOOS+"_"+build.Default.GOARCH),
		dirSrc:       filepath.Join(build.Default.GOPATH, dirSrc),

		dbDefFile: filepath.Join(build.Default.GOPATH, dirDB, DefDBName),

		noDeps: noDeps,
	}

	if len(env.path) == 0 {
		env.path = defPATH
	}

	err = env.scanStdPackages(env.dirGoRootSrc)
	if err != nil {
		return nil, err
	}

	return env, nil
}

func (env *Env) initGopath() {
}

// addExclude will add package to list of excluded packages. It will return
// true if importPath is not already exist in list; otherwise it will return
// false.
func (env *Env) addExclude(importPath string) bool {
	if len(importPath) == 0 {
		return false
	}
	for x := 0; x < len(env.pkgsExclude); x++ {
		if env.pkgsExclude[x] == importPath {
			return false
		}
	}

	env.pkgsExclude = append(env.pkgsExclude, importPath)
	return true
}

// Exclude mark list of packages to be excluded from future operations.
func (env *Env) Exclude(importPaths []string) {
	exPkg := new(Package)

	for _, exImportPath := range importPaths {
		ok := env.addExclude(exImportPath)
		if ok {
			env.dirty = true
		}

		pkgIdx, pkg := env.GetPackageFromDB(exImportPath, "")
		if pkg != nil {
			env.removePkgFromDBByIdx(pkgIdx)
		}

		exPkg.ImportPath = exImportPath
		env.updateMissing(exPkg, false)

		env.removeRequiredBy(exImportPath)
	}
}

// Freeze all packages in database. Install all registered packages in
// database and remove non-registered from "src" and "pkg" directories.
func (env *Env) Freeze() (err error) {
	var localPkg *Package

	for _, pkg := range env.pkgs {
		fmt.Printf("\n[ENV] Freeze >>> %s@%s\n", pkg.ImportPath, pkg.Version)

		localPkg, err = env.GetLocalPackage(pkg.ImportPath)
		if err != nil {
			return
		}
		if localPkg == nil {
			err = pkg.Install()
			if err != nil {
				return
			}
			continue
		}

		err = pkg.Freeze()
		if err != nil {
			return
		}
	}

	env.pkgsUnused = nil

	err = env.GetUnused(env.dirSrc)
	if err != nil {
		err = fmt.Errorf("Freeze: %s", err.Error())
		return
	}

	if len(env.pkgsUnused) == 0 {
		fmt.Println("\n[ENV] Freeze >>> No unused packages found.")
		goto out
	}

	fmt.Println("[ENV] Freeze >>> The following packages is unused,")
	for _, pkg := range env.pkgsUnused {
		fmt.Printf("  * %s\n", pkg.ImportPath)
	}

	fmt.Println()
out:
	fmt.Println("[ENV] Freeze >>> finished")

	return nil
}

// GetLocalPackage will return installed package from system.
func (env *Env) GetLocalPackage(importPath string) (pkg *Package, err error) {
	fullPath := filepath.Join(env.dirSrc, importPath)
	dirGit := filepath.Join(fullPath, gitDir)

	_, err = os.Stat(fullPath)
	if err != nil {
		err = nil
		return
	}

	_, err = os.Stat(dirGit)
	if err != nil {
		if libio.IsDirEmpty(fullPath) {
			err = nil
		} else {
			err = fmt.Errorf(errDirNotEmpty, fullPath)
		}
		return
	}

	pkg, err = NewPackage(env.dirSrc, importPath, importPath)
	if err != nil {
		return
	}

	return
}

// GetPackageFromDB will return index and package in database.
// If no package found, it will return -1 and nil.
func (env *Env) GetPackageFromDB(importPath, remoteURL string) (int, *Package) {
	for x := 0; x < len(env.pkgs); x++ {
		if remoteURL == env.pkgs[x].RemoteURL {
			return x, env.pkgs[x]
		}
		if importPath == env.pkgs[x].ImportPath {
			return x, env.pkgs[x]
		}
	}

	for x := 0; x < len(env.pkgs); x++ {
		if strings.HasPrefix(importPath, env.pkgs[x].ImportPath) {
			return x, env.pkgs[x]
		}
	}

	return -1, nil
}

// GetUnused will get all non-registered packages from "src" directory,
// without including all excluded packages.
func (env *Env) GetUnused(srcPath string) (err error) {
	fis, err := ioutil.ReadDir(srcPath)
	if err != nil {
		err = fmt.Errorf("CleanPackages: %s", err)
		return
	}

	var nextScan []string

	for _, fi := range fis {
		// (0)
		if !fi.IsDir() {
			continue
		}

		dirName := fi.Name()
		fullPath := filepath.Join(srcPath, dirName)
		dirGit := filepath.Join(fullPath, gitDir)

		if IsIgnoredDir(dirName) {
			continue
		}

		_, err = os.Stat(dirGit)
		if err != nil {
			nextScan = append(nextScan, fullPath)
			err = nil
			continue
		}

		importPath := strings.TrimPrefix(fullPath, env.dirSrc+"/")

		if env.IsExcluded(importPath) {
			continue
		}

		_, pkg := env.GetPackageFromDB(importPath, "")
		if pkg != nil {
			continue
		}

		pkg, err = NewPackage(env.dirSrc, importPath, importPath)
		if err != nil {
			return
		}

		env.pkgsUnused = append(env.pkgsUnused, pkg)
	}

	for x := 0; x < len(nextScan); x++ {
		err = env.GetUnused(nextScan[x])
		if err != nil {
			return
		}
	}

	return nil
}

// IsExcluded will return true if import path is registered as one of excluded
// package; otherwise it will return false.
func (env *Env) IsExcluded(importPath string) bool {
	if len(importPath) == 0 {
		return true
	}
	for x := 0; x < len(env.pkgsExclude); x++ {
		if strings.Contains(importPath, env.pkgsExclude[x]) {
			return true
		}
	}
	return false
}

// Scan will gather all package information in user system to start `beku`-ing.
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

// scanStdPackages will traverse each directory in GOROOT `src` recursively
// until no subdirectory found. All path to subdirectories will be saved on
// Environment `pkgsStd`.
func (env *Env) scanStdPackages(srcPath string) error {
	fis, err := ioutil.ReadDir(srcPath)
	if err != nil {
		err = fmt.Errorf("scanStdPackages: %s", err)
		return err
	}

	for _, fi := range fis {
		// Skip non directory.
		if !fi.IsDir() {
			continue
		}

		dirName := fi.Name()
		fullPath := filepath.Join(srcPath, dirName)

		if IsIgnoredDir(dirName) {
			continue
		}

		stdPkg := strings.TrimPrefix(fullPath, env.dirGoRootSrc+"/")
		env.pkgsStd = append(env.pkgsStd, stdPkg)
	}

	return nil
}

// scanPackages will traverse each directory in `src` recursively until
// it's found VCS metadata, e.g. `.git` directory.
func (env *Env) scanPackages(srcPath string) (err error) {
	if debug.Value >= 1 {
		fmt.Println("[ENV] scanPackages >>>", srcPath)
	}

	fis, err := ioutil.ReadDir(srcPath)
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
			return
		}
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
		fullPath := filepath.Join(srcPath, dirName)
		dirGit := filepath.Join(fullPath, gitDir)

		if IsIgnoredDir(dirName) {
			continue
		}

		// Skip directory that contains ".git".
		_, err = os.Stat(dirGit)
		if err != nil {
			nextRoot = append(nextRoot, fullPath)
			err = nil
			continue
		}

		err = env.newPackage(fullPath)
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

	return nil
}

// newPackage will append the directory at path as a package only if
// its contain version information.
func (env *Env) newPackage(fullPath string) (err error) {
	pkgName := strings.TrimPrefix(fullPath, env.dirSrc+"/")

	if env.IsExcluded(pkgName) {
		return
	}

	pkg, err := NewPackage(env.dirSrc, pkgName, pkgName)
	if err != nil {
		return
	}

	if debug.Value >= 1 {
		fmt.Println("[ENV] newPackage >>>", pkg.ImportPath)
	}

	err = pkg.Scan()
	if err != nil {
		if err == ErrVersion {
			err = nil
			return
		}
		return fmt.Errorf("%s: %s", pkgName, err)
	}

	_, curPkg := env.GetPackageFromDB(pkg.ImportPath, pkg.RemoteURL)
	if curPkg == nil {
		env.pkgs = append(env.pkgs, pkg)
		env.countNew++

		if len(pkg.ImportPath) > env.fmtMaxPath {
			env.fmtMaxPath = len(pkg.ImportPath)
		}

		return nil
	}

	if curPkg.Version != pkg.Version {
		if (curPkg.isTag && pkg.isTag) || (!curPkg.isTag && !pkg.isTag) {
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

// addPackageMissing will add import path to list of missing package only if
// not exist yet.
func (env *Env) addPackageMissing(importPath string) {
	for x := 0; x < len(env.pkgsMissing); x++ {
		if importPath == env.pkgsMissing[x] {
			return
		}
	}

	env.pkgsMissing = append(env.pkgsMissing, importPath)
}

// Load will read saved dependencies from file.
func (env *Env) Load(file string) (err error) {
	if len(file) == 0 {
		env.dbFile = env.dbDefFile
	} else {
		env.dbFile = file
	}

	if debug.Value >= 1 {
		fmt.Println("[ENV] Load >>>", env.dbFile)
	}

	env.db, err = ini.Open(env.dbFile)
	if err != nil {
		return
	}

	env.loadBeku()
	env.loadPackages()

	return
}

func (env *Env) loadBeku() {
	for _, v := range env.db.Gets(sectionBeku, "", keyExclude) {
		env.addExclude(v)
	}
}

func (env *Env) loadPackages() {
	sections := env.db.Subs(sectionPackage)
	for _, sec := range sections {
		subName := sec.SubName()
		if env.IsExcluded(subName) {
			continue
		}

		pkg := &Package{
			ImportPath: subName,
			FullPath:   filepath.Join(env.dirSrc, subName),
			state:      packageStateLoad,
		}

		pkg.load(sec)

		env.addPackage(pkg)
	}
}

// Query the package database. If package is not empty, it will only show the
// information about that package.
func (env *Env) Query(pkgs []string) {
	format := fmt.Sprintf("%%-%ds  %%s\n", env.fmtMaxPath)

	for x := 0; x < len(env.pkgs); x++ {
		if len(pkgs) == 0 {
			fmt.Fprintf(defStdout, format, env.pkgs[x].ImportPath,
				env.pkgs[x].Version)
			continue
		}
		for y := 0; y < len(pkgs); y++ {
			if env.IsExcluded(pkgs[y]) {
				continue
			}

			if env.pkgs[x].ImportPath == pkgs[y] {
				fmt.Fprintf(defStdout, format,
					env.pkgs[x].ImportPath,
					env.pkgs[x].Version)
			}
		}
	}
}

// Rescan for new packages.
func (env *Env) Rescan(firstTime bool) (ok bool, err error) {
	err = env.Scan()
	if err != nil {
		return
	}

	format := fmt.Sprintf("%%-%ds  %%-12s  %%-12s\n", env.fmtMaxPath)

	if env.countUpdate > 0 {
		fmt.Println("[ENV] Rescan >>> New updates,")
		fmt.Printf(format+"\n", "ImportPath", "Old Version", "New Version")

		for _, pkg := range env.pkgs {
			if pkg.state&packageStateChange == 0 {
				continue
			}

			fmt.Printf(format, pkg.ImportPath, pkg.Version, pkg.VersionNext)
		}
	}
	if env.countNew > 0 {
		fmt.Println("[ENV] Rescan >>> New packages,")
		fmt.Printf(format+"\n", "ImportPath", "Old Version", "New Version")

		for _, pkg := range env.pkgs {
			if pkg.state&packageStateNew == 0 {
				continue
			}

			fmt.Printf(format, pkg.ImportPath, "-", pkg.Version)
		}
	}

	if env.countUpdate == 0 && env.countNew == 0 {
		if firstTime {
			env.dirty = true
		} else {
			fmt.Println("[ENV] Rescan >>> Database is in sync.")
		}
		return true, nil
	}

	fmt.Println()

	if !env.NoConfirm {
		ok = libio.ConfirmYesNo(os.Stdin, msgContinue, false)
		if !ok {
			return
		}
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
			env.updateMissing(pkg, true)
		}
	}
	env.dirty = true

	return true, nil
}

// Remove package from environment. If recursive is true, it will also remove
// their dependencies, as long as they are not required by other package.
func (env *Env) Remove(rmPkg string, recursive bool) (err error) {
	if env.IsExcluded(rmPkg) {
		fmt.Printf(errExcluded, rmPkg)
		return
	}

	_, pkg := env.GetPackageFromDB(rmPkg, "")
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

	fmt.Println("[ENV] Remove >>> The following package will be removed,")
	for _, importPath := range listRemoved {
		fmt.Println(" *", importPath)
	}

	if !env.NoConfirm {
		ok := libio.ConfirmYesNo(os.Stdin, msgContinue, false)
		if !ok {
			return
		}
	}

	for _, importPath := range listRemoved {
		err = env.removePackage(importPath)
		if err != nil {
			err = fmt.Errorf("Remove: %s", err)
			return
		}

		pkgImportPath := filepath.Join(env.dirPkg, importPath)

		if debug.Value >= 1 {
			fmt.Println("[ENV] Remove >>> Removing", pkgImportPath)
		}

		err = os.RemoveAll(pkgImportPath)
		if err != nil {
			err = fmt.Errorf("Remove: %s", err)
			return
		}

		_ = libio.RmdirEmptyAll(pkgImportPath)
	}

	return nil
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
		_, dep = env.GetPackageFromDB(pkg.Deps[x], "")
		if dep == nil {
			continue
		}

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

// removePackage from list environment (including source and installed archive
// or binary). This also remove in other packages "RequiredBy" if exist.
func (env *Env) removePackage(importPath string) (err error) {
	pkgIdx, pkg := env.GetPackageFromDB(importPath, "")
	if pkg == nil {
		return
	}

	err = pkg.Remove()
	if err != nil {
		return
	}

	env.removeRequiredBy(importPath)
	env.removePkgFromDBByIdx(pkgIdx)

	return
}

// removePkgFromDBByIdx remove package from database by package index in the
// list.
func (env *Env) removePkgFromDBByIdx(idx int) {
	if idx < 0 {
		return
	}

	lenpkgs := len(env.pkgs)

	copy(env.pkgs[idx:], env.pkgs[idx+1:])
	env.pkgs[lenpkgs-1] = nil
	env.pkgs = env.pkgs[:lenpkgs-1]

	env.dirty = true
}

// removeRequiredBy will remove import path in package required-by.
func (env *Env) removeRequiredBy(importPath string) {
	for x := 0; x < len(env.pkgs); x++ {
		ok := env.pkgs[x].RemoveRequiredBy(importPath)
		if ok {
			env.dirty = true
		}
	}
}

// Save the dependencies to `file` only if it's dirty flag is true.
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

	if debug.Value >= 1 {
		fmt.Println("[ENV] Save >>>", file)
	}

	dir := filepath.Dir(file)

	err = os.MkdirAll(dir, 0700)
	if err != nil {
		return
	}

	env.db = &ini.Ini{}

	env.saveBeku()
	env.savePackages()

	err = env.db.Save(file)
	if err != nil {
		return err
	}

	return nil
}

func (env *Env) saveBeku() {
	for _, exclude := range env.pkgsExclude {
		env.db.Add(sectionBeku, "", keyExclude, exclude)
	}
}

func (env *Env) savePackages() {
	for _, pkg := range env.pkgs {
		env.db.Set(sectionPackage, pkg.ImportPath, keyVCSMode, pkg.vcsMode)
		env.db.Set(sectionPackage, pkg.ImportPath, keyRemoteName, pkg.RemoteName)
		env.db.Set(sectionPackage, pkg.ImportPath, keyRemoteURL, pkg.RemoteURL)
		if len(pkg.RemoteBranch) > 0 {
			env.db.Set(sectionPackage, pkg.ImportPath, keyRemoteBranch, pkg.RemoteBranch)
		}
		env.db.Set(sectionPackage, pkg.ImportPath, keyVersion, pkg.Version)

		for _, dep := range pkg.Deps {
			env.db.Add(sectionPackage, pkg.ImportPath, keyDeps, dep)
		}
		for _, req := range pkg.RequiredBy {
			env.db.Add(sectionPackage, pkg.ImportPath, keyRequiredBy, req)
		}
		for _, mis := range pkg.DepsMissing {
			env.db.Add(sectionPackage, pkg.ImportPath, keyDepsMissing, mis)
		}
	}
}

// String return formatted output of the environment instance.
func (env *Env) String() string {
	var buf bytes.Buffer

	fmt.Fprintf(&buf, `
[ENV]
             Prefix: %s
            Dir bin: %s
            Dir pkg: %s
            Dir src: %s
       Dir root src: %s
  Standard Packages: %s
`, env.prefix, env.dirBin, env.dirPkg, env.dirSrc, env.dirGoRootSrc, env.pkgsStd)

	for x := 0; x < len(env.pkgs); x++ {
		fmt.Fprintf(&buf, "%s", env.pkgs[x].String())
	}

	for x := range env.pkgsMissing {
		fmt.Fprintln(&buf, "missing =", env.pkgsMissing[x])
	}

	return buf.String()
}

// install a package.
// If destination directory is not empty, it will ask for user confirmation to
// clean the directory first.
func (env *Env) install(pkg *Package) (ok bool, err error) {
	if !libio.IsDirEmpty(pkg.FullPath) {
		fmt.Printf("[ENV] install >>> Directory %s is not empty.\n", pkg.FullPath)
		if !env.NoConfirm {
			ok = libio.ConfirmYesNo(os.Stdin, msgCleanDir, false)
			if !ok {
				return
			}
		}
		_ = pkg.Remove()
	}

	err = pkg.Install()
	if err != nil {
		_ = pkg.Remove()
		return
	}

	ok = true

	return
}

func (env *Env) update(curPkg, newPkg *Package) (ok bool, err error) {
	err = curPkg.FetchLatestVersion()
	if err != nil {
		return
	}

	if len(newPkg.Version) == 0 {
		newPkg.Version = curPkg.VersionNext
		newPkg.isTag = curPkg.isTag
	}
	if len(newPkg.RemoteBranch) == 0 {
		err = newPkg.gitGetBranch()
		if err != nil {
			return
		}
	}

	if debug.Value >= 1 {
		fmt.Println("[ENV] update >>>", newPkg)
	}

	if curPkg.IsEqual(newPkg) || !newPkg.IsNewer(curPkg) {
		fmt.Println("[ENV] update >>> All package is up todate.")
		ok = true
		return
	}

	fmt.Printf("[ENV] update >>> Updating package from,\n%s\nto,\n%s\n",
		curPkg.String(), newPkg.String())

	if !env.NoConfirm {
		ok = libio.ConfirmYesNo(os.Stdin, msgUpdateView, false)
		if ok {
			err = curPkg.CompareVersion(newPkg)
			if err != nil {
				return
			}
		}
	}

	if !env.NoConfirm {
		ok = libio.ConfirmYesNo(os.Stdin, msgUpdateProceed, true)
		if !ok {
			return
		}
	}

	err = curPkg.Update(newPkg)
	if err != nil {
		return
	}

	env.dirty = true

	return true, nil
}

// installMissing will install all missing packages.
func (env *Env) installMissing(pkg *Package) (err error) {
	if env.noDeps {
		return
	}

	fmt.Printf("[ENV] installMissing %s >>> %s\n", pkg.ImportPath, pkg.DepsMissing)

	for _, misImportPath := range pkg.DepsMissing {
		_, misPkg := env.GetPackageFromDB(misImportPath, "")
		if misPkg != nil {
			continue
		}

		fmt.Printf("[ENV] installMissing %s >>> %s\n", pkg.ImportPath,
			misImportPath)

		err = env.Sync(misImportPath, misImportPath)
		if err != nil {
			fmt.Fprintf(defStderr, "[ENV] installMissing >>> %s\n", err)
			continue
		}
	}

	return
}

// updateMissing will remove missing package if it's already provided by new
// package. If "addAsDep" is true and the new package provide the missing one,
// then it will be added as one of package dependencies.
func (env *Env) updateMissing(newPkg *Package, addAsDep bool) {
	var updated bool

	if debug.Value >= 1 {
		fmt.Println("[ENV] updateMissing >>>", newPkg.ImportPath)
	}

	for x := 0; x < len(env.pkgs); x++ {
		updated = env.pkgs[x].UpdateMissingDep(newPkg, addAsDep)
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

// Sync will download and install a package including their dependencies. If
// the importPath is defined, it will be downloaded into that directory.
func (env *Env) Sync(pkgName, importPath string) (err error) {
	err = ErrPackageName

	if len(pkgName) == 0 {
		return
	}
	var (
		ok      bool
		version string
	)

	pkgName, version = parsePkgVersion(pkgName)
	if len(pkgName) == 0 {
		return
	}

	if len(importPath) == 0 {
		importPath = pkgName
	}

	if env.IsExcluded(pkgName) || env.IsExcluded(importPath) {
		fmt.Printf(errExcluded, pkgName)
		err = nil
		return
	}

	newPkg, err := NewPackage(env.dirSrc, pkgName, importPath)
	if err != nil {
		return
	}

	if len(version) > 0 {
		newPkg.Version = version
		newPkg.isTag = IsTagVersion(version)
	}

	// Check if package already installed, by checking out from database.
	_, curPkg := env.GetPackageFromDB(newPkg.ImportPath, newPkg.RemoteURL)
	if curPkg != nil {
		newPkg.RemoteURL = curPkg.RemoteURL
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

	err = env.postSync(curPkg)
	if err != nil {
		return err
	}

	return nil
}

// SyncMany packages at once.
func (env *Env) SyncMany(pkgs []string) (err error) {
	for _, pkg := range pkgs {
		err = env.Sync(pkg, "")
		if err != nil {
			return
		}
	}

	return
}

// SyncAll packages into latest version (tag or commit).
func (env *Env) SyncAll() (err error) {
	var (
		countUpdate int
		buf         bytes.Buffer
	)

	format := fmt.Sprintf("%%-%ds  %%-12s  %%-12s %%s\n", env.fmtMaxPath)
	fmt.Fprintf(&buf, "[ENV] SyncAll >>> The following packages will be updated,\n\n")
	fmt.Fprintf(&buf, format+"\n", "ImportPath", "Old Version",
		"New Version", "Compare URL")

	fmt.Println("[ENV] SyncAll >>> Updating all packages ...")

	for _, pkg := range env.pkgs {
		if libio.IsDirEmpty(pkg.FullPath) {
			fmt.Printf("[ENV] SyncAll %s >>> Installing\n",
				pkg.ImportPath)

			err = pkg.Install()
			if err != nil {
				_ = pkg.Remove()
				return
			}
		}

		fmt.Printf("[ENV] SyncAll %s >>> Current version is %s\n",
			pkg.ImportPath, pkg.Version)

		err = pkg.FetchLatestVersion()
		if err != nil {
			return
		}

		if pkg.Version >= pkg.VersionNext {
			fmt.Printf("[ENV] SyncAll %s >>> No update.\n\n",
				pkg.ImportPath)
			pkg.VersionNext = pkg.Version
			continue
		}

		fmt.Printf("[ENV] SyncAll %s >>> Latest version is %s\n\n",
			pkg.ImportPath, pkg.VersionNext)

		compareURL := GetCompareURL(pkg.RemoteURL, pkg.Version,
			pkg.VersionNext)

		fmt.Fprintf(&buf, format, pkg.ImportPath, pkg.Version,
			pkg.VersionNext, compareURL)

		countUpdate++
	}

	if countUpdate == 0 {
		fmt.Println("[ENV] SyncAll >>> All packages are up to date.")
		return
	}

	fmt.Println(buf.String())

	if !env.NoConfirm {
		ok := libio.ConfirmYesNo(os.Stdin, msgContinue, false)
		if !ok {
			return
		}
	}

	for _, pkg := range env.pkgs {
		err = pkg.CheckoutVersion(pkg.VersionNext)
		if err != nil {
			return
		}
		if pkg.Version != pkg.VersionNext {
			pkg.Version = pkg.VersionNext
			pkg.state = packageStateDirty
		}
	}

	env.dirty = true

	for _, pkg := range env.pkgs {
		if pkg.state&packageStateDirty > 0 {
			_ = env.postSync(pkg)
			// Ignore error. Go install may failed due to missing
			// dependencies.
		}
	}

	fmt.Println("[ENV] SyncAll >>> Update completed.")

	return nil
}

func (env *Env) postSync(pkg *Package) (err error) {
	fmt.Printf("\n[ENV] postSync %s\n", pkg.ImportPath)
	// Update missing packages.
	env.updateMissing(pkg, true)

	err = env.build(pkg)
	if err != nil {
		return
	}

	// Run `go install` only if no missing package.
	if len(pkg.DepsMissing) == 0 {
		_ = pkg.GoInstall(env.path)
	}

	fmt.Println("[ENV] postSync >>> Package installed:\n", pkg)

	return
}

func (env *Env) build(pkg *Package) (err error) {
	// Re-scan package dependencies.
	err = pkg.ScanDeps(env)
	if err != nil {
		return
	}

	// Install missing dependencies.
	return env.installMissing(pkg)
}
