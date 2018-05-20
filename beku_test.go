package beku

import (
	"go/build"
	"io"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

const (
	testGitRepo      = "github.com/shuLhan/beku_test"
	testGitRepoShare = "github.com/shuLhan/share"
	testPkgNotExist  = "github.com/shuLhan/notexist"
)

var (
	testEnv     *Env
	gitCurPkg   *Package
	gitNewPkg   *Package
	gitPkgShare *Package
	testStdout  *os.File
	testStderr  *os.File
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

func TestMain(t *testing.M) {
	orgGOPATH := build.Default.GOPATH

	testGOPATH, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	testGOPATH += "/testdata"
	build.Default.GOPATH = testGOPATH

	defer func() {
		build.Default.GOPATH = orgGOPATH
	}()

	err = testInitOutput()
	if err != nil {
		log.Fatal(err)
	}

	testEnv, err = NewEnvironment()
	if err != nil {
		log.Fatal(err)
	}

	gitCurPkg = NewPackage(testGitRepo, testGitRepo, VCSModeGit)
	gitNewPkg = NewPackage(testGitRepo, testGitRepo, VCSModeGit)
	gitPkgShare = NewPackage(testGitRepoShare, testGitRepoShare, VCSModeGit)

	log.Printf("test env : %+v\n", *testEnv)
	log.Printf("gitCurPkg: %+v\n", *gitCurPkg)
	log.Printf("gitNewPkg: %+v\n", *gitNewPkg)

	os.Exit(t.Run())
}
