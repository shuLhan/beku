# Beku v0.2.0

## New Features

* Add operation to exclude package from databasea
* Add option "--noconfirm" to by pass confirmation

## Bug Fixes

* Fetch new package commits before updating version
* Fix scan on non-exist "$GOPATH/src" directory
* package: GoInstall: set default PATH if it's empty


# Beku v0.1.0

In this version, beku can handle the following operations,

* Scanning and saving all dependencies in GOPATH (-S)
* Installing a package (-S <pkg>)
* Updating a package (-S <pkg>)
* Removing a package with or without dependencies (-R[-s] <pkg>)
* Updating all packages in database (-Su)
* Freezing all packages (-B)
