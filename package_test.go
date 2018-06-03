package beku

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shuLhan/share/lib/ini"
	"github.com/shuLhan/share/lib/test"
)

func testPackageRemove(t *testing.T) {
	cases := []struct {
		desc    string
		pkgName string
		pkg     *Package
		expErr  string
	}{{
		desc:    `Package is not exist`,
		pkgName: testPkgNotExist,
	}, {
		desc: `Package exist`,
		pkg:  testGitPkgShare,
	}}

	for _, c := range cases {
		t.Log(c.desc)

		if len(c.pkgName) > 0 {
			c.pkg, _ = NewPackage(c.pkgName, c.pkgName, VCSModeGit)
		}

		err := c.pkg.Remove()
		if err != nil {
			test.Assert(t, "err", c.expErr, err, true)
			continue
		}

		expErr := fmt.Sprintf("stat %s: no such file or directory", c.pkg.FullPath)
		_, err = os.Stat(c.pkg.FullPath)

		test.Assert(t, "src dir should not exist", expErr, err.Error(), true)

		pkg := filepath.Join(testEnv.dirPkg, c.pkg.ImportPath)

		expErr = fmt.Sprintf("stat %s: no such file or directory", pkg)
		_, err = os.Stat(pkg)

		if err != nil {
			test.Assert(t, "pkg dir should not exist", expErr, err.Error(), true)
		}
	}
}

func testPackageInstall(t *testing.T) {
	cases := []struct {
		desc   string
		pkg    *Package
		expErr string
		expPkg *Package
	}{{
		desc: `Without version`,
		pkg:  testGitPkgShare,
		expPkg: &Package{
			ImportPath: testGitRepoShare,
			FullPath:   testGitPkgShare.FullPath,
			RemoteName: gitDefRemoteName,
			RemoteURL:  "https://" + testGitRepoShare,
			Version:    "17828b8",
			vcs:        VCSModeGit,
			state:      packageStateNew,
		},
	}, {
		desc:   `Install again`,
		pkg:    testGitPkgShare,
		expErr: fmt.Sprintf("gitInstall: gitClone: "+errDirNotEmpty, testGitPkgShare.FullPath),
	}}

	for _, c := range cases {
		t.Log(c.desc)

		testResetOutput(t, true)

		err := c.pkg.Install()

		testPrintOutput(t)

		if err != nil {
			test.Assert(t, "err", c.expErr, err.Error(), true)
			continue
		}

		test.Assert(t, "pkg", *c.expPkg, *c.pkg, true)
	}
}

