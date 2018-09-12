// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package beku

import (
	"fmt"

	"github.com/shuLhan/share/lib/git"
)

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
