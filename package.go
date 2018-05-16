// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package beku

import (
	"bytes"
	"fmt"
	"go/build"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/shuLhan/share/lib/ini"
)

//
// Package define Go package information: path to package, version, whether is
// tag or not, and VCS mode.
//
type Package struct {
	ImportPath  string
	FullPath    string
	RemoteName  string
	RemoteURL   string
	Version     string
	VersionNext string
	DepsMissing []string
	Deps        []string
	RequiredBy  []string
	vcs         VCSMode
	isTag       bool
}

//
// NewPackage create a package set the package version, tag status, and dependencies.
//
func NewPackage(pkgName, importPath string, vcsMode VCSMode) (
	pkg *Package,
) {
	pkg = &Package{
		ImportPath: importPath,
		RemoteURL:  "https://" + pkgName,
		FullPath:   build.Default.GOPATH + "/" + dirSrc + "/" + importPath,
		vcs:        vcsMode,
	}

	switch vcsMode {
	case VCSModeGit:
		pkg.RemoteName = gitDefRemoteName
	}

	return
}

//
// CompareVersion will compare package version using current package as base.
//
func (pkg *Package) CompareVersion(newPkg *Package) (err error) {
	switch pkg.vcs {
	case VCSModeGit:
		err = pkg.gitCompareVersion(newPkg)
	}

	return
}

