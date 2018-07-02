// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package beku

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/shuLhan/share/lib/ini"
)

//
// gitCheckoutVersion will set the HEAD to version stated in package.
//
// We are not using "git stash", because they introduce many problems when
// rebuilding the package after update.
//
func (pkg *Package) gitCheckoutVersion(version string) (err error) {
	if len(version) == 0 {
		fmt.Printf("[PKG] gitCheckoutVersion %s >>> empty version\n",
			pkg.ImportPath)
		return
	}

	cmd := exec.Command("git", "clean", "-qdff")
	cmd.Dir = pkg.FullPath
	cmd.Stdout = defStdout
	cmd.Stderr = defStderr

	fmt.Printf("[PKG] gitCheckoutVersion %s >>> %s %s\n", pkg.ImportPath,
		cmd.Dir, cmd.Args)

	_ = cmd.Run()

	cmd = exec.Command("git", "checkout", "-t", "origin/master", "-B", "master")
	cmd.Dir = pkg.FullPath
	cmd.Stdout = defStdout
	cmd.Stderr = defStderr

	fmt.Printf("[PKG] gitCheckoutVersion %s >>> %s %s\n", pkg.ImportPath,
		cmd.Dir, cmd.Args)

	_ = cmd.Run()

	//nolint:gas
	cmd = exec.Command("git", "reset", "--hard", version)
	cmd.Dir = pkg.FullPath
	cmd.Stdout = defStdout
	cmd.Stderr = defStderr

	fmt.Printf("[PKG] gitCheckoutVersion %s >>> %s %s\n", pkg.ImportPath,
		cmd.Dir, cmd.Args)

	err = cmd.Run()
	if err != nil {
		err = fmt.Errorf("gitCheckoutVersion %s: %s", pkg.FullPath, err)
		return
	}

	return
}

//
// gitClone the package into "{prefix}/src/{ImportPath}".
// If destination directory is not empty it will return an error.
//
func (pkg *Package) gitClone() (err error) {
	err = os.MkdirAll(pkg.FullPath, 0700)
	if err != nil {
		err = fmt.Errorf("gitClone: %s", err)
		return
	}

	empty := IsDirEmpty(pkg.FullPath)
	if !empty {
		err = fmt.Errorf("gitClone: "+errDirNotEmpty, pkg.FullPath)
		return
	}

	//nolint:gas
	cmd := exec.Command("git", "clone", pkg.RemoteURL, ".")
	cmd.Dir = pkg.FullPath
	cmd.Stdout = defStdout
	cmd.Stderr = defStderr

	if Debug >= DebugL1 {
		fmt.Printf("[PKG] gitClone %s >>> %s %s\n", pkg.ImportPath,
			cmd.Dir, cmd.Args)
	}

	err = cmd.Run()
	if err != nil {
		err = fmt.Errorf("gitClone: %s", err)
		return
	}

	return
}

//
// gitCompareVersion compare the version of current package with new package.
//
func (pkg *Package) gitCompareVersion(newPkg *Package) (err error) {
	//nolint:gas
	cmd := exec.Command("git", "log", "--oneline", pkg.Version+"..."+newPkg.Version)
	cmd.Dir = pkg.FullPath
	cmd.Stdout = defStdout
	cmd.Stderr = defStderr

	if Debug >= DebugL1 {
		fmt.Printf("[PKG] gitCompareVersion %s >>> %s %s\n",
			pkg.ImportPath, cmd.Dir, cmd.Args)
	}

	err = cmd.Run()
	if err != nil {
		err = fmt.Errorf("gitCompareVersion: %s", err)
		return
	}

	return
}

//
// gitFetch will fetch the latest commit from remote. On success, it will set
// the package next version to latest tag (if current package is using tag) or
// to latest commit otherwise.
//
func (pkg *Package) gitFetch() (err error) {
	//nolint:gas
	cmd := exec.Command("git", "fetch", "--all")
	cmd.Dir = pkg.FullPath
	cmd.Stdout = defStdout
	cmd.Stderr = defStderr

	if Debug >= DebugL1 {
		fmt.Printf("[PKG] gitFetch %s >>> %s %s\n", pkg.ImportPath,
			cmd.Dir, cmd.Args)
	}

	err = cmd.Run()
	if err != nil {
		err = fmt.Errorf("gitFetch: %s", err)
		return
	}

	if pkg.isTag {
		pkg.VersionNext, err = pkg.gitGetTagLatest()
	} else {
		ref := filepath.Join(pkg.RemoteName, gitDefBranch)
		pkg.VersionNext, err = pkg.gitGetCommit(ref)
	}

	return
}

//
// gitGetCommit will try to get the latest commit hash from "ref"
// (origin/master).
//
func (pkg *Package) gitGetCommit(ref string) (commit string, err error) {
	//nolint:gas
	cmd := exec.Command("git", "rev-parse", "--short", ref)
	cmd.Dir = pkg.FullPath

	if Debug >= DebugL1 {
		fmt.Printf("[PKG] gitGetCommit %s >>> %s %s\n",
			pkg.ImportPath, cmd.Dir, cmd.Args)
	}

	bcommit, err := cmd.Output()
	if err != nil {
		err = fmt.Errorf("gitGetCommit: %s", err)
		return
	}

	commit = string(bytes.TrimSpace(bcommit))

	return
}

//
// gitGetTag will try to get the current tag from HEAD.
//
func (pkg *Package) gitGetTag() (tag string, err error) {
	//nolint:gas
	cmd := exec.Command("git", "describe", "--tags", "--exact-match")
	cmd.Dir = pkg.FullPath

	if Debug >= DebugL1 {
		fmt.Printf("[PKG] gitGetTag %s >>> %s %s\n", pkg.ImportPath,
			cmd.Dir, cmd.Args)
	}

	btag, err := cmd.Output()
	if err != nil {
		err = fmt.Errorf("gitGetTag: %s", err)
		return
	}

	tag = string(bytes.TrimSpace(btag))

	return
}

