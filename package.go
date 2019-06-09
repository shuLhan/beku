// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package beku

import (
	"bytes"
	"fmt"
	"go/build"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/tools/go/vcs"

	"github.com/shuLhan/share/lib/debug"
	"github.com/shuLhan/share/lib/git"
	"github.com/shuLhan/share/lib/ini"
	libio "github.com/shuLhan/share/lib/io"
)

const (
	VCSModeGit = "git"
)

//
// Package define Go package information: path to package, version, whether is
// tag or not, and VCS mode.
//
type Package struct {
	ImportPath   string
	FullPath     string
	RemoteName   string
	RemoteURL    string
	RemoteBranch string
	Version      string
	VersionNext  string
	DepsMissing  []string
	Deps         []string
	RequiredBy   []string
	vcsMode      string
	state        packageState
	isTag        bool
}

//
// NewPackage create a package set the package version, tag status, and
// dependencies.
//
func NewPackage(gopathSrc, name, importPath string) (pkg *Package, err error) {
	repoRoot, err := vcs.RepoRootForImportPath(name, debug.Value >= 1)
	if err != nil {
		fmt.Fprintf(defStderr, "[PKG] NewPackage >>> error: %s\n", err.Error())
		return
	}

	if repoRoot.VCS.Cmd != VCSModeGit {
		err = fmt.Errorf(errVCS, repoRoot.VCS.Cmd)
		return nil, err
	}

	pkg = &Package{
		ImportPath: importPath,
		FullPath:   filepath.Join(gopathSrc, importPath),
		RemoteName: gitDefRemoteName,
		RemoteURL:  repoRoot.Repo,
		vcsMode:    repoRoot.VCS.Cmd,
		state:      packageStateNew,
	}

	if debug.Value >= 2 {
		fmt.Printf("[PKG] NewPackage >>> %+v\n", pkg)
	}

	return
}

//
// CheckoutVersion will set the package version to new version.
//
func (pkg *Package) CheckoutVersion(newVersion string) (err error) {
	if pkg.vcsMode == VCSModeGit {
		if len(pkg.RemoteBranch) == 0 {
			err = pkg.gitGetBranch()
			if err != nil {
				return
			}
		}
		err = git.CheckoutRevision(pkg.FullPath, pkg.RemoteName, pkg.RemoteBranch, newVersion)
	}

	return

}

//
// CompareVersion will compare package version using current package as base.
//
func (pkg *Package) CompareVersion(newPkg *Package) (err error) {
	if pkg.vcsMode == VCSModeGit {
		err = git.LogRevisions(pkg.FullPath, pkg.Version, newPkg.Version)
	}

	return
}

//
// FetchLatestVersion will try to update the package and get the latest
// version (tag or commit).
//
func (pkg *Package) FetchLatestVersion() (err error) {
	if pkg.vcsMode == VCSModeGit {
		err = git.FetchAll(pkg.FullPath)
		if err != nil {
			return
		}
		if pkg.isTag {
			pkg.VersionNext, err = git.LatestTag(pkg.FullPath)
		} else {
			pkg.VersionNext, err = git.LatestCommit(pkg.FullPath, "")
		}
	}

	return
}

func (pkg *Package) Freeze() (err error) {
	if pkg.vcsMode == VCSModeGit {
		err = pkg.gitFreeze()
	}
	return
}

//
// GoClean will remove the package binaries and archives.
//
func (pkg *Package) GoClean() (err error) {
	_, err = os.Stat(pkg.FullPath)
	if err != nil {
		err = nil
		return
	}

	cmd := exec.Command("go", "clean", "-i", "-cache", "-testcache", "./...")
	cmd.Dir = pkg.FullPath
	cmd.Env = append(cmd.Env, "GO111MODULE=off")
	cmd.Env = append(cmd.Env, "GOPATH="+build.Default.GOPATH)
	cmd.Stdout = defStdout
	cmd.Stderr = defStderr

	if debug.Value >= 1 {
		fmt.Printf("[PKG] GoClean %s >>> %s %s\n", pkg.ImportPath, cmd.Dir, cmd.Args)
	}

	err = cmd.Run()
	if err != nil {
		err = fmt.Errorf("GoClean: %s", err)
		return
	}

	return
}

