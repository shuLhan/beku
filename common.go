// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package beku

import (
	"bytes"
	"fmt"
	"strings"
)

// GetCompareURL return the URL that compare two versions of package from
// given remote URL. Remote URL can be in git format
// ("git@github.com:<username>/<reponame>") or in HTTP format.
//
// On package that hosted on Github, the compare URL format is,
//
//	https://github.com/<username>/<reponame>/compare/<old-version>...<new-version>
func GetCompareURL(remoteURL, oldVer, newVer string) (url string) {
	if len(remoteURL) == 0 {
		return
	}

	remoteURL = strings.TrimPrefix(remoteURL, "git@")
	remoteURL = strings.TrimPrefix(remoteURL, "https://")
	remoteURL = strings.TrimPrefix(remoteURL, "www.")
	remoteURL = strings.TrimSuffix(remoteURL, ".git")

	var host, user, repo string

	colIdx := strings.IndexByte(remoteURL, ':')
	if colIdx > 0 {
		names := strings.Split(remoteURL[colIdx+1:], "/")

		host = remoteURL[0:colIdx]
		user = names[0]
		repo = names[len(names)-1]
	} else {
		names := strings.Split(remoteURL, "/")
		if len(names) < 3 {
			return
		}
		host = names[0]
		user = names[1]
		repo = names[len(names)-1]
	}

	switch host {
	case "github.com":
		url = fmt.Sprintf("https://%s/%s/%s/compare/%s...%s", host,
			user, repo, oldVer, newVer)
	case "golang.org":
		url = fmt.Sprintf("https://github.com/golang/%s/compare/%s...%s",
			repo, oldVer, newVer)
	}

	return url
}

// IsIgnoredDir will return true if directory start with "_" or ".", or
// equal with "vendor" or "testdata"; otherwise it will return false.
func IsIgnoredDir(name string) bool {
	prefix := name[0]

	if prefix == '_' || prefix == '.' {
		return true
	}
	if name == dirTestdata || name == dirVendor {
		return true
	}

	return false
}

// IsTagVersion return true if "version" prefixed with "v" or contains at
// least one dot "." character.
func IsTagVersion(version string) bool {
	version = strings.TrimSpace(version)
	if len(version) == 0 {
		return false
	}
	if version[0] == prefixTag && len(version) > 1 {
		return true
	}
	if strings.IndexByte(version, sepVersion) > 0 {
		return true
	}
	return false
}

// parsePkgVersion given the following package-version format "pkg@v1.0.0", it
// will return "pkg" and "v1.0.0".
func parsePkgVersion(pkgVersion string) (pkgName, version string) {
	if len(pkgVersion) == 0 {
		return
	}

	var (
		x   int
		buf bytes.Buffer
	)

	for ; x < len(pkgVersion); x++ {
		if pkgVersion[x] == sepImportVersion {
			x++
			break
		}

		buf.WriteByte(pkgVersion[x])
	}

	if buf.Len() > 0 {
		pkgName = buf.String()
		pkgName = strings.TrimSpace(pkgName)
		buf.Reset()
	}

	for ; x < len(pkgVersion); x++ {
		buf.WriteByte(pkgVersion[x])
	}

	if buf.Len() > 0 {
		version = buf.String()
		version = strings.TrimSpace(version)
	}

	return pkgName, version
}
