package beku

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/shuLhan/share/lib/ini"
)

func (pkg *Package) gitBrowseCompare(fork *Package) (err error) {
	forkRemoteURL := strings.Split(fork.RemoteURL, "/")
	if len(forkRemoteURL) < 4 {
		err = fmt.Errorf("gitBrowseCompare: %s", ErrRemote)
		return
	}

	cmpURL := pkg.RemoteURL + "/compare/" + pkg.Version + "..."

	if pkg.RemoteURL != fork.RemoteURL {
		cmpURL += forkRemoteURL[3] + ":" + fork.Version
	} else {
		cmpURL += fork.Version
	}

	cmd := exec.Command("xdg-open", cmpURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		err = fmt.Errorf("gitBrowseCompare: %s", err)
		return
	}

	return
}

func (pkg *Package) gitFetch() (err error) {
	cmd := exec.Command("git", "fetch", "--all")
	cmd.Dir = pkg.FullPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

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