//
// Install a package. Clone package "src" directory, set to the latest tag if
// exist or to the latest commit, and scan dependencies.
//
func (pkg *Package) Install() (err error) {
	if pkg.vcsMode == VCSModeGit {
		err = pkg.gitInstall()
	}

	return
}

//
// IsEqual will return true if current package have the same import path,
// remote name, remote URL, and version with other package; otherwise it will
// return false.
//
func (pkg *Package) IsEqual(other *Package) bool {
	if other == nil {
		return false
	}
	if pkg.ImportPath != other.ImportPath {
		return false
	}
	if pkg.RemoteName != other.RemoteName {
		return false
	}
	if pkg.RemoteURL != other.RemoteURL {
		return false
	}
	if pkg.RemoteBranch != other.RemoteBranch {
		return false
	}
	if pkg.Version != other.Version {
		return false
	}

	return true
}

//
// IsNewer will return true if current package is using tag and have newer
// version that other package.  If current package is not using tag, it's
// always return true.
//
func (pkg *Package) IsNewer(older *Package) bool {
	if !pkg.isTag {
		return true
	}
	return pkg.Version >= older.Version
}

//
// Remove package installed binaries, archives, and source.
//
func (pkg *Package) Remove() (err error) {
	err = pkg.GoClean()
	if err != nil {
		err = fmt.Errorf("Remove: %s", err)
		return
	}

	if debug.Value >= 1 {
		fmt.Printf("[PKG] Remove %s >>> %s\n", pkg.ImportPath,
			pkg.FullPath)
	}

	err = os.RemoveAll(pkg.FullPath)
	if err != nil {
		err = fmt.Errorf("Remove: %s", err)
		return
	}

	_ = libio.RmdirEmptyAll(pkg.FullPath)

	return
}

//
// RemoveRequiredBy will remove package importPath from list of required-by.
// It will return true if importPath is removed from list, otherwise it will
// return false.
//
func (pkg *Package) RemoveRequiredBy(importPath string) (ok bool) {
	var requiredBy []string

	for x := 0; x < len(pkg.RequiredBy); x++ {
		if pkg.RequiredBy[x] == importPath {
			ok = true
			continue
		}
		requiredBy = append(requiredBy, pkg.RequiredBy[x])
	}
	if ok {
		pkg.RequiredBy = requiredBy
	}
	return
}

//
// Scan will set the package version, `isTag` status, and remote URL using
// metadata in package repository.
//
func (pkg *Package) Scan() (err error) {
	if pkg.vcsMode == VCSModeGit {
		err = pkg.gitScan()
	}

	if err != nil {
		return
	}

	pkg.isTag = IsTagVersion(pkg.Version)

	return
}

//
// ScanDeps will scan package dependencies, removing standard packages, keep
// only external dependencies.
//
func (pkg *Package) ScanDeps(env *Env) (err error) {
	if debug.Value >= 1 {
		fmt.Println("[PKG] ScanDeps", pkg.ImportPath)
	}

	imports, err := pkg.GetRecursiveImports(env)
	if err != nil {
		return
	}

	for x := 0; x < len(imports); x++ {
		pkg.addDep(env, imports[x])
	}

	return
}

//
// GetRecursiveImports will get all import path recursively using `go list`
// and return it as slice of string without any duplication.
//
func (pkg *Package) GetRecursiveImports(env *Env) (
	imports []string, err error,
) {
	cmd := exec.Command("go", "list", "-e", "-f", `{{ join .Imports "\n"}}`, "./...")
	cmd.Dir = pkg.FullPath
	cmd.Stderr = defStderr

	if debug.Value >= 1 {
		fmt.Printf("= GetRecursiveImports %s %s\n", cmd.Dir, cmd.Args)
	}

	out, err := cmd.Output()
	if err != nil {
		err = fmt.Errorf("GetRecursiveImports: %s", err)
		return
	}

	var found bool
	importsDup := strings.Split(string(out), "\n")

	for x := 0; x < len(importsDup); x++ {
		if env.vendor {
			importsDup[x] = strings.TrimPrefix(importsDup[x], env.prefix+"/")
		}

		found = false
		for y := 0; y < len(imports); y++ {
			if importsDup[x] == imports[y] {
				found = true
				break
			}
		}
		if found {
			continue
		}
		imports = append(imports, importsDup[x])
	}

	sort.Strings(imports)

	return imports, nil
}

