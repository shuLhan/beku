// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package beku

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestGitScan(t *testing.T) {
	cases := []struct {
		desc       string
		expErr     string
		expVersion string
		expIsTag   bool
	}{{
		desc:       "Using current package",
		expVersion: "v0.2.0",
		expIsTag:   true,
	}}

	var err error
	for _, c := range cases {
		t.Log(c.desc)

		testGitPkgCur.Version = ""
		testGitPkgCur.isTag = false

		err = testGitPkgCur.Scan()
		if err != nil {
			test.Assert(t, "err", c.expErr, err, true)
			continue
		}

		test.Assert(t, "Version", c.expVersion, testGitPkgCur.Version, true)
		test.Assert(t, "isTag", c.expIsTag, testGitPkgCur.isTag, true)
	}
}

func TestGitScanDeps(t *testing.T) {
	cases := []struct {
		expErr         string
		expDeps        []string
		expDepsMissing []string
		expPkgsMissing []string
	}{{}}

	var err error

	for _, c := range cases {
		testGitPkgCur.Deps = nil
		testGitPkgCur.DepsMissing = nil

		err = testGitPkgCur.ScanDeps(testEnv)
		if err != nil {
			test.Assert(t, "err", c.expErr, err.Error(), true)
		}

		test.Assert(t, "Deps", c.expDeps, testGitPkgCur.Deps, true)
		test.Assert(t, "DepsMissing", c.expDepsMissing,
			testGitPkgCur.DepsMissing, true)
		test.Assert(t, "env.pkgsMissing", c.expPkgsMissing,
			testEnv.pkgsMissing, true)
	}
}
