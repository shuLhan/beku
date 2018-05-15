// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package beku

import (
	"bytes"
	"fmt"
	"strings"
)

//
// IsIgnoredDir will return true if directory `name` start with `_` or `.`, or
// equal with `vendor` or `testdata`; otherwise it will return false.
//
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

func confirm(msg string, defIsYes bool) bool {
	var response string
	yon := "[y/N]"

	if defIsYes {
		yon = "[Y/n]"
	}

	fmt.Printf("%s %s: ", msg, yon)

	_, err := fmt.Scanln(&response)
	if err != nil {
		return defIsYes
	}

	response = strings.TrimSpace(response)

	if response[0] == 'y' || response[0] == 'Y' {
		return true
	}
	if len(response) == 0 {
		return defIsYes
	}

	return false
}

func parsePkgVersion(pkgVersion string) (pkgName, version string, err error) {
	if len(pkgVersion) == 0 {
		err = ErrPackageName
		return
	}

	var (
		x   int
		buf bytes.Buffer
	)

	for ; x < len(pkgVersion); x++ {
		if pkgVersion[x] == sepImportVersion {
			break
		}

		buf.WriteByte(pkgVersion[x])
	}

	pkgName = buf.String()
	buf.Reset()

	for ; x < len(pkgVersion); x++ {
		buf.WriteByte(pkgVersion[x])
	}

	if buf.Len() > 0 {
		version = buf.String()
	}

	return
}