func testIsEqual(t *testing.T) {
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

func testAddDep(t *testing.T) {
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

	testGitPkgCur.Deps = nil
	testGitPkgCur.DepsMissing = nil

	for _, c := range cases {
		t.Log(c.desc)

		testEnv.pkgs = c.envPkgs
		testEnv.pkgsMissing = nil
		got = testGitPkgCur.addDep(testEnv, c.importPath)

		test.Assert(t, "return", c.exp, got, true)

		if !got {
			continue
		}

		test.Assert(t, "Deps", c.expDeps, testGitPkgCur.Deps, true)
		test.Assert(t, "DepsMissing", c.expDepsMissing, testGitPkgCur.DepsMissing, true)
		test.Assert(t, "env.pkgsMissing", c.expPkgsMissing,
			testEnv.pkgsMissing, true)
	}

	testGitPkgCur.Deps = nil
	testGitPkgCur.DepsMissing = nil
}

func testPushRequiredBy(t *testing.T) {
	cases := []struct {
		desc          string
		importPath    string
		exp           bool
		expRequiredBy []string
	}{{
		desc:       "Not exist yet",
		importPath: testGitRepoShare,
		exp:        true,
		expRequiredBy: []string{
			testGitRepoShare,
		},
	}, {
		desc:       "Already exist",
		importPath: testGitRepoShare,
		expRequiredBy: []string{
			testGitRepoShare,
		},
	}}

	testGitPkgCur.RequiredBy = nil

	var got bool

	for _, c := range cases {
		t.Log(c.desc)

		got = testGitPkgCur.pushRequiredBy(c.importPath)

		test.Assert(t, "return value", c.exp, got, true)
		test.Assert(t, "RequiredBy", c.expRequiredBy,
			testGitPkgCur.RequiredBy, true)
	}
}

func testPackageRemoveRequiredBy(t *testing.T) {
	cases := []struct {
		desc          string
		pkg           *Package
		importPath    string
		exp           bool
		expRequiredBy []string
	}{{
		desc:       `With importPath not found`,
		pkg:        testGitPkgCur,
		importPath: testPkgNotExist,
		exp:        false,
		expRequiredBy: []string{
			testGitRepoShare,
		},
	}, {
		desc:       `With importPath found`,
		pkg:        testGitPkgCur,
		importPath: testGitRepoShare,
		exp:        true,
	}}

	for _, c := range cases {
		t.Log(c.desc)
		t.Log(">>> RequiredBy:", c.pkg.RequiredBy)

		got := c.pkg.RemoveRequiredBy(c.importPath)

		test.Assert(t, "return value", c.exp, got, true)
		test.Assert(t, "RequiredBy", c.expRequiredBy, c.pkg.RequiredBy, true)
	}
}

func testPackageLoad(t *testing.T) {
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

func testGoInstall(t *testing.T) {
	cases := []struct {
		desc       string
		pkg        *Package
		isVerbose  bool
		expBin     string
		expArchive string
		expStdout  string
		expStderr  string
	}{{
		desc: "Running #1",
		pkg:  testGitPkgCur,
	}, {
		desc:      "Running with verbose",
		pkg:       testGitPkgCur,
		isVerbose: true,
	}}

	for _, c := range cases {
		t.Log(c.desc)

		testResetOutput(t, true)

		err := c.pkg.GoInstall()

		testResetOutput(t, false)
		stdout, stderr := testGetOutput(t)

		if err != nil {
			errLines := strings.Split(stderr, "\n")
			test.Assert(t, "stderr", c.expStderr, errLines[0], true)
		} else {
			outLines := strings.Split(stdout, "\n")
			test.Assert(t, "stdout", c.expStdout, outLines[0], true)
		}

		if len(c.expBin) > 0 {
			_, err = os.Stat(c.expBin)
			if err != nil {
				t.Fatal(err)
			}
		}

		if len(c.expArchive) > 0 {
			_, err = os.Stat(c.expArchive)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
}

func testPackageString(t *testing.T) {
	cases := []struct {
		pkg *Package
		exp string
	}{{
		pkg: testGitPkgCur,
		exp: `
[package "github.com/shuLhan/beku_test"]
          VCS = 1
   RemoteName = origin
    RemoteURL = https://` + testGitRepo + `
      Version = v0.2.0
        IsTag = true
         Deps = []
   RequiredBy = []
  DepsMissing = []
`,
	}}

	for _, c := range cases {
		got := c.pkg.String()
		test.Assert(t, "string", c.exp, got, true)
	}
}

func testUpdate(t *testing.T) {
	cases := []struct {
		desc   string
		curPkg *Package
		newPkg *Package
		expErr error
		expPkg *Package
	}{{
		desc: "Update remote URL",
		curPkg: &Package{
			vcs:        VCSModeGit,
			ImportPath: testGitRepo,
			FullPath:   filepath.Join(testEnv.dirSrc, testGitRepo),
			RemoteName: gitDefRemoteName,
			RemoteURL:  "https://" + testGitRepo,
		},
		newPkg: &Package{
			vcs:        VCSModeGit,
			ImportPath: testGitRepo,
			FullPath:   filepath.Join(testEnv.dirSrc, testGitRepo),
			RemoteName: gitDefRemoteName,
			RemoteURL:  "git@github.com:shuLhan/beku_test.git",
		},
		expPkg: &Package{
			vcs:        VCSModeGit,
			ImportPath: testGitRepo,
			FullPath:   filepath.Join(testEnv.dirSrc, testGitRepo),
			RemoteName: gitDefRemoteName,
			RemoteURL:  "git@github.com:shuLhan/beku_test.git",
		},
	}, {
		desc: "Update version",
		curPkg: &Package{
			vcs:        VCSModeGit,
			ImportPath: testGitRepo,
			FullPath:   filepath.Join(testEnv.dirSrc, testGitRepo),
			RemoteName: gitDefRemoteName,
			RemoteURL:  "https://" + testGitRepo,
		},
		newPkg: &Package{
			vcs:        VCSModeGit,
			ImportPath: testGitRepo,
			FullPath:   filepath.Join(testEnv.dirSrc, testGitRepo),
			RemoteName: gitDefRemoteName,
			RemoteURL:  "git@github.com:shuLhan/beku_test.git",
			Version:    "v0.1.0",
			isTag:      true,
		},
		expPkg: &Package{
			vcs:         VCSModeGit,
			ImportPath:  testGitRepo,
			FullPath:    filepath.Join(testEnv.dirSrc, testGitRepo),
			RemoteName:  gitDefRemoteName,
			RemoteURL:   "git@github.com:shuLhan/beku_test.git",
			Version:     "v0.1.0",
			VersionNext: "c9f69fb",
			isTag:       true,
		},
	}, {
		desc: "Update version back",
		curPkg: &Package{
			vcs:        VCSModeGit,
			ImportPath: testGitRepo,
			FullPath:   filepath.Join(testEnv.dirSrc, testGitRepo),
			RemoteName: gitDefRemoteName,
			RemoteURL:  "https://" + testGitRepo,
		},
		newPkg: &Package{
			vcs:        VCSModeGit,
			ImportPath: testGitRepo,
			FullPath:   filepath.Join(testEnv.dirSrc, testGitRepo),
			RemoteName: gitDefRemoteName,
			RemoteURL:  "git@github.com:shuLhan/beku_test.git",
			Version:    "c9f69fb",
			isTag:      true,
		},
		expPkg: &Package{
			vcs:         VCSModeGit,
			ImportPath:  testGitRepo,
			FullPath:    filepath.Join(testEnv.dirSrc, testGitRepo),
			RemoteName:  gitDefRemoteName,
			RemoteURL:   "git@github.com:shuLhan/beku_test.git",
			Version:     "c9f69fb",
			VersionNext: "c9f69fb",
			isTag:       false,
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		testResetOutput(t, true)

		err := c.curPkg.Update(c.newPkg)

		testResetOutput(t, false)
		stdout, stderr := testGetOutput(t)

		if err != nil {
			t.Log("stderr:", stderr)
			test.Assert(t, "err", c.expErr, err.Error(), true)
			continue
		}

		if len(stdout) > 0 {
			t.Log("stdout:", stdout)
		}

		test.Assert(t, "current pkg", *c.expPkg, *c.curPkg, true)
	}
}

func testUpdateMissingDep(t *testing.T) {
	cases := []struct {
		desc      string
		curPkg    *Package
		misPkg    *Package
		exp       bool
		expCurPkg *Package
		expMisPkg *Package
	}{{
		desc: "No missing found",
		curPkg: &Package{
			ImportPath: "curpkg",
			DepsMissing: []string{
				"a",
				"b",
			},
		},
		misPkg: &Package{
			ImportPath: "c",
		},
		expCurPkg: &Package{
			ImportPath: "curpkg",
			DepsMissing: []string{
				"a",
				"b",
			},
		},
		expMisPkg: &Package{
			ImportPath: "c",
		},
	}, {
		desc: "Missing package found",
		curPkg: &Package{
			ImportPath: "curpkg",
			DepsMissing: []string{
				"a",
				"b",
				"c",
			},
		},
		misPkg: &Package{
			ImportPath: "c",
		},
		exp: true,
		expCurPkg: &Package{
			ImportPath: "curpkg",
			DepsMissing: []string{
				"a",
				"b",
			},
			Deps: []string{
				"c",
			},
			state: packageStateDirty,
		},
		expMisPkg: &Package{
			ImportPath: "c",
			RequiredBy: []string{
				"curpkg",
			},
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		got := c.curPkg.UpdateMissingDep(c.misPkg, true)

		test.Assert(t, "return value", c.exp, got, true)
		test.Assert(t, "package", c.expCurPkg, c.curPkg, true)
	}
}

func testPackageGoClean(t *testing.T) {
	cases := []struct {
		desc      string
		pkgName   string
		pkg       *Package
		pkgBin    string
		expErr    string
		expBinErr string
	}{{
		desc:    `With package not exist`,
		pkgName: testPkgNotExist,
	}, {
		desc:      `With package exist`,
		pkg:       testGitPkgCur,
		pkgBin:    filepath.Join(testEnv.dirBin, "beku_test"),
		expBinErr: "stat %s: no such file or directory",
	}}

	var err error
	for _, c := range cases {
		t.Log(c.desc)

		if len(c.pkgName) > 0 {
			c.pkg, _ = NewPackage(c.pkgName, c.pkgName, VCSModeGit)
		}

		err = c.pkg.GoClean()
		if err != nil {
			test.Assert(t, "err", c.expErr, err, true)
			continue
		}

		if len(c.pkgBin) > 0 {
			_, err = os.Stat(c.pkgBin)
			if err != nil {
				exp := fmt.Sprintf(c.expBinErr, c.pkgBin)
				test.Assert(t, "pkgBin", exp, err.Error(), true)
			}
		}
	}
}

func testPackagePost(t *testing.T) {
	err := testGitPkgShare.Remove()
	if err != nil {
		t.Fatal(err)
	}
}

func TestPackage(t *testing.T) {
	t.Run("Remove", testPackageRemove)
	t.Run("Install", testPackageInstall)

	t.Run("GoInstall", testGoInstall)
	t.Run("IsEqual", testIsEqual)
	t.Run("addDep", testAddDep)
	t.Run("pushRequiredBy", testPushRequiredBy)
	t.Run("RemoveRequiredBy", testPackageRemoveRequiredBy)
	t.Run("load", testPackageLoad)
	t.Run("String", testPackageString)
	t.Run("Update", testUpdate)
	t.Run("UpdateMissingDep", testUpdateMissingDep)

	t.Run("GoClean", testPackageGoClean)

	t.Run("Post", testPackagePost)
}
