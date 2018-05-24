package beku

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/shuLhan/share/lib/ini"
)

//
// gitCheckoutVersion will set the HEAD to version stated in package.
//
func (pkg *Package) gitCheckoutVersion(version string) (err error) {
	//nolint:gas
	cmd := exec.Command("git", "checkout", "-q", version)
	fmt.Println(">>>", cmd.Args)
	cmd.Dir = pkg.FullPath
	cmd.Stdout = defStdout
	cmd.Stderr = defStderr

	err = cmd.Run()
	if err != nil {
		err = fmt.Errorf("gitCheckoutVersion: %s", err)
		return
	}

	return
}

//
// gitClone the package into "$GOPATH/src/{ImportPath}".
// If destination directory is not empty it will return an error.
//
func (pkg *Package) gitClone() (err error) {
	err = os.MkdirAll(pkg.FullPath, 0700)
	if err != nil {
		err = fmt.Errorf("gitClone: %s", err)
		return
	}

	empty := IsDirEmpty(pkg.FullPath)
	if !empty {
		err = fmt.Errorf("gitClone: "+errDirNotEmpty, pkg.FullPath)
		return
	}

	//nolint:gas
	cmd := exec.Command("git", "clone", pkg.RemoteURL, ".")
	fmt.Println(">>>", cmd.Args)
	cmd.Dir = pkg.FullPath
	cmd.Stdout = defStdout
	cmd.Stderr = defStderr

	err = cmd.Run()
	if err != nil {
		err = fmt.Errorf("gitClone: %s", err)
		return
	}

	return
}

//
// gitCompareVersion compare the version of current package with new package.
//
func (pkg *Package) gitCompareVersion(newPkg *Package) (err error) {
	//nolint:gas
	cmd := exec.Command("git", "log", "--oneline", pkg.Version+"..."+newPkg.Version)
	fmt.Println(">>>", cmd.Args)
	cmd.Dir = pkg.FullPath
	cmd.Stdout = defStdout
	cmd.Stderr = defStderr

	err = cmd.Run()
	if err != nil {
		err = fmt.Errorf("gitCompareVersion: %s", err)
		return
	}

	return
}

//
// gitFetch will fetch the latest commit from remote. On success, it will set
// the package next version to latest tag (if current package is using tag) or
// to latest commit otherwise.
//
func (pkg *Package) gitFetch() (err error) {
	//nolint:gas
	cmd := exec.Command("git", "fetch", "--all")
	fmt.Println(">>>", cmd.Args)
	cmd.Dir = pkg.FullPath
	cmd.Stdout = defStdout
	cmd.Stderr = defStderr

	err = cmd.Run()
	if err != nil {
		err = fmt.Errorf("gitFetch: %s", err)
		return
	}

	if pkg.isTag {
		pkg.VersionNext, err = pkg.gitGetTagLatest()
	} else {
		ref := filepath.Join(pkg.RemoteName, gitDefBranch)
		pkg.VersionNext, err = pkg.gitGetCommit(ref)
	}

	return
}

//
// gitGetCommit will try to get the latest commit hash from "ref"
// (origin/master).
//
func (pkg *Package) gitGetCommit(ref string) (commit string, err error) {
	//nolint:gas
	cmd := exec.Command("git", "rev-parse", "--short", ref)
	fmt.Println(">>>", cmd.Args)
	cmd.Dir = pkg.FullPath

	bcommit, err := cmd.Output()
	if err != nil {
		err = fmt.Errorf("gitGetCommit: %s", err)
		return
	}

	commit = string(bytes.TrimSpace(bcommit))

	return
}

//
// gitGetTag will try to get the current tag from HEAD.
//
func (pkg *Package) gitGetTag() (tag string, err error) {
	//nolint:gas
	cmd := exec.Command("git", "describe", "--tags", "--exact-match")
	fmt.Println(">>>", cmd.Args)
	cmd.Dir = pkg.FullPath

	btag, err := cmd.Output()
	if err != nil {
		err = fmt.Errorf("gitGetTag: %s", err)
		return
	}

	tag = string(bytes.TrimSpace(btag))

	return
}

