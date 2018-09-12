// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package beku

import (
	"fmt"
	"go/build"
	"os"
	"testing"

	"github.com/shuLhan/share/lib/test/mock"
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

	defStdout = mock.Stdout()
	defStderr = mock.Stderr()

	testEnv, err = NewEnvironment(false, false)
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
