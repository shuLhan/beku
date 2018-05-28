package beku

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

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

		testGitPkgCur.Version = c.curVersion
		testGitPkgNew.Version = c.newVersion

		testResetOutput(t, true)

		err = testGitPkgCur.CompareVersion(testGitPkgNew)
		if err != nil {
			test.Assert(t, "err", c.expErr, err.Error(), true)
		}

		testResetOutput(t, false)
		stdout, stderr = testGetOutput(t)

		test.Assert(t, "stdout", c.expStdout, stdout, true)
		test.Assert(t, "stderr", c.expStderr, stderr, true)
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

		testGitPkgCur.Version = c.curVersion
		testGitPkgCur.isTag = c.isTag

		testResetOutput(t, true)

		err = testGitPkgCur.Fetch()
		if err != nil {
			test.Assert(t, "err", c.expErr, err.Error(), true)
		}

		testResetOutput(t, false)
		stdout, stderr = testGetOutput(t)

		test.Assert(t, "VersionNext", c.expVersionNext, testGitPkgCur.VersionNext, true)
		test.Assert(t, "stdout", c.expStdout, stdout, true)
		test.Assert(t, "stderr", c.expStderr, stderr, true)
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

func TestPackageGit(t *testing.T) {
	t.Run("CompareVersion", testGitCompareVersion)
	t.Run("Fetch", testGitFetch)
	t.Run("Scan", testGitScan)
	t.Run("ScanDeps", testGitScanDeps)
}
