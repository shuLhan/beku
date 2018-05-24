// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package beku

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
)

//
// IsDirEmpty will return true if directory is not exist or empty; otherwise
// it will return false.
//
func IsDirEmpty(dir string) (ok bool) {
	d, err := os.Open(dir)
	if err != nil {
		ok = true
		return
	}

	_, err = d.Readdirnames(1)
	if err != nil {
		if err == io.EOF {
			ok = true
		}
	}

	_ = d.Close()

	return
}

//
// IsIgnoredDir will return true if directory start with "_" or ".", or
// equal with "vendor" or "testdata"; otherwise it will return false.
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

//
// IsTagVersion return true if "version" prefixed with "v" or contains at
// least one dot "." character.
//
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

//
// confirm display a question to standard output and read for answer
// from "in" for simple "y" or "n" answer.
// If "in" is nil, it will set to standard input.
// If "defIsYes" is true and answer is empty (only new line), then it will
// return true.
//
func confirm(in io.Reader, msg string, defIsYes bool) bool {
	var (
		r         *bufio.Reader
		b, answer byte
		err       error
	)

	if in == nil {
		r = bufio.NewReader(os.Stdin)
	} else {
		r = bufio.NewReader(in)
	}

	yon := "[y/N]"

	if defIsYes {
		yon = "[Y/n]"
	}

	fmt.Printf("%s %s: ", msg, yon)

	for {
		b, err = r.ReadByte()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			break
		}
		if b == ' ' || b == '\t' {
			continue
		}
		if b == '\n' {
			break
		}
		if answer == 0 {
			answer = b
		}
	}

	if answer == 'y' || answer == 'Y' {
		return true
	}
	if answer == 0 {
		return defIsYes
	}

	return false
}

//
// parsePkgVersion given the following package-version format "pkg@v1.0.0", it
// will return "pkg" and "v1.0.0".
//
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

	return
}
