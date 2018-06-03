// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package beku

import (
	"fmt"
	"go/build"
	"io"
	"io/ioutil"
	"os"
	"testing"
)

const (
	testDBLoad         = "testdata/beku.db"
	testDBSaveExclude  = "testdata/beku.db.exclude"
	testGitRepo        = "github.com/shuLhan/beku_test"
	testGitRepoVersion = "c9f69fb"
	testGitRepoShare   = "github.com/shuLhan/share"
	testPkgNotExist    = "github.com/shuLhan/notexist"
)

var (
	testEnv         *Env
	testGitPkgCur   *Package
	testGitPkgNew   *Package
	testGitPkgShare *Package
	testStdout      *os.File
	testStderr      *os.File
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

func testPrintOutput(t *testing.T) {
	testResetOutput(t, false)
	stdout, stderr := testGetOutput(t)
	t.Log(">>> stdout:\n", stdout)
	t.Log(">>> stderr:\n", stderr)
}

func TestMain(m *testing.M) {
	orgGOPATH := build.Default.GOPATH

	testGOPATH, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	testGOPATH += "/testdata"
	build.Default.GOPATH = testGOPATH

	defer func() {
		build.Default.GOPATH = orgGOPATH
	}()

	err = testInitOutput()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	testEnv, err = NewEnvironment(false)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	testGitPkgCur, _ = NewPackage(testEnv, testGitRepo, testGitRepo)
	testGitPkgNew, _ = NewPackage(testEnv, testGitRepo, testGitRepo)
	testGitPkgShare, _ = NewPackage(testEnv, testGitRepoShare, testGitRepoShare)

	// Always set the git test repo to latest version.
	testEnv.NoConfirm = true
	err = testEnv.Sync(testGitRepo, testGitRepo)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	testEnv.NoConfirm = false

	fmt.Printf("test env : %+v\n", *testEnv)
	fmt.Printf("testGitPkgCur: %+v\n", *testGitPkgCur)
	fmt.Printf("testGitPkgNew: %+v\n", *testGitPkgNew)

	os.Exit(m.Run())
}
