// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package beku

import (
	"io"
	"io/ioutil"
	"os"
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

func TestIsDirEmpty(t *testing.T) {
	emptyDir := "testdata/dirempty"
	err := os.MkdirAll(emptyDir, 0700)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		desc string
		path string
		exp  bool
	}{{
		desc: `With dir not exist`,
		path: `testdata/notexist`,
		exp:  true,
	}, {
		desc: `With dir exist and not empty`,
		path: `testdata`,
	}, {
		desc: `With dir exist and empty`,
		path: `testdata/dirempty`,
		exp:  true,
	}}

	for _, c := range cases {
		t.Log(c.desc)

		got := IsDirEmpty(c.path)

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

func TestRmdirEmptyAll(t *testing.T) {
	cases := []struct {
		desc        string
		createDir   string
		createFile  string
		path        string
		expExist    string
		expNotExist string
	}{{
		desc:     "With path as file",
		path:     "testdata/beku.db",
		expExist: "testdata/beku.db",
	}, {
		desc:      "With empty path",
		createDir: "testdata/a/b/c/d",
		expExist:  "testdata/a/b/c/d",
	}, {
		desc:        "With non empty at middle",
		createDir:   "testdata/a/b/c/d",
		createFile:  "testdata/a/b/file",
		path:        "testdata/a/b/c/d",
		expExist:    "testdata/a/b/file",
		expNotExist: "testdata/a/b/c",
	}, {
		desc:        "With first path not exist",
		createDir:   "testdata/a/b/c",
		path:        "testdata/a/b/c/d",
		expExist:    "testdata/a/b/file",
		expNotExist: "testdata/a/b/c",
	}, {
		desc:        "With non empty at parent",
		createDir:   "testdata/dirempty/a/b/c/d",
		path:        "testdata/dirempty/a/b/c/d",
		expExist:    "testdata",
		expNotExist: "testdata/dirempty",
	}}

	var (
		err error
		f   *os.File
	)
	for _, c := range cases {
		t.Log(c.desc)

		if len(c.createDir) > 0 {
			err = os.MkdirAll(c.createDir, 0700)
			if err != nil {
				t.Fatal(err)
			}
		}
		if len(c.createFile) > 0 {
			f, err = os.Create(c.createFile)
			if err != nil {
				t.Fatal(err)
			}
			err = f.Close()
			if err != nil {
				t.Fatal(err)
			}
		}

		err = RmdirEmptyAll(c.path)
		if err != nil {
			t.Fatal(err)
		}

		if len(c.expExist) > 0 {
			_, err = os.Stat(c.expExist)
			if err != nil {
				t.Fatal(err)
			}
		}
		if len(c.expNotExist) > 0 {
			_, err = os.Stat(c.expNotExist)
			if !os.IsNotExist(err) {
				t.Fatal(err)
			}
		}
	}
}

func TestConfirm(t *testing.T) {
	cases := []struct {
		defIsYes bool
		answer   string
		exp      bool
	}{{
		defIsYes: true,
		exp:      true,
	}, {
		defIsYes: true,
		answer:   "  ",
		exp:      true,
	}, {
		defIsYes: true,
		answer:   "  no",
		exp:      false,
	}, {
		defIsYes: true,
		answer:   " yes",
		exp:      true,
	}, {
		defIsYes: true,
		answer:   " Ys",
		exp:      true,
	}, {
		defIsYes: false,
		exp:      false,
	}, {
		defIsYes: false,
		answer:   "",
		exp:      false,
	}, {

		defIsYes: false,
		answer:   "  no",
		exp:      false,
	}, {
		defIsYes: false,
		answer:   "  yes",
		exp:      true,
	}}

	var got bool

	in, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}

	defer in.Close()

	for _, c := range cases {
		t.Log(c)

		in.WriteString(c.answer + "\n")

		_, err = in.Seek(0, io.SeekStart)
		if err != nil {
			t.Fatal(err)
		}

		got = confirm(in, "confirm", c.defIsYes)

		test.Assert(t, "answer", c.exp, got, true)

		err = in.Truncate(0)
		if err != nil {
			t.Fatal(err)
		}

		_, err = in.Seek(0, io.SeekStart)
		if err != nil {
			t.Fatal(err)
		}
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