//
// addDep will add `importPath` to package dependencies only if it's,
//
// (0) not empty
// (1) not self importing
// (2) not vendor
// (3) not pseudo-package ("C")
// (4) not standard packages
//
// (5) If all above filter passed, then it will do package normalization to
// check their dependencies with existing package in environment.
//
// (5.1) if match found, link the package deps to existing package instance.
// (5.2) If no match found, add to list of missing `depsMissing`
//
// It will return true if import path is added as dependencies or as missing
// one; otherwise it will return false.
//
func (pkg *Package) addDep(env *Env, importPath string) bool {
	// (0)
	if len(importPath) == 0 {
		return false
	}

	// (1)
	if strings.HasPrefix(importPath, pkg.ImportPath) {
		if debug.Value >= 2 {
			fmt.Printf("[PKG] addDep %s >>> skip self import: %s\n",
				pkg.ImportPath, importPath)
		}
		return false
	}

	// (2)
	pkgs := strings.Split(importPath, sepImport)
	if pkgs[0] == dirVendor {
		return false
	}

	// (3)
	if importPath == "C" {
		return false
	}

	// (4)
	for x := 0; x < len(env.pkgsStd); x++ {
		if pkgs[0] != env.pkgsStd[x] {
			continue
		}
		if debug.Value >= 2 {
			fmt.Printf("[PKG] addDep %s >>> skip std: %s\n",
				pkg.ImportPath, importPath)
		}
		return false
	}

	// (5)
	for x := 0; x < len(env.pkgs); x++ {
		if !strings.HasPrefix(importPath, env.pkgs[x].ImportPath) {
			continue
		}

		// (5.1)
		pkg.pushDep(env.pkgs[x].ImportPath)
		env.pkgs[x].pushRequiredBy(pkg.ImportPath)
		return true
	}

	if debug.Value >= 2 {
		fmt.Printf("[PKG] addDep %s >>> missing: %s\n",
			pkg.ImportPath, importPath)
	}

	// (5.2)
	pkg.pushMissing(importPath)
	env.addPackageMissing(importPath)

	return true
}

//
// load package metadata from database (INI Section).
//
func (pkg *Package) load(sec *ini.Section) {
	for _, v := range sec.Vars {
		switch v.KeyLower {
		case keyVCSMode:
			switch v.Value {
			case VCSModeGit:
				pkg.vcsMode = VCSModeGit
			default:
				pkg.vcsMode = VCSModeGit
			}
		case keyRemoteName:
			pkg.RemoteName = v.Value
		case keyRemoteURL:
			pkg.RemoteURL = v.Value
		case keyRemoteBranch:
			pkg.RemoteBranch = v.Value
		case keyVersion:
			pkg.Version = v.Value
			pkg.isTag = IsTagVersion(pkg.Version)
		case keyDeps:
			pkg.pushDep(v.Value)
		case keyDepsMissing:
			pkg.pushMissing(v.Value)
		case keyRequiredBy:
			pkg.pushRequiredBy(v.Value)
		}
	}
}

//
// GoInstall a package recursively ("./...").
//
// Set PATH to let go install that require gcc work when invoked from
// non-interactive shell (e.g. buildbot).
//
func (pkg *Package) GoInstall(envPath string) (err error) {
	cmd := exec.Command("go", "install")
	if debug.Value >= 2 {
		cmd.Args = append(cmd.Args, "-v")
	}
	cmd.Args = append(cmd.Args, "./...")

	envGOCACHE := os.Getenv("GOCACHE")
	envHOME := os.Getenv("HOME")

	cmd.Env = append(cmd.Env, "GO111MODULE=off")
	cmd.Env = append(cmd.Env, "GOPATH="+build.Default.GOPATH)
	cmd.Env = append(cmd.Env, "PATH="+envPath)
	cmd.Env = append(cmd.Env, "GOCACHE="+envGOCACHE)
	cmd.Env = append(cmd.Env, "HOME="+envHOME)
	cmd.Dir = pkg.FullPath
	cmd.Stdout = defStdout
	cmd.Stderr = defStderr

	if debug.Value == 0 {
		fmt.Printf("= GoInstall %s\n", cmd.Dir)
	} else {
		fmt.Printf("= GoInstall %s\n%s\n%s\n", cmd.Dir, cmd.Env, cmd.Args)
	}

	err = cmd.Run()

	return
}

