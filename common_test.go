// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package beku

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestGetCompareURL(t *testing.T) {
	cases := []struct {
		desc      string
		remoteURL string
		oldVer    string
		newVer    string
		exp       string
	}{{
		desc: "With empty remoteURL",
	}, {
		desc:      "With git format",
		remoteURL: "git@github.com:shuLhan/beku.git",
		oldVer:    "A",
		newVer:    "B",
		exp:       "https://github.com/shuLhan/beku/compare/A...B",
	}, {
		desc:      "With HTTP format",
		remoteURL: "https://github.com/shuLhan/beku",
		oldVer:    "A",
		newVer:    "B",
		exp:       "https://github.com/shuLhan/beku/compare/A...B",
	}, {
		desc:      "With golang.org as hostname",
		remoteURL: "https://golang.org/x/net",
		oldVer:    "A",
		newVer:    "B",
		exp:       "https://github.com/golang/net/compare/A...B",
	}, {
		desc:      "With unknown hostname",
		remoteURL: "https://gopkg.in/yaml.v2",
		oldVer:    "A",
		newVer:    "B",
	}}

	for _, c := range cases {
		t.Log(c.desc)

		got := GetCompareURL(c.remoteURL, c.oldVer, c.newVer)

		test.Assert(t, "", c.exp, got, true)
	}
}

func TestIsIgnoredDir(t *testing.T) {
	cases := []struct {
		name string
		exp  bool
	}{{
		name: "notignored",
		exp:  false,
	}, {
		name: "_dashed",
		exp:  true,
	}, {
		name: "d_ashed",
		exp:  false,
	}, {
		name: "dashed_",
		exp:  false,
	}, {
		name: ".dotted",
		exp:  true,
	}, {
		name: "d.otted",
		exp:  false,
	}, {
		name: "dotted.",
		exp:  false,
	}, {
		name: "vendored",
		exp:  false,
	}, {
		name: "vendor",
		exp:  true,
	}, {
		name: "testdata",
		exp:  true,
	}, {
		name: "test_data",
		exp:  false,
	}}

	var got bool
	for _, c := range cases {
		t.Log(c)
		got = IsIgnoredDir(c.name)
		test.Assert(t, "", c.exp, got, true)
	}
}

func TestIsTagVersion(t *testing.T) {
	cases := []struct {
		ver string
		exp bool
	}{{
		ver: "",
	}, {
		ver: " v",
	}, {
		ver: "v1",
		exp: true,
	}, {
		ver: "1",
		exp: false,
	}, {
		ver: "1.0",
		exp: true,
	}, {
		ver: "alpha",
		exp: false,
	}, {
		ver: "abcdef1",
		exp: false,
	}}

	var got bool
	for _, c := range cases {
		t.Log(c)

		got = IsTagVersion(c.ver)

		test.Assert(t, "", c.exp, got, true)
	}
}

func TestParsePkgVersion(t *testing.T) {
	cases := []struct {
		pkgName string
		expPkg  string
		expVer  string
	}{{
		pkgName: "",
	}, {
		pkgName: " pkg",
		expPkg:  "pkg",
	}, {
		pkgName: " pkg @",
		expPkg:  "pkg",
	}, {
		pkgName: " @ 123",
		expVer:  "123",
	}, {
		pkgName: " pkg @ 123",
		expPkg:  "pkg",
		expVer:  "123",
	}}

	var gotPkg, gotVer string

	for _, c := range cases {
		t.Log(c)

		gotPkg, gotVer = parsePkgVersion(c.pkgName)

		test.Assert(t, "package name", c.expPkg, gotPkg, true)
		test.Assert(t, "version", c.expVer, gotVer, true)
	}
}
