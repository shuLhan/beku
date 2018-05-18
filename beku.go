package beku

import (
	"errors"
	"os"
)

const (
	// DefDBName define default database name, where the dependencies will
	// be saved and loaded.
	DefDBName = "gopath.deps"
)

const (
	dbgLinkDep  = "linking dep"
	dbgMissDep  = "missing dep"
	dbgSkipSelf = "skip self dep"
	dbgSkipStd  = "skip std dep"

	dirDB       = "/var/beku"
	dirBin      = "bin"
	dirPkg      = "pkg"
	dirSrc      = "src"
	dirTestdata = "testdata"
	dirVendor   = "vendor"

	envDEBUG = "BEKU_DEBUG"

	msgUpdateProceed = "Proceed with update?"
	msgUpdateView    = "View commit logs?"

	prefixTag = 'v'

	sepImport        = "/"
	sepImportVersion = '@'
	sepVersion       = '.'
)

// List of error messages.
var (
	ErrGOPATH = errors.New("GOPATH is not defined")
	ErrGOROOT = errors.New("GOROOT is not defined")

	// ErrVersion define an error when directory have VCS metadata (e.g.
	// `.git` directory) but did not have any tag or commit.
	ErrVersion = errors.New("No tag or commit found")

	// ErrRemote define an error when package remote URL is empty or
	// invalid.
	ErrRemote = errors.New("Empty or invalid remote URL found")

	// ErrPackageName define an error if package name is empty or invalid.
	ErrPackageName = errors.New("Empty or invalid package name")

	errDBPackageName = "missing package name, line %d at %s"
)

var (
	// Debug level for this package. Set from environment variable
	// BEKU_DEBUG.
	Debug debugMode
)

var (
	defStdout = os.Stdout
	defStderr = os.Stderr

	sectionPackage = "package"

	keyDeps        = "deps"
	keyDepsMissing = "missing"
	keyRemoteName  = "remote-name"
	keyRemoteURL   = "remote-url"
	keyRequiredBy  = "required-by"
	keyVCSMode     = "vcs"
	keyVersion     = "version"

	gitCfgRemote     = "remote"
	gitCfgRemoteURL  = "url"
	gitDefBranch     = "master"
	gitDefRemoteName = "origin"
	gitDir           = ".git"
	gitRefHEAD       = "HEAD"
)
