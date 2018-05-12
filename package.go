// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package beku

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"sort"
	"strings"

	"github.com/shuLhan/share/lib/ini"
)

const (
	tagPrefix  = 'v'
	versionSep = '.'
	importSep  = "/"

	dbgSkipSelf = "skip self dep"
	dbgSkipStd  = "skip std dep"
	dbgMissDep  = "missing dep"
	dbgLinkDep  = "linking dep"

	gitCfgRemote    = "remote"
	gitCfgRemoteURL = "url"
	gitDefRemote    = "origin"
)

var (
	// ErrVersion define an error when package have VCS metadata (e.g.
	// `.git` directory, but did not found any tag or commit.
	ErrVersion = errors.New("No tag or commit found")

	// ErrRemote define an error when package does not have remote URL.
	ErrRemote = errors.New("No remote URL found")
)

var (
	keyVCSMode     = "vcs"
	keyRemoteName  = "remote-name"
	keyRemoteURL   = "remote-url"
	keyVersion     = "version"
	keyDeps        = "deps"
	keyDepsMissing = "missing"
	keyRequiredBy  = "required-by"
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
	DepsMissing []string
	Deps        []string
	RequiredBy  []string
	vcs         VCSMode
	isTag       bool
}

//
// NewPackage will set the package version, tag status, and dependencies.
//
func NewPackage(env *Env, importPath, fullPath string, vcs VCSMode) (
	pkg *Package, err error,
) {
	pkg = &Package{
		ImportPath: importPath,
		FullPath:   fullPath,
		vcs:        vcs,
	}

	err = pkg.Scan(env)

	return
}

//
// Scan will set the package version, `isTag` status, and remote.
//
func (pkg *Package) Scan(env *Env) (err error) {
	if env.Debug >= DebugL2 {
		log.Println("Scanning package:", pkg.ImportPath)
	}

	switch pkg.vcs {
	case VCSModeGit:
		err = pkg.gitScan()
	}

	if err != nil {
		return
	}

	pkg.setIsTag()

	return
}

func (pkg *Package) gitScan() (err error) {
	err = pkg.gitScanVersion()
	if err != nil {
		return
	}

	err = pkg.gitScanRemote()

	return
}

//
// gitScanVersion will try to,
// (1) get latest tag from repository first, or
// (2) if it's fail it will get the commit hash at HEAD.
//
// nolint: gas
func (pkg *Package) gitScanVersion() (err error) {
	// (1)
	cmd := exec.Command("git", "-C", pkg.FullPath, "describe", "--tags",
		"--exact-match")

	ver, err := cmd.Output()
	if err == nil {
		goto out
	}

	// (2)
	cmd = exec.Command("git", "-C", pkg.FullPath, "rev-parse", "--short",
		"HEAD")

	ver, err = cmd.Output()
	if err != nil {
		return ErrVersion
	}
out:
	pkg.Version = string(bytes.TrimSpace(ver))

	return
}

func (pkg *Package) gitScanRemote() (err error) {
	gitConfig := pkg.FullPath + "/" + gitDir + "/config"

	gitIni, err := ini.Open(gitConfig)
	if err != nil {
		return
	}

	url, ok := gitIni.Get(gitCfgRemote, gitDefRemote, gitCfgRemoteURL)
	if !ok {
		return ErrRemote
	}

	pkg.RemoteName = gitDefRemote
	pkg.RemoteURL = url

	return
}

//
// setIsTag will set isTag to true if `Version` prefixed with `v` or contains
// dot `.` character.
//
func (pkg *Package) setIsTag() {
	if len(pkg.Version) == 0 {
		pkg.isTag = false
		return
	}
	if pkg.Version[0] == tagPrefix {
		pkg.isTag = true
		return
	}
	if strings.IndexByte(pkg.Version, versionSep) > 0 {
		pkg.isTag = true
	}
}

//
// ScanDeps will scan package dependencies, removing standard packages, keep
// only external dependencies.
//
func (pkg *Package) ScanDeps(env *Env) (err error) {
	imports, err := pkg.GetRecursiveImports(env)
	if err != nil {
		return
	}

	if env.Debug >= DebugL2 && len(imports) > 0 {
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
// nolint: gas
func (pkg *Package) GetRecursiveImports(env *Env) (
	imports []string, err error,
) {
	cmd := exec.Command("go", "list", "-f", `{{ join .Deps "\n"}}`,
		"./...")
	cmd.Dir = pkg.FullPath

	out, err := cmd.Output()
	if err != nil {
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
// It will return true if import path is added as link or missing; otherwise
// it will return false.
//
func (pkg *Package) addDep(env *Env, importPath string) bool {
	// (0)
	if len(importPath) == 0 {
		return false
	}

	// (1)
	if strings.HasPrefix(importPath, pkg.ImportPath) {
		if env.Debug >= DebugL2 {
			log.Printf("%15s >>> %s\n", dbgSkipSelf, importPath)
		}
		return false
	}

	// (2)
	pkgs := strings.Split(importPath, importSep)
	if pkgs[0] == vendorDir {
		return false
	}

	// (3)
	for x := 0; x < len(env.pkgsStd); x++ {
		if pkgs[0] != env.pkgsStd[x] {
			continue
		}
		if env.Debug >= DebugL2 {
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
		pkg.linkDep(env, env.pkgs[x])
		env.pkgs[x].linkRequiredBy(env, pkg)
		return true
	}

	if env.Debug >= DebugL2 {
		log.Printf("%15s >>> %s\n", dbgMissDep, importPath)
	}

	pkg.DepsMissing = append(pkg.DepsMissing, importPath)
	env.addPackageMissing(importPath)

	return true
}

//
// linkDep will link the package `dep` only if it's not exist yet.
//
func (pkg *Package) linkDep(env *Env, dep *Package) bool {
	for x := 0; x < len(pkg.Deps); x++ {
		if dep.ImportPath == pkg.Deps[x] {
			return false
		}
	}

	pkg.Deps = append(pkg.Deps, dep.ImportPath)

	if env.Debug >= DebugL2 {
		log.Printf("%15s >>> %s\n", dbgLinkDep, dep.ImportPath)
	}

	return true
}

func (pkg *Package) linkRequiredBy(env *Env, parentPkg *Package) bool {
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
			pkg.setIsTag()
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
