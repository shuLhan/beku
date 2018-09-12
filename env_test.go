// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package beku

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/shuLhan/share/lib/test"
	"github.com/shuLhan/share/lib/test/mock"
	"github.com/shuLhan/tekstus/diff"
)

func testEnvAddExclude(t *testing.T) {
	testEnv.pkgsExclude = nil

	cases := []struct {
		desc       string
		exclude    string
		expExclude []string
		expReturn  bool
	}{{
		desc: "With empty excluded",
	}, {
		desc:       "Exclude A",
		exclude:    "A",
		expExclude: []string{"A"},
		expReturn:  true,
	}, {
		desc:       "Exclude A again",
		exclude:    "A",
		expExclude: []string{"A"},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		got := testEnv.addExclude(c.exclude)

		test.Assert(t, "expExclude", c.expExclude, testEnv.pkgsExclude, true)
		test.Assert(t, "expReturn", c.expReturn, got, true)
	}
}

func testEnvExclude(t *testing.T) {
	testEnv.pkgsExclude = nil
	testEnv.pkgs = nil
	testEnv.pkgsMissing = nil

	err := testEnv.Load(testDBLoad)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		desc       string
		excludes   []string
		expExclude []string
		expPkgsLen int
		expPkg     *Package
		expMissing []string
	}{{
		desc:       "Exclude package not in DB",
		excludes:   []string{"github.com/shuLhan/notInDB"},
		expExclude: []string{"github.com/shuLhan/notInDB"},
		expPkgsLen: 49,
		expMissing: []string{
			"github.com/BurntSushi/toml",
			"gopkg.in/urfave/cli.v1",
			"gopkg.in/yaml.v2",
			"google.golang.org/appengine/log",
			"golang.org/x/crypto/ssh/terminal",
			"google.golang.org/appengine",
			"github.com/modern-go/concurrent",
			"github.com/modern-go/reflect2",
			"gopkg.in/gemnasium/logrus-airbrake-hook.v2",
			"github.com/yosssi/gohtml",
			"cloud.google.com/go/compute/metadata",
			"github.com/spf13/pflag",
		},
	}, {
		desc:     "Exclude package in missing",
		excludes: []string{"github.com/spf13/pflag"},
		expExclude: []string{
			"github.com/shuLhan/notInDB",
			"github.com/spf13/pflag",
		},
		expPkgsLen: 49,
		expPkg: &Package{
			ImportPath: "gotest.tools",
			FullPath:   filepath.Join(testEnv.dirSrc, "gotest.tools"),
			RemoteName: "origin",
			RemoteURL:  "https://github.com/gotestyourself/gotestyourself",
			Version:    "v2.0.0",
			isTag:      true,
			Deps: []string{
				"github.com/pkg/errors",
				"golang.org/x/tools",
				"github.com/google/go-cmp",
			},
			RequiredBy: []string{
				"github.com/alecthomas/gometalinter",
			},
			state:   packageStateDirty,
			vcsMode: VCSModeGit,
		},
		expMissing: []string{
			"github.com/BurntSushi/toml",
			"gopkg.in/urfave/cli.v1",
			"gopkg.in/yaml.v2",
			"google.golang.org/appengine/log",
			"golang.org/x/crypto/ssh/terminal",
			"google.golang.org/appengine",
			"github.com/modern-go/concurrent",
			"github.com/modern-go/reflect2",
			"gopkg.in/gemnasium/logrus-airbrake-hook.v2",
			"github.com/yosssi/gohtml",
			"cloud.google.com/go/compute/metadata",
		},
	}, {
		desc: "Exclude package in DB",
		excludes: []string{
			"github.com/shuLhan/beku",
		},
		expExclude: []string{
			"github.com/shuLhan/notInDB",
			"github.com/spf13/pflag",
			"github.com/shuLhan/beku",
		},
		expPkgsLen: 48,
		expPkg: &Package{
			ImportPath: "github.com/shuLhan/share",
			FullPath:   filepath.Join(testEnv.dirSrc, "github.com/shuLhan/share"),
			RemoteName: "origin",
			RemoteURL:  "git@github.com:shuLhan/share.git",
			Version:    "b2c8fd7",
			state:      packageStateLoad,
			vcsMode:    VCSModeGit,
		},
		expMissing: []string{
			"github.com/BurntSushi/toml",
			"gopkg.in/urfave/cli.v1",
			"gopkg.in/yaml.v2",
			"google.golang.org/appengine/log",
			"golang.org/x/crypto/ssh/terminal",
			"google.golang.org/appengine",
			"github.com/modern-go/concurrent",
			"github.com/modern-go/reflect2",
			"gopkg.in/gemnasium/logrus-airbrake-hook.v2",
			"github.com/yosssi/gohtml",
			"cloud.google.com/go/compute/metadata",
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		testEnv.Exclude(c.excludes)

		test.Assert(t, "exp excludes", c.expExclude, testEnv.pkgsExclude, true)
		test.Assert(t, "pkgs length", c.expPkgsLen, len(testEnv.pkgs), true)
		test.Assert(t, "pkgs missing", c.expMissing, testEnv.pkgsMissing, true)

		if c.expPkg == nil {
			continue
		}

		_, got := testEnv.GetPackageFromDB(c.expPkg.ImportPath, "")

		test.Assert(t, "expPkg", c.expPkg, got, true)
	}

	err = testEnv.Save(testDBSaveExclude)
	if err != nil {
		t.Fatal(err)
	}
}

func testEnvLoad(t *testing.T) {
	testEnv.pkgsExclude = nil
	testEnv.pkgs = nil
	testEnv.pkgsMissing = nil

	cases := []struct {
		desc   string
		file   string
		expErr string
	}{{
		desc:   "With invalid file",
		file:   "invalid",
		expErr: "open invalid: no such file or directory",
	}, {
		desc:   `With empty file (default DB file)`,
		file:   "",
		expErr: fmt.Sprintf("open %s: no such file or directory", testEnv.dbDefFile),
	}, {
		desc: `With valid file`,
		file: testDBLoad,
	}}

	var err error

	for _, c := range cases {
		t.Log(c.desc)

		err = testEnv.Load(c.file)
		if err != nil {
			test.Assert(t, "err", c.expErr, err.Error(), true)
			continue
		}
	}
}

func testEnvGetPackageFromDB(t *testing.T) {
	cases := []struct {
		desc       string
		importPath string
		remoteURL  string
		exp        *Package
	}{{
		desc:       "By import path",
		importPath: "github.com/alecthomas/gometalinter",
		exp: &Package{
			ImportPath: "github.com/alecthomas/gometalinter",
			FullPath:   filepath.Join(testEnv.dirSrc, "github.com/alecthomas/gometalinter"),
			RemoteName: "origin",
			RemoteURL:  "https://github.com/alecthomas/gometalinter",
			Version:    "0725fc6",
			vcsMode:    VCSModeGit,
			state:      packageStateLoad,
			Deps: []string{
				"github.com/stretchr/testify",
				"gotest.tools",
				"github.com/pkg/errors",
				"github.com/google/go-cmp",
			},
		},
	}, {
		desc:      "By remote URL",
		remoteURL: "https://github.com/gotestyourself/gotestyourself",
		exp: &Package{
			ImportPath: "gotest.tools",
			FullPath:   filepath.Join(testEnv.dirSrc, "gotest.tools"),
			RemoteName: "origin",
			RemoteURL:  "https://github.com/gotestyourself/gotestyourself",
			Version:    "v2.0.0",
			isTag:      true,
			vcsMode:    VCSModeGit,
			state:      packageStateLoad,
			Deps: []string{
				"github.com/pkg/errors",
				"golang.org/x/tools",
				"github.com/google/go-cmp",
			},
			DepsMissing: []string{
				"github.com/spf13/pflag",
			},
			RequiredBy: []string{
				"github.com/alecthomas/gometalinter",
			},
		},
	}, {
		desc: "By nothing",
	}}

	for _, c := range cases {
		t.Log(c.desc)

		_, got := testEnv.GetPackageFromDB(c.importPath, c.remoteURL)

		test.Assert(t, "", c.exp, got, true)
	}
}

func testEnvQuery(t *testing.T) {
	cases := []struct {
		desc      string
		pkgs      []string
		expStdout string
		expStderr string
	}{{
		desc: `Query all`,
		expStdout: `github.com/alecthomas/gometalinter  0725fc6
github.com/codegangsta/cli          8e01ec4
github.com/go-fsnotify/fsnotify     v1.4.7
github.com/go-log/log               v0.1.0
github.com/golang/lint              c5fb716
github.com/golang/protobuf          bbd03ef
github.com/google/go-github         08e68b5
github.com/google/go-querystring    53e6ce1
github.com/hashicorp/consul         v1.1.0
github.com/influxdata/chronograf    0204873d5
github.com/jguer/go-alpm            6150b61
github.com/json-iterator/go         1.1.3
github.com/kevinburke/go-bindata    2197b05
github.com/kisielk/errcheck         23699b7
github.com/kisielk/gotool           d6ce626
github.com/loadimpact/k6            1ddf285
github.com/mattn/go-colorable       efa5899
github.com/mattn/go-isatty          6ca4dbf
github.com/mgutz/ansi               9520e82
github.com/miekg/dns                906238e
github.com/mikkeloscar/aur          9050804
github.com/mikkeloscar/gopkgbuild   32274fc
github.com/mitchellh/hashstructure  2bca23e
github.com/naoina/denco             a2656d3
github.com/pborman/uuid             c65b2f8
github.com/pkg/errors               816c908
github.com/shuLhan/beku             3fb3f96
github.com/shuLhan/dsv              bbe8681
github.com/shuLhan/go-bindata       eb5746d
github.com/shuLhan/gontacts         d4786e8
github.com/shuLhan/haminer          42be4cb
github.com/shuLhan/numerus          104dd6b
github.com/shuLhan/share            b2c8fd7
github.com/shuLhan/tabula           14d5c16
github.com/shuLhan/tekstus          651065d
github.com/sirupsen/logrus          68cec9f
github.com/skratchdot/open-golang   75fb7ed
github.com/yosssi/ace               v0.0.5
github.com/zyedidia/micro           nightly
golang.org/x/net                    2491c5d
golang.org/x/oauth2                 cdc340f
golang.org/x/sys                    37707fd
golang.org/x/tools                  c1547a3f
golang.org/x/tour                   65fff99
github.com/stretchr/testify         v1.2.1
gotest.tools                        v2.0.0
github.com/google/go-cmp            v0.2.0
golang.org/x/text                   v0.3.0
github.com/ksubedi/gomove           0.2.17
`,
	}, {
		desc: `With valid and invalid packages`,
		pkgs: []string{
			"github.com/zyedidia/micro",
			"github.com/x/test",
		},
		expStdout: `github.com/zyedidia/micro           nightly
`,
	}}

	for _, c := range cases {
		t.Log(c.desc)

		mock.Reset(true)

		testEnv.Query(c.pkgs)

		mock.Reset(false)
		stdout := mock.Output()
		stderr := mock.Error()

		test.Assert(t, "expStdout", c.expStdout, stdout, true)
		test.Assert(t, "expStderr", c.expStderr, stderr, true)
	}
}

func testEnvFilterUnusedDeps(t *testing.T) {
	cases := []struct {
		importPath string
		exp        map[string]bool
	}{{
		importPath: "github.com/ksubedi/gomove",
		exp: map[string]bool{
			"golang.org/x/tools":            false,
			"github.com/ksubedi/gomove":     true,
			"github.com/codegangsta/cli":    true,
			"github.com/mattn/go-colorable": true,
			"github.com/mattn/go-isatty":    true,
			"github.com/mgutz/ansi":         true,
		},
	}, {
		importPath: "golang.org/x/text",
		exp: map[string]bool{
			"golang.org/x/text":  true,
			"golang.org/x/tools": false,
		},
	}, {
		importPath: "golang.org/x/tour",
		exp: map[string]bool{
			"golang.org/x/tour":  true,
			"golang.org/x/net":   false,
			"golang.org/x/tools": false,
		},
	}, {
		importPath: "github.com/zyedidia/micro",
		exp: map[string]bool{
			"github.com/zyedidia/micro": true,
		},
	}}

	for _, c := range cases {
		t.Log(c.importPath)

		_, pkg := testEnv.GetPackageFromDB(c.importPath, "")
		unused := make(map[string]bool)
		testEnv.filterUnusedDeps(pkg, unused)

		test.Assert(t, "exp unused", c.exp, unused, true)
	}
}

func testEnvSave(t *testing.T) {
	cases := []struct {
		desc       string
		dirty      bool
		file       string
		expStatErr string
	}{{
		desc:       "Not dirty",
		dirty:      false,
		file:       "testdata/beku.db.save",
		expStatErr: "stat testdata/beku.db.save: no such file or directory",
	}, {
		desc:       "Dirty",
		dirty:      true,
		file:       "testdata/beku.db.save",
		expStatErr: "stat testdata/beku.db.save: no such file or directory",
	}}

	for _, c := range cases {
		t.Log(c.desc)

		testEnv.dirty = c.dirty

		err := testEnv.Save(c.file)
		if err != nil {
			t.Fatal(err)
		}

		_, err = os.Stat(c.file)
		if err != nil {
			test.Assert(t, "expStatErr", c.expStatErr, err.Error(), true)
			continue
		}

		diffs, err := diff.Files(testDBLoad, c.file, diff.LevelLines)
		if err != nil {
			t.Fatal(err)
		}

		fmt.Printf("diffs: %s\n", diffs)
	}
}

func testEnvUpdateMissing(t *testing.T) {
	cases := []struct {
		newPkg     *Package
		expPkg     string
		expDeps    []string
		expMissing []string
	}{{
		newPkg: &Package{
			ImportPath: "google.golang.org/appengine",
		},
		expPkg: "github.com/google/go-github",
		expDeps: []string{
			"github.com/google/go-querystring",
			"golang.org/x/net",
			"golang.org/x/oauth2",
			"google.golang.org/appengine",
		},
		expMissing: []string{
			"golang.org/x/crypto/ssh/terminal",
		},
	}}

	for _, c := range cases {
		testEnv.updateMissing(c.newPkg, true)

		_, got := testEnv.GetPackageFromDB(c.expPkg, "")

		test.Assert(t, "expDeps", c.expDeps, got.Deps, true)
		test.Assert(t, "expMissing", c.expMissing, got.DepsMissing, true)
	}
}

func testEnvScan(t *testing.T) {
	cases := []struct {
		desc       string
		expPkgs    []*Package
		expMissing []string
	}{{
		desc: "Using testdata as GOPATH",
		expPkgs: []*Package{{
			ImportPath: testGitRepo,
			FullPath:   filepath.Join(testEnv.dirSrc, testGitRepo),
			ScanPath:   filepath.Join(testEnv.dirSrc, testGitRepo),
			RemoteName: "origin",
			RemoteURL:  "https://github.com/shuLhan/beku_test",
			Version:    "v0.2.0",
			isTag:      true,
			vcsMode:    VCSModeGit,
			state:      packageStateNew,
		}},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		err := testEnv.Scan()
		if err != nil {
			t.Fatal(err)
			continue
		}

		test.Assert(t, "expPkgs", c.expPkgs, testEnv.pkgs, true)
		test.Assert(t, "expMissing", c.expMissing, testEnv.pkgsMissing, true)
	}
}

func testEnvSync(t *testing.T) {
	cases := []struct {
		desc       string
		pkgName    string
		importPath string
		expErr     string
	}{{
		desc:   "With empty pkgname",
		expErr: ErrPackageName.Error(),
	}}

	for _, c := range cases {
		t.Log(c.desc)

		err := testEnv.Sync(c.pkgName, c.importPath)
		if err != nil {
			test.Assert(t, "err", c.expErr, err.Error(), true)
			continue
		}
	}
}

func TestEnv(t *testing.T) {
	t.Run("addExclude", testEnvAddExclude)
	t.Run("Exclude", testEnvExclude)
	t.Run("Load", testEnvLoad)
	t.Run("GetPackageFromDB", testEnvGetPackageFromDB)
	t.Run("Query", testEnvQuery)
	t.Run("filterUnusedDeps", testEnvFilterUnusedDeps)
	t.Run("Save", testEnvSave)
	t.Run("updateMissing", testEnvUpdateMissing)

	testEnv.pkgs = nil
	testEnv.pkgsMissing = nil
	testEnv.db = nil

	t.Run("Scan", testEnvScan)
	t.Run("Sync", testEnvSync)
}
