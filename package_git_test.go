package beku

import (
	"go/build"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

var (
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
		expStdout: `582b912 Add feature B.
ec65455 Add feature A.
`,
	}, {
		desc:       "With empty on current version #1",
		newVersion: "v0.1.0",
		expStdout: `582b912 Add feature B.
ec65455 Add feature A.
`,
	}, {
		desc:       "With empty on current version #2",
		newVersion: "v0.2.0",
	}, {
		desc:       "With empty on new version (latest tag)",
		curVersion: "v0.2.0",
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
		bout, berr     []byte
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

		bout, err = ioutil.ReadAll(defStdout)
		if err != nil {
			t.Fatal(err)
		}
		berr, err = ioutil.ReadAll(testStderr)
		if err != nil {
			t.Fatal(err)
		}

		stdout = string(bout)
		stderr = string(berr)

		test.Assert(t, "stdout", c.expStdout, stdout, true)
		test.Assert(t, "stderr", c.expStderr, stderr, true)

		testResetOutput(t, true)
	}
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

	gitCurPkg = NewPackage("git", "git", VCSModeGit)
	gitNewPkg = NewPackage("git", "git", VCSModeGit)

	t.Logf("gitCurPkg: %+v\n", *gitCurPkg)
	t.Logf("gitNewPkg: %+v\n", *gitNewPkg)

	t.Run("CompareVersion", testGitCompareVersion)
}
