// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package beku

const (
	testdataDir = "testdata"
	vendorDir   = "vendor"
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
	if name == testdataDir || name == vendorDir {
		return true
	}

	return false
}
