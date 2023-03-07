// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package beku

import (
	"fmt"
	"go/build"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/shuLhan/share/lib/debug"
	"github.com/shuLhan/share/lib/test/mock"
)

const (
	testDBLoad        = "testdata/beku.db"
	testDBSaveExclude = "testdata/beku.db.exclude"
	testGitRepo       = "github.com/shuLhan/beku_test"
	testPkgNotExist   = "github.com/shuLhan/notexist"
)

var (
	testEnv           *Env
	testGitPkgCur     *Package
	testGitPkgNew     *Package
	testGitPkgInstall *Package

	testGitRepoSrcLocal = `/testdata/beku_test.git`
)

func TestMain(m *testing.M) {
	orgGOPATH := build.Default.GOPATH

	testGOPATH, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	testGOPATH = filepath.Join(testGOPATH, `testdata`)
	build.Default.GOPATH = testGOPATH

	defer func() {
		build.Default.GOPATH = orgGOPATH
	}()

	defStdout = mock.Stdout()
	defStderr = mock.Stderr()

	testEnv, err = NewEnvironment(false)
	if err != nil {
		log.Fatal(err)
	}

	err = os.RemoveAll(testEnv.dirSrc)
	if err != nil {
		log.Fatal(err)
	}

	testGitPkgCur, _ = NewPackage(testEnv.dirSrc, testGitRepo, testGitRepo)
	testGitPkgNew, _ = NewPackage(testEnv.dirSrc, testGitRepo, testGitRepo)
	testGitPkgInstall, _ = NewPackage(testEnv.dirSrc, testGitRepo, testGitRepo)

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	testGitRepoSrcLocal = "file://" + wd + testGitRepoSrcLocal
	testGitPkgInstall.RemoteURL = testGitRepoSrcLocal

	if debug.Value >= 1 {
		fmt.Printf("test env : %+v\n", *testEnv)
		fmt.Printf("testGitPkgCur: %+v\n", *testGitPkgCur)
		fmt.Printf("testGitPkgNew: %+v\n", *testGitPkgNew)
	}

	err = testGitPkgInstall.Install()
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(m.Run())
}