//
// String return formatted output of the package instance.
//
func (pkg *Package) String() string {
	var buf bytes.Buffer

	fmt.Fprintf(&buf, `
[package "%s"]
     FullPath = %s
          VCS = %s
   RemoteName = %s
    RemoteURL = %s
 RemoteBranch = %s
      Version = %s
  VersionNext = %s
        IsTag = %v
         Deps = %v
   RequiredBy = %v
  DepsMissing = %v
`, pkg.ImportPath, pkg.FullPath, pkg.vcsMode, pkg.RemoteName,
		pkg.RemoteURL, pkg.RemoteBranch, pkg.Version, pkg.VersionNext,
		pkg.isTag, pkg.Deps, pkg.RequiredBy, pkg.DepsMissing)

	return buf.String()
}

//
// Update the current package to the new package. The new package may contain
// new remote or new version.
//
func (pkg *Package) Update(newPkg *Package) (err error) {
	if pkg.ImportPath != newPkg.ImportPath {
		err = os.Rename(pkg.FullPath, newPkg.FullPath)
		if err != nil {
			err = fmt.Errorf("Update: %s", err)
			return
		}

		pkg.ImportPath = newPkg.ImportPath
		pkg.FullPath = newPkg.FullPath
	}

	if len(pkg.RemoteBranch) == 0 {
		err = pkg.gitGetBranch()
		if err != nil {
			return
		}
	}

	if pkg.vcsMode == VCSModeGit {
		err = pkg.gitUpdate(newPkg)
	}
	if err != nil {
		return
	}

	pkg.RemoteName = newPkg.RemoteName
	pkg.RemoteURL = newPkg.RemoteURL
	pkg.Version = newPkg.Version
	pkg.isTag = IsTagVersion(newPkg.Version)

	return nil
}

//
// UpdateMissingDep will remove missing package if it's already provided by
// new package import-path.
//
// If "addAsDep" is true, it will also,
// (1) add new package as one of package dependencies of current package, and
// (2) add current package as required by new package.
//
// It will return true if new package solve the missing deps on current
// package, otherwise it will return false.
//
func (pkg *Package) UpdateMissingDep(newPkg *Package, addAsDep bool) (found bool) {
	var missing []string
	for x := 0; x < len(pkg.DepsMissing); x++ {
		if !strings.HasPrefix(pkg.DepsMissing[x], newPkg.ImportPath) {
			missing = append(missing, pkg.DepsMissing[x])
			continue
		}

		if addAsDep {
			pkg.pushDep(newPkg.ImportPath)
			newPkg.pushRequiredBy(pkg.ImportPath)
		}
		found = true
	}

	if found {
		pkg.DepsMissing = missing
		pkg.state = packageStateDirty
	}

	return
}

//
// pushDep will append import path into list of dependencies only if it's not
// exist. If import path exist it will return false.
//
func (pkg *Package) pushDep(importPath string) bool {
	for x := 0; x < len(pkg.Deps); x++ {
		if importPath == pkg.Deps[x] {
			return false
		}
	}

	pkg.Deps = append(pkg.Deps, importPath)

	if debug.Value >= 2 {
		fmt.Printf("[PKG] pushDep %s >>> %s\n", pkg.ImportPath,
			importPath)
	}

	return true
}

//
// pushMissing import path only if not exist yet.
//
func (pkg *Package) pushMissing(importPath string) bool {
	for x := 0; x < len(pkg.DepsMissing); x++ {
		if pkg.DepsMissing[x] == importPath {
			return false
		}
	}

	pkg.DepsMissing = append(pkg.DepsMissing, importPath)

	return true
}

//
// pushRequiredBy add the import path as required by current package.
//
func (pkg *Package) pushRequiredBy(importPath string) bool {
	for x := 0; x < len(pkg.RequiredBy); x++ {
		if importPath == pkg.RequiredBy[x] {
			return false
		}
	}

	pkg.RequiredBy = append(pkg.RequiredBy, importPath)

	return true
}
