# Beku

Beku is a library and program to manage packages in user's environment (GOPATH
or vendor directory).

For beku as library see the following
[![GoDoc](https://godoc.org/github.com/shuLhan/beku?status.svg)](https://godoc.org/github.com/shuLhan/beku).

For beku as program see the below documentation or at
[![GoDoc](https://godoc.org/github.com/shuLhan/beku/cmd/beku?status.svg)](https://godoc.org/github.com/shuLhan/beku/cmd/beku).


# Beku program

Beku is command line program to manage packages in user's environment (GOPATH
or vendor directory). Beku provide syntax like `pacman`.

Beku read and write the package database into a file named "beku.db".

At first execution, beku will try to open the package database in current
directory. If no file found, it will try to open
"{prefix}/var/beku/beku.db". When both locations does not provide
package database, beku will scan entire "{prefix}/src" and write the
package database into "{prefix}/var/beku/beku.db".

## Global Options

	--noconfirm

No confirmation will be asked on any operation. Useful when running beku
inside a script.

	-d, --nodeps

Do not install any missing dependencies.  This options can be used on freeze
or sync operations.

	-V, --vendor

Operate in vendor mode.  This option used only when first scanning
(`beku -V -S`).
Any operation after that, will use the "vendor" directory in current
working directory as installation prefix.

## Freeze Operation

	-B, --freeze

Operate on the package database and user's environment. This operation will
ensure that all packages listed on database file is installed with their
specific version.  Also, all packages that are not registered will
be removed from "src" and "pkg" directories.


## Database Operation

	-D, --database

Modify the package database. This operation required one of the options
below.

### Options

	-e, --exclude <pkg ...>

Remove list of package by import path from database and add mark it as
excluded package.  Excluded package will be ignored on future operations.

### Examples

	$ beku -De github.com/shuLhan/beku

Exclude package "github.com/shuLhan/beku" from future scanning,
installation, or removal operations.


## Query Operation

	-Q, --query [pkg ...]

Query the package database.

## Remove Operation

	-R, --remove [pkg]

Remove package from environment, including source and installed binaries and
archives.

### Options

	[-s,--recursive]

Also remove all target dependencies, as long as is not required by other
packages.

### Examples

	$ beku -R github.com/shuLhan/beku

Remove package "github.com/shuLhan/beku" source in "{prefix}/src",
their installed binaries in "{prefix}/bin", and their installed archives on
"{prefix}/pkg/{GOOS}_{GOARCH}".

	$ beku -R github.com/shuLhan/beku --recursive
	$ beku -Rs github.com/shuLhan/beku

Remove package "github.com/shuLhan/beku" source in "{prefix}/src",
their installed binaries in "{prefix}/bin", their installed archives on
"{prefix}/pkg/{GOOS}_{GOARCH}", and all their dependencies.


## Sync Operation

	-S, --sync <pkg[@version]>

Synchronizes package. Given a package import path, beku will try to clone
the package into source directory and set the package version to
latest the tag. If no tag found, it will use the latest commit on master
branch. A specific version can be set using "@version" suffix.

If package already exist, it will reset the HEAD to the version that is set
on database file.

If no parameter is given, beku will do a rescan, checking for new packages.

After downloading a package, beku will check for known vendor files and run
vendor command on the package directory to install their dependencies on
package's vendor directory.  The following vendor file will be detected,

* `Gopkg.toml`, will invoke [dep](https://github.com/golang/dep)

If no vendor files found, beku will install the dependencies manually.

Installation of vendor tools is not handled by beku automatically, user must
install them manually, either by using `go get` or by using `beku -S`, e.g.

	$ beku -S https://github.com/golang/dep

The following vendor management tool will not be supported until
they can fix their issue(s),

* govendor [2], cannot handle transitive dependencies (error when building
Consul)

* gdm [3].  gdm is not vendor tool, its use GOPATH the same as beku.  Using
gdm will result in inconsistent build if two or more package depends on the
same dependency.  For example, package A and B depends on X, package A
depends on X v0.4.0 and package B depends on X v0.5.0, while our repository
is depends on X v0.6.0.  Running beku with the following order: our repo,
`gdm` on A, and then `gdm` on B, will result in package X will be set to
v0.5.0, not v0.6.0.

### Options

	[--into <destination>]

This option will install the package import path into custom directory.
It is useful if you have the fork of the main package but want to install
it to the legacy directory.

	[-u,--update]

Fetch new tag or commit from remote repository. User will be asked for
confirmation before upgrade.

### Examples

	$ beku -S golang.org/x/text

Download package `golang.org/x/text` into `{prefix}/src/golang.org/x/text`,
and set their version to the latest commit on branch master.

	$ beku -S github.com/golang/text --into golang.org/x/text

Download package `github.com/golang/text` into
`{prefix}/src/golang.org/x/text`, and set their version to the latest commit
on branch master.

	$ beku -S golang.org/x/text@v0.3.0

Download package `golang.org/x/text` into `{prefix}/src/golang.org/x/text`
and checkout the tag `v0.3.0` as the working version.

	$ beku -S golang.org/x/text@5c1cf69

Download package `golang.org/x/text` into `{prefix}/src/golang.org/x/text`
and checkout the commit `5c1cf69` as the working version.

	$ beku -Su

Update all packages in database to new tag or commits with approval from
user.


# Known Limitations

* Only work with package hosted with Git on HTTPS or SSH.

* Tested only on Git v2.17 or greater


# References

[1] https://www.archlinux.org/pacman/

[2] https://github.com/kardianos/govendor/issues/348

[3] https://github.com/sparrc/gdm