func (pkg *Package) gitGetTagLatest() (tag string, err error) {
	//nolint:gas
	cmd := exec.Command("git", "rev-list", "--tags", "--max-count=1")
	fmt.Println(">>>", cmd.Args)
	cmd.Dir = pkg.FullPath

	bout, err := cmd.Output()
	if err != nil {
		err = fmt.Errorf("gitGetTagLatest: %s", err)
		return
	}

	out := string(bytes.TrimSpace(bout))

	//nolint:gas
	cmd = exec.Command("git", "describe", "--tags", "--abbrev=0", out)
	fmt.Println(">>>", cmd.Args)
	cmd.Dir = pkg.FullPath

	bout, err = cmd.Output()
	if err != nil {
		err = fmt.Errorf("gitGetTagLatest: %s", err)
		return
	}

	tag = string(bytes.TrimSpace(bout))

	return
}

//
// gitInstall the package into GOPATH source directory.
//
func (pkg *Package) gitInstall() (err error) {
	err = pkg.gitClone()
	if err != nil {
		err = fmt.Errorf("gitInstall: %s", err)
		return
	}

	var rev string
	if len(pkg.Version) == 0 {
		rev, err = pkg.gitGetTagLatest()
		if err == nil {
			pkg.Version = rev
			pkg.isTag = IsTagVersion(rev)
		} else {
			rev, err = pkg.gitGetCommit(gitRefHEAD)
			if err != nil {
				err = fmt.Errorf("gitInstall: %s", err)
				return
			}

			pkg.Version = rev
		}
	}

	err = pkg.gitCheckoutVersion(pkg.Version)
	if err != nil {
		err = fmt.Errorf("gitInstall: %s", err)
		return
	}

	return
}

//
// gitRemoteChange current package remote name (e.g. "origin") or URL to new
// package remote-name or url.
//
func (pkg *Package) gitRemoteChange(newPkg *Package) (err error) {
	//nolint:gas
	cmd := exec.Command("git", "remote", "remove", pkg.RemoteName)
	fmt.Println(">>>", cmd.Args)
	cmd.Dir = pkg.FullPath
	cmd.Stdout = defStdout
	cmd.Stderr = defStderr

	err = cmd.Run()
	if err != nil {
		fmt.Fprintln(defStderr, "gitRemoteChange:", err)
	}

	//nolint:gas
	cmd = exec.Command("git", "remote", "add", newPkg.RemoteName, newPkg.RemoteURL)
	fmt.Println(">>>", cmd.Args)
	cmd.Dir = pkg.FullPath
	cmd.Stdout = defStdout
	cmd.Stderr = defStderr

	err = cmd.Run()
	if err != nil {
		err = fmt.Errorf("gitRemoteChange: %s", err)
		return
	}

	return
}

//
// gitScan will scan the package version and remote URL.
//
func (pkg *Package) gitScan() (err error) {
	pkg.Version, err = pkg.gitScanVersion()
	if err != nil {
		return
	}

	err = pkg.gitScanRemote()

	return
}

func (pkg *Package) gitScanRemote() (err error) {
	gitConfig := filepath.Join(pkg.FullPath, gitDir, "config")

	gitIni, err := ini.Open(gitConfig)
	if err != nil {
		err = fmt.Errorf("gitScanRemote: %s", err)
		return
	}

	url, ok := gitIni.Get(gitCfgRemote, gitDefRemoteName, gitCfgRemoteURL)
	if !ok {
		err = fmt.Errorf("gitScanRemote: %s", ErrRemote)
		return
	}

	pkg.RemoteName = gitDefRemoteName
	pkg.RemoteURL = url

	return
}

//
// gitScanVersion will try to,
// (1) get latest tag from repository first, or if it's fail
// (2) get the commit hash at HEAD.
//
func (pkg *Package) gitScanVersion() (version string, err error) {
	// (1)
	version, err = pkg.gitGetTag()
	if err == nil {
		return
	}

	// (2)
	version, err = pkg.gitGetCommit(gitRefHEAD)
	if err != nil {
		err = ErrVersion
	}

	return
}

//
// gitUpdate will change the currrent package remote name, URL, or version
// based on new package information.
//
func (pkg *Package) gitUpdate(newPkg *Package) (err error) {
	if pkg.RemoteName != newPkg.RemoteName || pkg.RemoteURL != newPkg.RemoteURL {
		err = pkg.gitRemoteChange(newPkg)
		if err != nil {
			return
		}
	}

	if pkg.Version == newPkg.Version {
		return
	}

	err = pkg.gitCheckoutVersion(newPkg.Version)
	if err != nil {
		err = fmt.Errorf("gitUpdate: %s", err)
		return
	}

	return
}
