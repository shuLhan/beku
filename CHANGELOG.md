# Beku v0.4.0 (2018-07-xx)

## Breaking Changes

* Remove vendor tools: gdm and govendor

govendor [1], cannot handle transitive dependencies (error when building
Consul)

Turn out gdm [2] is not vendor tool, its use GOPATH the same as beku.  Using
`gdm` will result in inconsistent build if two or more package depends on the
same dependency.  For example, package A and B depends on X, package A
depends on X v0.4.0 and package B depends on X v0.5.0, while our repository
is depends on X v0.6.0.  Running beku with the following order: `beku` on X,
`gdm` on A, and then `gdm` on B, will result in package X will be set to
v0.5.0, not v0.6.0.

* Do not use "git stash" in pre and post version checking.  Using "git stash"
introduce many problems when rebuilding package after update.

[1] https://github.com/kardianos/govendor/issues/348
[2] https://github.com/sparrc/gdm

## Enhancements

* Add newline on each freeze commands and on each package when doing reinstall
  all.
* Add option "--version", to display current command version.


# Beku v0.3.0 (2018-06-06)

## New Features

* Use vendor tools to install dependencies.  The following vendor tools is
known by beku: gdm, govendor, dep.
* Add option "-d" or "--nodeps" to disable installing dependencies
* Add common option "-V, --vendor" to work with vendor directory
* Able to install missing dependencies
* Handle custom import URL

## Bug Fixes

* Fix panic if package not found in database
* Clean non-empty directory on installation, after confirmed by user
* Save database on first time sync


# Beku v0.2.0 (2018-05-31)

## New Features

* Add operation to exclude package from database
* Add option "--noconfirm" to by pass confirmation

## Bug Fixes

* Fetch new package commits before updating version
* Fix scan on non-exist "$GOPATH/src" directory
* package: GoInstall: set default PATH if it's empty


# Beku v0.1.0 (2018-05-27)

In this version, beku can handle the following operations,

* Scanning and saving all dependencies in GOPATH (-S)
* Installing a package (-S <pkg>)
* Updating a package (-S <pkg>)
* Removing a package with or without dependencies (-R[-s] <pkg>)
* Updating all packages in database (-Su)
* Freezing all packages (-B)
