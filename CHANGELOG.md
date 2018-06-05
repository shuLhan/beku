# Beku v0.3.0 (2018-05-06)

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
