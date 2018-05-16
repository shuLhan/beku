package beku

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/shuLhan/share/lib/ini"
)

//
// gitCompareVersion compare the version of current package with new package.
//
func (pkg *Package) gitCompareVersion(newPkg *Package) (err error) {
	cmd := exec.Command("git", "log", "--oneline", pkg.Version+"..."+newPkg.Version)
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
	cmd := exec.Command("git", "fetch", "--all")
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
		ref := pkg.RemoteName + "/" + gitDefBranch
		pkg.VersionNext, err = pkg.gitGetCommit(ref)
	}

	return
}

//
// gitGetCommit will try to get the latest commit hash from "ref"
// (origin/master).
//
func (pkg *Package) gitGetCommit(ref string) (commit string, err error) {
	cmd := exec.Command("git", "rev-parse", "--short", ref)
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
	cmd := exec.Command("git", "describe", "--tags", "--exact-match")
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
	cmd := exec.Command("git", "rev-list", "--tags", "--max-count=1")
	cmd.Dir = pkg.FullPath

	bout, err := cmd.Output()
	if err != nil {
		err = fmt.Errorf("gitGetTagLatest: %s", err)
		return
	}

	out := string(bytes.TrimSpace(bout))

	cmd = exec.Command("git", "describe", "--tags", "--abbrev=0", out)
	cmd.Dir = pkg.FullPath

	bout, err = cmd.Output()
	if err != nil {
		err = fmt.Errorf("gitGetTagLatest: %s", err)
		return
	}

	tag = string(bytes.TrimSpace(bout))

	return
}

func (pkg *Package) gitRemoteChange(newPkg *Package) (err error) {
	cmd := exec.Command("git", "remote", "remove", pkg.RemoteName)
	cmd.Dir = pkg.FullPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		log.Println("gitRemoteChange:", err)
	}

	cmd = exec.Command("git", "remote", "add", newPkg.RemoteName, newPkg.RemoteURL)
	cmd.Dir = pkg.FullPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

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
	gitConfig := pkg.FullPath + "/" + gitDir + "/config"

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

func (pkg *Package) gitUpdate(newPkg *Package) (err error) {
	if pkg.RemoteName != newPkg.RemoteName || pkg.RemoteURL != newPkg.RemoteURL {
		err = pkg.gitRemoteChange(newPkg)
		if err != nil {
			return
		}
	}

	cmd := exec.Command("git", "checkout", "-q", newPkg.Version)
	cmd.Dir = newPkg.FullPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		err = fmt.Errorf("gitUpdate: %s", err)
		return
	}

	return
}
