package beku

import (
	"testing"

	"github.com/shuLhan/share/lib/ini"
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

func TestAddDep(t *testing.T) {
	cases := []struct {
		desc           string
		envPkgs        []*Package
		importPath     string
		exp            bool
		expDeps        []string
		expDepsMissing []string
		expPkgsMissing []string
	}{{
		desc:       "Is the same path as package",
		importPath: testGitRepo,
	}, {
		desc:       "Is vendor package",
		importPath: "vendor/github.com/shuLhan/beku",
	}, {
		desc:       "Is standard package",
		importPath: "os/exec",
	}, {
		desc: "Is exist on environment",
		envPkgs: []*Package{{
			ImportPath: "github.com/shuLhan/beku",
		}, {
			ImportPath: "github.com/shuLhan/share",
		}},
		importPath: "github.com/shuLhan/share/lib/test",
		exp:        true,
		expDeps: []string{
			"github.com/shuLhan/share",
		},
	}, {
		desc: "Is exist on environment (again)",
		envPkgs: []*Package{{
			ImportPath: "github.com/shuLhan/beku",
		}, {
			ImportPath: "github.com/shuLhan/share",
		}},
		importPath: "github.com/shuLhan/share/lib/test",
		exp:        true,
		expDeps: []string{
			"github.com/shuLhan/share",
		},
	}, {
		desc:       "Is not exist on environment (missing)",
		importPath: "github.com/shuLhan/tekstus",
		exp:        true,
		expDeps: []string{
			"github.com/shuLhan/share",
		},
		expDepsMissing: []string{
			"github.com/shuLhan/tekstus",
		},
		expPkgsMissing: []string{
			"github.com/shuLhan/tekstus",
		},
	}, {
		desc:       "Is not exist on environment (again)",
		importPath: "github.com/shuLhan/tekstus",
		exp:        true,
		expDeps: []string{
			"github.com/shuLhan/share",
		},
		expDepsMissing: []string{
			"github.com/shuLhan/tekstus",
		},
		expPkgsMissing: []string{
			"github.com/shuLhan/tekstus",
		},
	}}

	var got bool

	gitCurPkg.Deps = nil
	gitCurPkg.DepsMissing = nil

	for _, c := range cases {
		t.Log(c.desc)

		testEnv.pkgs = c.envPkgs
		testEnv.pkgsMissing = nil
		got = gitCurPkg.addDep(testEnv, c.importPath)

		test.Assert(t, "return", c.exp, got, true)

		if !got {
			continue
		}

		test.Assert(t, "Deps", c.expDeps, gitCurPkg.Deps, true)
		test.Assert(t, "DepsMissing", c.expDepsMissing, gitCurPkg.DepsMissing, true)
		test.Assert(t, "env.pkgsMissing", c.expPkgsMissing,
			testEnv.pkgsMissing, true)
	}

	gitCurPkg.Deps = nil
	gitCurPkg.DepsMissing = nil
}

func TestPushRequiredBy(t *testing.T) {
	cases := []struct {
		desc          string
		parentPkg     *Package
		exp           bool
		expRequiredBy []string
	}{{
		desc: "Not exist yet",
		parentPkg: &Package{
			ImportPath: "github.com/shuLhan/share",
		},
		exp: true,
		expRequiredBy: []string{
			"github.com/shuLhan/share",
		},
	}, {
		desc: "Already exist",
		parentPkg: &Package{
			ImportPath: "github.com/shuLhan/share",
		},
		expRequiredBy: []string{
			"github.com/shuLhan/share",
		},
	}}

	gitCurPkg.RequiredBy = nil

	var got bool

	for _, c := range cases {
		t.Log(c.desc)

		got = gitCurPkg.pushRequiredBy(c.parentPkg.ImportPath)

		test.Assert(t, "return value", c.exp, got, true)
		test.Assert(t, "RequiredBy", c.expRequiredBy,
			gitCurPkg.RequiredBy, true)
	}

	gitCurPkg.RequiredBy = nil
}

func TestPackageLoad(t *testing.T) {
	cases := []struct {
		desc    string
		pkgName string
		exp     *Package
	}{{
		desc:    "With invalid vcs mode",
		pkgName: "test_vcs",
		exp: &Package{
			vcs: VCSModeGit,
		},
	}, {
		desc:    "Duplicate remote name",
		pkgName: "dup_remote_name",
		exp: &Package{
			vcs:        VCSModeGit,
			RemoteName: "last",
		},
	}, {
		desc:    "Duplicate remote URL",
		pkgName: "dup_remote_url",
		exp: &Package{
			vcs:        VCSModeGit,
			RemoteName: "last",
			RemoteURL:  "remote url 2",
		},
	}, {
		desc:    "Duplicate version",
		pkgName: "dup_version",
		exp: &Package{
			vcs:        VCSModeGit,
			RemoteName: "last",
			RemoteURL:  "remote url 2",
			Version:    "v1.1.0",
			isTag:      true,
		},
	}, {
		desc:    "Version not tag",
		pkgName: "version_not_tag",
		exp: &Package{
			vcs:        VCSModeGit,
			RemoteName: "last",
			RemoteURL:  "remote url 2",
			Version:    "12345678",
			isTag:      false,
		},
	}, {
		desc:    "With deps",
		pkgName: "deps",
		exp: &Package{
			vcs:        VCSModeGit,
			RemoteName: "last",
			RemoteURL:  "remote url 2",
			Version:    "12345678",
			isTag:      false,
			Deps: []string{
				"dep/1",
				"dep/2",
				"dep/3",
			},
		},
	}, {
		desc:    "With missing deps",
		pkgName: "deps_missing",
		exp: &Package{
			vcs:        VCSModeGit,
			RemoteName: "last",
			RemoteURL:  "remote url 2",
			Version:    "12345678",
			isTag:      false,
			DepsMissing: []string{
				"missing/1",
				"missing/2",
				"missing/3",
			},
		},
	}, {
		desc:    "With required-by",
		pkgName: "required-by",
		exp: &Package{
			vcs:        VCSModeGit,
			RemoteName: "last",
			RemoteURL:  "remote url 2",
			Version:    "12345678",
			isTag:      false,
			RequiredBy: []string{
				"required-by/3",
				"required-by/2",
				"required-by/1",
			},
		},
	}}

	cfg, err := ini.Open("testdata/package_load.conf")
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range cases {
		t.Log(c.desc)

		pkg := new(Package)
		sec := cfg.GetSection(sectionPackage, c.pkgName)

		pkg.load(sec)

		test.Assert(t, "", c.exp, pkg, true)
	}
}
