package beku

import (
	"io"
	"io/ioutil"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

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
