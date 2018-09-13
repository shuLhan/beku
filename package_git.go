// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package beku

import (
	"fmt"
	"strings"

	"github.com/shuLhan/share/lib/debug"
	"github.com/shuLhan/share/lib/git"
)

func (pkg *Package) gitFreeze() (err error) {
	err = git.FetchAll(pkg.FullPath)
	if err != nil {
		return
	}
	if len(pkg.RemoteBranch) == 0 {
		err = pkg.gitGetBranch()
		if err != nil {
			return
		}
	}

	err = git.CheckoutRevision(pkg.FullPath, pkg.RemoteName,
		pkg.RemoteBranch, pkg.Version)

	return
}

//
// gitInstall the package into source directory.
//
func (pkg *Package) gitInstall() (err error) {
	err = git.Clone(pkg.RemoteURL, pkg.FullPath)
	if err != nil {
		err = fmt.Errorf("gitInstall: %s", err)
		return
	}

	var rev string
	if len(pkg.Version) == 0 {
		rev, err = git.LatestTag(pkg.FullPath)
		if err == nil {
			pkg.Version = rev
			pkg.isTag = IsTagVersion(rev)
		} else {
			rev, err = git.LatestCommit(pkg.FullPath, "")
			if err != nil {
				err = fmt.Errorf("gitInstall: %s", err)
				return
			}

			pkg.Version = rev
		}
	}

	if pkg.isTag {
		err = git.CheckoutRevision(pkg.FullPath, "", "", pkg.Version)
		if err != nil {
			err = fmt.Errorf("gitInstall: %s", err)
			return
		}
	}

	return
}

//
// gitScan will scan the package version and remote URL.
//
func (pkg *Package) gitScan() (err error) {
	pkg.Version, err = git.LatestVersion(pkg.FullPath)
	if err != nil {
		return
	}

	pkg.RemoteURL, err = git.GetRemoteURL(pkg.FullPath, "")
	if err != nil {
		err = fmt.Errorf("gitScan: %s", err)
		return
	}

	err = pkg.gitGetBranch()

	return
}

func (pkg *Package) gitGetBranch() (err error) {
	branches, err := git.RemoteBranches(pkg.FullPath)
	if err != nil {
		err = fmt.Errorf("gitGetBranch: %s", err)
		return
	}

	// Select branch by version, master, or the last branch.
	midx := -1
	vidx := -1
	for x := 0; x < len(branches); x++ {
		if branches[x] == gitDefBranch {
			midx = x
		}
		if branches[x][0] == 'v' {
			if vidx < 0 {
				vidx = x
				continue
			}
			if strings.Compare(branches[vidx], branches[x]) == -1 {
				vidx = x
			}
		}
	}
	if midx >= 0 {
		pkg.RemoteBranch = branches[midx]
	} else if vidx >= 0 {
		pkg.RemoteBranch = branches[vidx]
	} else {
		pkg.RemoteBranch = branches[len(branches)-1]
	}
	if debug.Value >= 1 {
		fmt.Printf("= gitGetBranch: %s\n", pkg.RemoteBranch)
	}
	return
}

//
// gitUpdate will change the currrent package remote name, URL, or version
// based on new package information.
//
func (pkg *Package) gitUpdate(newPkg *Package) (err error) {
	if pkg.RemoteName != newPkg.RemoteName || pkg.RemoteURL != newPkg.RemoteURL {
		err = git.RemoteChange(pkg.FullPath, pkg.RemoteName,
			newPkg.RemoteName, newPkg.RemoteURL)
		if err != nil {
			return
		}
	}

	err = git.FetchAll(pkg.FullPath)
	if err != nil {
		err = fmt.Errorf("gitUpdate: %s", err)
		return
	}

	err = git.CheckoutRevision(pkg.FullPath, "", "", newPkg.Version)
	if err != nil {
		err = fmt.Errorf("gitUpdate: %s", err)
	}

	return
}
