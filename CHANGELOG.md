#  Beku v0.7.1 (2022-01-30)

This is the last release for v0 of beku.
The next release will pivot the beku program to maintain Go versions.

## Chores

*  go.mod: update all dependencies

   This changes affect code that use lib/test.Assert().

*  all: simplify running linter

   Instead of enabling all linters, run the default linter provided by
   the latest golangci-lint.

*  all: replace beku_test in testdata with gitsubmodule

   Previously, the beku_test is bare clone of git repository stored as is.
   This cause an initial clone of the beku repository itself does not
   recognize it as git repository, which make the test fail:

     go test -v -coverprofile=cover.out ./... || rm -f cover.out
     fatal: '/home/ms/go/src/git.sr.ht/~shulhan/beku/testdata/beku_test.git' does not appear to be a git repository
     fatal: Could not read from remote repository.

     Please make sure you have the correct access rights
     and the repository exists.
     2022/01/30 10:56:19 gitInstall: Clone: exit status 128
     FAIL    github.com/shuLhan/beku 0.008s

   This changes make the testdata/beku_test.git as git submodule, to make
   the test run well.


#  Beku v0.7.0 (2019-06-28)

##  Enhancements

-  Do not auto cleanup unused repositories.

   Previously, we remove all unused repositories from $GOPATH, which caused
   some important repositories that we forgot to track get removed during
   freeze operation.

   This commit remove the auto cleanup and only print the unused repositories
   to the screen.

-  Add $GOCACHE and $HOME to environments on GoInstall

   The latest Go release will require $GOCACHE or $HOME environment variable
   upon running "go install" command.


#  Beku v0.6.0 (2019-03-30)

##  New Features

-  Make Go command works with Go v1.12 and later.
   This means turning off GO111MODULE when running Go command.

-  Install package if missing when issuing SyncAll

##  Enhancement

-  Use copy of "golang.org/x/tools/go/vcs".

   The indirect dependencies of package "golang.org/x/tools/go/vcs" is
   overwhelming.  Using "go get" on this package alone will pull up all
   dependency of "golang.org/x/tools".

   To minimize unneeded download of unneeded packages we copy the package
   vcs to our own repository including their license file.


#  Beku v0.5.2 (2018-12-14)

##  Enhancement

-  No need to reinstall all packages after freezing

##  Bug Fix

-  git: set package remote URL according to value in database


#  Beku v0.5.1 (2018-11-02)

##  Bug Fix

-  Fix sync all that cause version set to true


# Beku v0.5.0 (2018-11-01)

## Enhancements

- Refactoring test to clone from local directory

- Get and save package remote branch in database
  Some package does not have "master" branch. This will minimize parsing
  and filter operation to get default branch before checking out revision.

- Scan package only if its not exist on local system.
  This will minimize freeze operations, removing unneeded fetching revision
  (tag/commit) and parsing remote URL.

- Move all commons functions to shared package
  "github.com/shuLhan/share/lib/{git,io}"

## Bug Fixes

- cmd/beku: fix parsing multiple subcommand on Sync
  Sync operation should accept both update and no dependency options in one
  line as in "-Sud".

- Scan: Update package version only if current and new package both are tag

- env: fix get package from database that return first match by prefix
  In case two packages have the same prefix, for example "a" and "a-a",
  the GetPackageFromDB will always return "a" when the parameter importPath
  is "a-a".

- Fix sync "--into" command

# Beku v0.4.0 (2018-09-04)

## Breaking Changes

- Remove vendor tools: gdm and govendor

govendor [1], cannot handle transitive dependencies (error when building
Consul)

Turn out gdm [2] is not vendor tool, its use GOPATH the same as beku. Using
`gdm` will result in inconsistent build if two or more package depends on the
same dependency. For example, package A and B depends on X, package A
depends on X v0.4.0 and package B depends on X v0.5.0, while our repository
is depends on X v0.6.0. Running beku with the following order: `beku` on X,
`gdm` on A, and then `gdm` on B, will result in package X will be set to
v0.5.0, not v0.6.0.

- Do not use "git stash" in pre and post version checking. Using "git stash"
  introduce many problems when rebuilding package after update.

[1] https://github.com/kardianos/govendor/issues/348
[2] https://github.com/sparrc/gdm

## Enhancements

- Add newline on each freeze commands and on each package when doing reinstall
  all.
- Add option "--version", to display current command version.

# Beku v0.3.0 (2018-06-06)

## New Features

- Use vendor tools to install dependencies. The following vendor tools is
  known by beku: gdm, govendor, dep.
- Add option "-d" or "--nodeps" to disable installing dependencies
- Add common option "-V, --vendor" to work with vendor directory
- Able to install missing dependencies
- Handle custom import URL

## Bug Fixes

- Fix panic if package not found in database
- Clean non-empty directory on installation, after confirmed by user
- Save database on first time sync

# Beku v0.2.0 (2018-05-31)

## New Features

- Add operation to exclude package from database
- Add option "--noconfirm" to by pass confirmation

## Bug Fixes

- Fetch new package commits before updating version
- Fix scan on non-exist "$GOPATH/src" directory
- package: GoInstall: set default PATH if it's empty

# Beku v0.1.0 (2018-05-27)

In this version, beku can handle the following operations,

- Scanning and saving all dependencies in GOPATH (-S)
- Installing a package (-S <pkg>)
- Updating a package (-S <pkg>)
- Removing a package with or without dependencies (-R[-s] <pkg>)
- Updating all packages in database (-Su)
- Freezing all packages (-B)