//
// Fetch will try to update the package and get the latest version (tag or
// commit).
//
func (pkg *Package) Fetch() (err error) {
	switch pkg.vcs {
	case VCSModeGit:
		err = pkg.gitFetch()
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
	if pkg.Version != other.Version {
		return false
	}

	return true
}

//
// Scan will set the package version, `isTag` status, and remote URL using
// metadata in package repository.
//
func (pkg *Package) Scan() (err error) {
	switch pkg.vcs {
	case VCSModeGit:
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
	imports, err := pkg.GetRecursiveImports()
	if err != nil {
		return
	}

	if Debug >= DebugL2 && len(imports) > 0 {
		log.Println("   imports recursive:", imports)
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
func (pkg *Package) GetRecursiveImports() (
	imports []string, err error,
) {
	cmd := exec.Command("go", "list", "-f", `{{ join .Deps "\n"}}`,
		"./...")
	cmd.Dir = pkg.FullPath

	out, err := cmd.Output()
	if err != nil {
		err = fmt.Errorf("GetRecursiveImports: %s", err)
		return
	}

	var found bool
	importsDup := strings.Split(string(out), "\n")

	for x := 0; x < len(importsDup); x++ {
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

	return
}

//
// addDep will add `importPath` to package dependencies only if it's,
//
// (0) not empty
// (1) not self importing
// (2) not vendor
// (3) not standard packages.
//
// (4) If all above filter passed, then it will do package normalization to
// check their dependencies with existing package in `$GOPATH/src`.
//
// (4.1) if match found, link the package deps to existing package instance.
// (4.2) If no match found, add to `depsMissing` as string
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
		if Debug >= DebugL2 {
			log.Printf("%15s >>> %s\n", dbgSkipSelf, importPath)
		}
		return false
	}

	// (2)
	pkgs := strings.Split(importPath, sepImport)
	if pkgs[0] == dirVendor {
		return false
	}

	// (3)
	for x := 0; x < len(env.pkgsStd); x++ {
		if pkgs[0] != env.pkgsStd[x] {
			continue
		}
		if Debug >= DebugL2 {
			log.Printf("%15s >>> %s\n", dbgSkipStd, importPath)
		}
		return false
	}

	// (4)
	for x := 0; x < len(env.pkgs); x++ {
		if !strings.HasPrefix(importPath, env.pkgs[x].ImportPath) {
			continue
		}

		// (4.1)
		pkg.linkDep(env.pkgs[x])
		env.pkgs[x].linkRequiredBy(pkg)
		return true
	}

	if Debug >= DebugL2 {
		log.Printf("%15s >>> %s\n", dbgMissDep, importPath)
	}

	pkg.addMissing(importPath)
	env.addPackageMissing(importPath)

	return true
}

//
// linkDep will link the package `dep` only if it's not exist yet.
//
func (pkg *Package) linkDep(dep *Package) bool {
	for x := 0; x < len(pkg.Deps); x++ {
		if dep.ImportPath == pkg.Deps[x] {
			return false
		}
	}

	pkg.Deps = append(pkg.Deps, dep.ImportPath)

	if Debug >= DebugL2 {
		log.Printf("%15s >>> %s\n", dbgLinkDep, dep.ImportPath)
	}

	return true
}

func (pkg *Package) linkRequiredBy(parentPkg *Package) bool {
	for x := 0; x < len(pkg.RequiredBy); x++ {
		if parentPkg.ImportPath == pkg.RequiredBy[x] {
			return false
		}
	}

	pkg.RequiredBy = append(pkg.RequiredBy, parentPkg.ImportPath)

	return true
}

func (pkg *Package) load(sec *ini.Section) {
	for _, v := range sec.Vars {
		switch v.KeyLower {
		case keyVCSMode:
			switch v.Value {
			case valVCSModeGit:
				pkg.vcs = VCSModeGit
			}
		case keyRemoteName:
			pkg.RemoteName = v.Value
		case keyRemoteURL:
			pkg.RemoteURL = v.Value
		case keyVersion:
			pkg.Version = v.Value
			pkg.isTag = IsTagVersion(pkg.Version)
		case keyDeps:
			pkg.Deps = append(pkg.Deps, v.Value)
		case keyDepsMissing:
			pkg.DepsMissing = append(pkg.DepsMissing, v.Value)
		case keyRequiredBy:
			pkg.RequiredBy = append(pkg.RequiredBy, v.Value)
		}
	}
}

//
// RunGoInstall will run go install recursively on package directory.
//
func (pkg *Package) RunGoInstall(isVerbose bool) (err error) {
	log.Println(">>> Running go install ...")

	var opts []string

	if isVerbose {
		opts = append(opts, "-v")
	}

	opts = append(opts, "./...")

	cmd := exec.Command("go", "install")
	cmd.Args = append(cmd.Args, opts...)
	cmd.Dir = pkg.FullPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

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
          VCS = %d
   RemoteName = %s
    RemoteURL = %s
      Version = %s
        IsTag = %v
         Deps = %v
   RequiredBy = %v
  DepsMissing = %v
`, pkg.ImportPath, pkg.vcs, pkg.RemoteName, pkg.RemoteURL, pkg.Version,
		pkg.isTag, pkg.Deps, pkg.RequiredBy, pkg.DepsMissing)

	return buf.String()
}

//
// Update the current package to the new package remote or version.
//
func (pkg *Package) Update(newPkg *Package) (err error) {
	switch pkg.vcs {
	case VCSModeGit:
		err = pkg.gitUpdate(newPkg)
	}
	if err != nil {
		return
	}

	return
}

//
// UpdateMissingDeps will remove missing package if it's already provided by
// new package, and add it as one of package dependencies.
//
func (pkg *Package) UpdateMissingDeps(newPkg *Package) {
	var missing []string
	for x := 0; x < len(pkg.DepsMissing); x++ {
		if !strings.HasPrefix(pkg.DepsMissing[x], newPkg.ImportPath) {
			missing = append(missing, pkg.DepsMissing[x])
			continue
		}

		pkg.linkDep(newPkg)
		newPkg.linkRequiredBy(pkg)
	}

	pkg.DepsMissing = missing
}

func (pkg *Package) addMissing(importPath string) {
	for x := 0; x < len(pkg.DepsMissing); x++ {
		if pkg.DepsMissing[x] == importPath {
			return
		}
	}

	pkg.DepsMissing = append(pkg.DepsMissing, importPath)
}