func (pkg *Package) gitGetTagLatest() (tag string, err error) {
	//nolint:gas
	cmd := exec.Command("git", "rev-list", "--tags", "--max-count=1")
	cmd.Dir = pkg.FullPath

	if Debug >= DebugL1 {
		fmt.Printf("[PKG] gitGetTagLatest %s >>> %s %s\n",
			pkg.ImportPath, cmd.Dir, cmd.Args)
	}

	bout, err := cmd.Output()
	if err != nil {
		err = fmt.Errorf("gitGetTagLatest: %s", err)
		return
	}

	out := string(bytes.TrimSpace(bout))

	//nolint:gas
	cmd = exec.Command("git", "describe", "--tags", "--abbrev=0", out)
	cmd.Dir = pkg.FullPath

	if Debug >= DebugL1 {
		fmt.Printf("[PKG] gitGetTagLatest %s >>> %s %s\n",
			pkg.ImportPath, cmd.Dir, cmd.Args)
	}

	bout, err = cmd.Output()
	if err != nil {
		err = fmt.Errorf("gitGetTagLatest: %s", err)
		return
	}

	tag = string(bytes.TrimSpace(bout))

	return
}

//
// gitInstall the package into source directory.
//
func (pkg *Package) gitInstall() (err error) {
	err = pkg.gitClone()
	if err != nil {
		err = fmt.Errorf("gitInstall: %s", err)
		return
	}

	var rev string
	if len(pkg.Version) == 0 {
		rev, err = pkg.gitGetTagLatest()
		if err == nil {
			pkg.Version = rev
			pkg.isTag = IsTagVersion(rev)
		} else {
			rev, err = pkg.gitGetCommit(gitRefHEAD)
			if err != nil {
				err = fmt.Errorf("gitInstall: %s", err)
				return
			}

			pkg.Version = rev
		}
	}

	err = pkg.gitCheckoutVersion(pkg.Version)
	if err != nil {
		err = fmt.Errorf("gitInstall: %s", err)
		return
	}

	return
}

//
// gitRemoteChange current package remote name (e.g. "origin") or URL to new
// package remote-name or url.
//
func (pkg *Package) gitRemoteChange(newPkg *Package) (err error) {
	//nolint:gas
	cmd := exec.Command("git", "remote", "remove", pkg.RemoteName)
	cmd.Dir = pkg.FullPath
	cmd.Stdout = defStdout
	cmd.Stderr = defStderr

	if Debug >= DebugL1 {
		fmt.Printf("[PKG] gitRemoteChange %s >>> %s %s\n",
			pkg.ImportPath, cmd.Dir, cmd.Args)
	}

	err = cmd.Run()
	if err != nil {
		fmt.Fprintln(defStderr, "gitRemoteChange:", err)
	}

	//nolint:gas
	cmd = exec.Command("git", "remote", "add", newPkg.RemoteName, newPkg.RemoteURL)
	cmd.Dir = pkg.FullPath
	cmd.Stdout = defStdout
	cmd.Stderr = defStderr

	if Debug >= DebugL1 {
		fmt.Printf("[PKG] gitRemoteChange %s >>> %s %s\n",
			pkg.ImportPath, cmd.Dir, cmd.Args)
	}

	err = cmd.Run()
	if err != nil {
		err = fmt.Errorf("gitRemoteChange: %s", err)
		return
	}

	return
}

//
// gitScan will scan the package version and remote URL.
//
func (pkg *Package) gitScan() (err error) {
	pkg.Version, err = pkg.gitScanVersion()
	if err != nil {
		return
	}

	err = pkg.gitScanRemote()

	return
}

func (pkg *Package) gitScanRemote() (err error) {
	gitConfig := filepath.Join(pkg.FullPath, gitDir, "config")

	gitIni, err := ini.Open(gitConfig)
	if err != nil {
		err = fmt.Errorf("gitScanRemote: %s", err)
		return
	}

	url, ok := gitIni.Get(gitCfgRemote, gitDefRemoteName, gitCfgRemoteURL)
	if !ok {
		err = fmt.Errorf("gitScanRemote: %s", ErrRemote)
		return
	}

	pkg.RemoteName = gitDefRemoteName
	pkg.RemoteURL = url

	return
}

//
// gitScanVersion will try to,
// (1) get latest tag from repository first, or if it's fail
// (2) get the commit hash at HEAD.
//
func (pkg *Package) gitScanVersion() (version string, err error) {
	// (1)
	version, err = pkg.gitGetTag()
	if err == nil {
		return
	}

	// (2)
	version, err = pkg.gitGetCommit(gitRefHEAD)
	if err != nil {
		err = ErrVersion
	}

	return
}

//
// gitUpdate will change the currrent package remote name, URL, or version
// based on new package information.
//
func (pkg *Package) gitUpdate(newPkg *Package) (err error) {
	if pkg.RemoteName != newPkg.RemoteName || pkg.RemoteURL != newPkg.RemoteURL {
		err = pkg.gitRemoteChange(newPkg)
		if err != nil {
			return
		}
	}

	err = pkg.gitFetch()
	if err != nil {
		err = fmt.Errorf("gitUpdate: %s", err)
		return
	}

	err = pkg.gitCheckoutVersion(newPkg.Version)
	if err != nil {
		err = fmt.Errorf("gitUpdate: %s", err)
	}

	return
}
