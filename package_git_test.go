package beku

import (
	"go/build"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

const (
	testGitRepo = "github.com/shuLhan/beku_test"
)

var (
	testEnv    *Env
	gitCurPkg  *Package
	gitNewPkg  *Package
	testStdout *os.File
	testStderr *os.File
)

func testInitOutput() (err error) {
	testStdout, err = ioutil.TempFile("", "")
	if err != nil {
		return
	}

	testStderr, err = ioutil.TempFile("", "")
	if err != nil {
		return
	}

	defStdout = testStdout
	defStderr = testStderr

	return
}

func testGetOutput(t *testing.T) (stdout, stderr string) {
	bout, err := ioutil.ReadAll(defStdout)
	if err != nil {
		t.Fatal(err)
	}
	berr, err := ioutil.ReadAll(testStderr)
	if err != nil {
		t.Fatal(err)
	}

	stdout = string(bout)
	stderr = string(berr)

	return
}

func testResetOutput(t *testing.T, truncate bool) {
	_, err := testStdout.Seek(0, io.SeekStart)
	if err != nil {
		t.Fatal(err)
	}

	_, err = testStderr.Seek(0, io.SeekStart)
	if err != nil {
		t.Fatal(err)
	}

	if truncate {
		testStdout.Truncate(0)
		testStderr.Truncate(0)
	}
}

func testGitCompareVersion(t *testing.T) {
	cases := []struct {
		desc       string
		curVersion string
		newVersion string
		expErr     string
		expStdout  string
		expStderr  string
	}{{
		desc: "With empty versions",
	}, {
		desc:       "With invalid new version",
		curVersion: "v0.1.0",
		newVersion: "abcdef01",
		expErr:     "gitCompareVersion: exit status 128",
		expStderr: `fatal: ambiguous argument 'v0.1.0...abcdef01': unknown revision or path not in the working tree.
Use '--' to separate paths from revisions, like this:
'git <command> [<revision>...] -- [<file>...]'
`,
	}, {
		desc:       "With empty on new version",
		curVersion: "v0.1.0",
		expStdout: `c9f69fb Rename test to main.go
582b912 Add feature B.
ec65455 Add feature A.
`,
	}, {
		desc:       "With empty on current version #1",
		newVersion: "v0.1.0",
		expStdout: `c9f69fb Rename test to main.go
582b912 Add feature B.
ec65455 Add feature A.
`,
	}, {
		desc:       "With empty on current version #2",
		newVersion: "v0.2.0",
		expStdout: `c9f69fb Rename test to main.go
`,
	}, {
		desc:       "With empty on new version (latest tag)",
		curVersion: "v0.2.0",
		expStdout: `c9f69fb Rename test to main.go
`,
	}, {
		desc:       "With valid versions",
		curVersion: "v0.1.0",
		newVersion: "v0.2.0",
		expStdout: `582b912 Add feature B.
ec65455 Add feature A.
`,
	}, {
		desc:       "With valid versions, but reversed",
		curVersion: "v0.2.0",
		newVersion: "v0.1.0",
		expStdout: `582b912 Add feature B.
ec65455 Add feature A.
`,
	}}

	var (
		err            error
		stdout, stderr string
	)

	for _, c := range cases {
		t.Log(c.desc)

		gitCurPkg.Version = c.curVersion
		gitNewPkg.Version = c.newVersion

		err = gitCurPkg.CompareVersion(gitNewPkg)
		if err != nil {
			test.Assert(t, "err", c.expErr, err.Error(), true)
		}

		testResetOutput(t, false)
		stdout, stderr = testGetOutput(t)

		test.Assert(t, "stdout", c.expStdout, stdout, true)
		test.Assert(t, "stderr", c.expStderr, stderr, true)

		testResetOutput(t, true)
	}
}

//
// WARNING: This test require internet connection.
//
func testGitFetch(t *testing.T) {
	cases := []struct {
		desc           string
		curVersion     string
		isTag          bool
		expErr         string
		expVersionNext string
		expStdout      string
		expStderr      string
	}{{
		desc:           "With tag #1",
		curVersion:     "v0.1.0",
		isTag:          true,
		expVersionNext: "v0.2.0",
		expStdout: `Fetching origin
`,
	}, {
		desc:           "With tag #2",
		curVersion:     "v0.2.0",
		isTag:          true,
		expVersionNext: "v0.2.0",
		expStdout: `Fetching origin
`,
	}, {
		desc:           "With commit hash",
		curVersion:     "d6ad9da",
		expErr:         "gitGetCommit: exit status 128",
		expVersionNext: "c9f69fb",
		expStdout: `Fetching origin
`,
	}}

	var (
		err            error
		stdout, stderr string
	)

	for _, c := range cases {
		t.Log(c.desc)

		gitCurPkg.Version = c.curVersion
		gitCurPkg.isTag = c.isTag

		err = gitCurPkg.Fetch()
		if err != nil {
			test.Assert(t, "err", c.expErr, err.Error(), true)
		}

		testResetOutput(t, false)
		stdout, stderr = testGetOutput(t)

		test.Assert(t, "VersionNext", c.expVersionNext, gitCurPkg.VersionNext, true)
		test.Assert(t, "stdout", c.expStdout, stdout, true)
		test.Assert(t, "stderr", c.expStderr, stderr, true)

		testResetOutput(t, true)
	}
}

func testGitScan(t *testing.T) {
	cases := []struct {
		desc       string
		expErr     string
		expVersion string
		expIsTag   bool
	}{{
		desc:       "Using current package",
		expVersion: "c9f69fb",
		expIsTag:   false,
	}}

	var err error
	for _, c := range cases {
		t.Log(c.desc)

		gitCurPkg.Version = ""
		gitCurPkg.isTag = false

		err = gitCurPkg.Scan()
		if err != nil {
			test.Assert(t, "err", c.expErr, err, true)
			continue
		}

		test.Assert(t, "Version", c.expVersion, gitCurPkg.Version, true)
		test.Assert(t, "isTag", c.expIsTag, gitCurPkg.isTag, true)
	}
}

func testGitScanDeps(t *testing.T) {
	cases := []struct {
		expErr         string
		expDeps        []string
		expDepsMissing []string
		expPkgsMissing []string
	}{{
		expDepsMissing: []string{
			"github.com/shuLhan/share/lib/text",
		},
		expPkgsMissing: []string{
			"github.com/shuLhan/share/lib/text",
		},
	}}

	var err error

	for _, c := range cases {
		gitCurPkg.Deps = nil
		gitCurPkg.DepsMissing = nil

		err = gitCurPkg.ScanDeps(testEnv)
		if err != nil {
			test.Assert(t, "err", c.expErr, err.Error(), true)
		}

		test.Assert(t, "Deps", c.expDeps, gitCurPkg.Deps, true)
		test.Assert(t, "DepsMissing", c.expDepsMissing,
			gitCurPkg.DepsMissing, true)
		test.Assert(t, "env.pkgsMissing", c.expPkgsMissing,
			testEnv.pkgsMissing, true)
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

func TestGit(t *testing.T) {
	orgGOPATH := build.Default.GOPATH

	testGOPATH, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	testGOPATH += "/testdata"
	build.Default.GOPATH = testGOPATH

	defer func() {
		build.Default.GOPATH = orgGOPATH
	}()

	err = testInitOutput()
	if err != nil {
		t.Fatal(err)
	}

	testEnv, err = NewEnvironment()
	if err != nil {
		t.Fatal(err)
	}

	gitCurPkg = NewPackage(testGitRepo, testGitRepo, VCSModeGit)
	gitNewPkg = NewPackage(testGitRepo, testGitRepo, VCSModeGit)

	t.Logf("test env : %+v\n", *testEnv)
	t.Logf("gitCurPkg: %+v\n", *gitCurPkg)
	t.Logf("gitNewPkg: %+v\n", *gitNewPkg)

	t.Run("CompareVersion", testGitCompareVersion)
	t.Run("Fetch", testGitFetch)
	t.Run("Scan", testGitScan)
	t.Run("ScanDeps", testGitScanDeps)
	t.Run("addDep", testAddDep)
}
