package beku

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestIsEqual(t *testing.T) {
	cases := []struct {
		desc  string
		pkg   *Package
		other *Package
		exp   bool
	}{{
		desc: "With nil on other",
		pkg:  &Package{},
	}, {
		desc: "With different ImportPath",
		pkg: &Package{
			ImportPath: "git",
		},
		other: &Package{
			ImportPath: "gitt",
		},
	}, {
		desc: "With different RemoteName",
		pkg: &Package{
			ImportPath: "git",
			RemoteName: "origin",
		},
		other: &Package{
			ImportPath: "git",
			RemoteName: "upstream",
		},
	}, {
		desc: "With different RemoteURL",
		pkg: &Package{
			ImportPath: "git",
			RemoteName: "origin",
			RemoteURL:  "https://github.com/shuLhan/beku",
		},
		other: &Package{
			ImportPath: "git",
			RemoteName: "origin",
			RemoteURL:  "https://gopkg.in/shuLhan/beku",
		},
	}, {
		desc: "With different Version",
		pkg: &Package{
			ImportPath: "git",
			RemoteName: "origin",
			RemoteURL:  "https://github.com/shuLhan/beku",
			Version:    "v0.1.0",
		},
		other: &Package{
			ImportPath: "git",
			RemoteName: "origin",
			RemoteURL:  "https://github.com/shuLhan/beku",
			Version:    "v0.1.1",
		},
	}, {
		desc: "With equal",
		pkg: &Package{
			ImportPath: "git",
			RemoteName: "origin",
			RemoteURL:  "https://github.com/shuLhan/beku",
			Version:    "v0.1.0",
		},
		other: &Package{
			ImportPath: "git",
			RemoteName: "origin",
			RemoteURL:  "https://github.com/shuLhan/beku",
			Version:    "v0.1.0",
		},
		exp: true,
	}}

	var got bool
	for _, c := range cases {
		t.Log(c.desc)

		got = c.pkg.IsEqual(c.other)

		test.Assert(t, "", c.exp, got, true)
	}
}
