// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package beku

import (
	"errors"
	"fmt"
	"go/build"
	"os"
	"path/filepath"
	"strings"

	libhttp "git.sr.ht/~shulhan/pakakeh.go/lib/http"
)

const (
	// DefDBName define default database name, where the dependencies will
	// be saved and loaded.
	DefDBName = "beku.db"
)

const (
	defPATH = "/bin:/usr/bin:/usr/sbin:/usr/local/bin:/usr/local/sbin"

	dirDB       = "var/beku"
	dirBin      = "bin"
	dirPkg      = "pkg"
	dirSrc      = "src"
	dirTestdata = "testdata"
	dirVendor   = "vendor"

	envPATH = "PATH"

	msgCleanDir      = "Clean destination directory?"
	msgContinue      = "Continue?"
	msgUpdateProceed = "Proceed with update?"
	msgUpdateView    = "View commit logs?"

	prefixTag = 'v'

	sepImport        = "/"
	sepImportVersion = '@'
	sepVersion       = '.'
)

const (
	sectionBeku    = "beku"
	sectionPackage = "package"

	keyExclude = "exclude"

	keyDeps         = "deps"
	keyDepsMissing  = "missing"
	keyRemoteName   = "remote-name"
	keyRemoteURL    = "remote-url"
	keyRemoteBranch = "remote-branch"
	keyRequiredBy   = "required-by"
	keyVCSMode      = "vcs"
	keyVersion      = "version"

	gitDefBranch     = "master"
	gitDefRemoteName = "origin"
	gitDir           = ".git"
)

// Debug flag to debug running command.
// Set from environment variable BEKU_DEBUG.
var Debug int64 = 0

// List of error messages.
var (
	ErrGOROOT = errors.New("GOROOT is not defined")

	// ErrVersion define an error when directory have VCS metadata (e.g.
	// `.git` directory) but did not have any tag or commit.
	ErrVersion = errors.New("no tag or commit found")

	// ErrRemote define an error when package remote URL is empty or
	// invalid.
	ErrRemote = errors.New("empty or invalid remote URL found")

	// ErrPackageName define an error if package name is empty or invalid.
	ErrPackageName = errors.New("empty or invalid package name")

	errDirNotEmpty = "directory %s is not empty"
	errExcluded    = "package '%s' is in excluded list\n"
	errVCS         = "unknown VCS mode %s"
)

var (
	defStdout = os.Stdout //nolint: gochecknoglobals
	defStderr = os.Stderr //nolint: gochecknoglobals
)

type Beku struct {
	env   *Env
	httpc *libhttp.Client
}

func NewBeku(env *Env) (beku *Beku) {
	beku = &Beku{
		env: env,
	}
	return beku
}

// Install the Go binary or a module.
func (beku *Beku) Install(module string) (err error) {
	module = strings.ToLower(module)

	var (
		nameVersion = strings.Split(module, "@")
		modName     = nameVersion[0]
		modVersion  string
	)

	if len(nameVersion) > 1 {
		modVersion = nameVersion[1]
	}

	if modName == "go" {
		if len(modVersion) == 0 {
			return beku.installGoFromSource()
		}
		return beku.installGo(modVersion)
	}

	return nil
}

// checkInstalledGo check if specific Go tools version already installed.
// It will return nil if it not exist.
func (beku *Beku) checkInstalledGo(version string) (err error) {
	pathGo := filepath.Join(beku.env.prefix, "beku", "share", "go"+version)
	_, err = os.Stat(pathGo)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return fmt.Errorf("%s already installed", pathGo)
}

func (beku *Beku) downloadGo(version string) (err error) {
	var (
		logp        = "downloadGo"
		compression = "tar.gz"
	)

	if build.Default.GOOS == "windows" {
		compression = "zip"
	}

	goDownload := fmt.Sprintf("go%s.%s-%s.%s", version,
		build.Default.GOOS, build.Default.GOARCH,
		compression)

	pathCache := filepath.Join(beku.env.prefix, "beku", "cache", goDownload)

	fi, err := os.Stat(pathCache)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("%s: %w", logp, err)
		}
		err = nil
	}
	if fi != nil {
		// The downloaded file already exist on cache directory.
		return nil
	}

	fout, err := os.Create(pathCache)
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}

	beku.initHttpClient()

	downloadReq := libhttp.DownloadRequest{
		ClientRequest: libhttp.ClientRequest{
			Path: "/dl/" + goDownload,
		},
		Output: fout,
	}

	_, err = beku.httpc.Download(downloadReq)
	if err != nil {
		errClose := fout.Close()
		if errClose != nil {
			err = fmt.Errorf("%s: %w: %s", logp, err, errClose)
		} else {
			err = fmt.Errorf("%s: %w", logp, err)
		}
		return err
	}

	err = fout.Close()
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}

	return nil
}

func (beku *Beku) initHttpClient() {
	if beku.httpc != nil {
		return
	}
	opts := libhttp.ClientOptions{
		ServerURL: "https://go.dev",
	}
	beku.httpc = libhttp.NewClient(opts)
}

// installGo install specific Go tools version by downloading the binary from
// official server and extract it into $GOPATH/beku/share/go{version}.
func (beku *Beku) installGo(version string) (err error) {
	var (
		logp = "installGo " + version
	)

	err = beku.checkInstalledGo(version)
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}

	err = beku.downloadGo(version)
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}

	return nil
}

func (beku *Beku) installGoFromSource() (err error) {
	return nil
}
