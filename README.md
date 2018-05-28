# Beku

Beku is a library and program to manage packages in `$GOPATH`.

For beku as library see the following
[![GoDoc](https://godoc.org/github.com/shuLhan/beku?status.svg)](https://godoc.org/github.com/shuLhan/beku).

For beku as program see the below documentation or at
[![GoDoc](https://godoc.org/github.com/shuLhan/beku/cmd/beku?status.svg)](https://godoc.org/github.com/shuLhan/beku/cmd/beku).


# Beku program

Beku is command line program to manage packages in $GOPATH. Beku provide
syntax like `pacman`.

Beku read and write the package database into a file named "beku.db".

At first execution, beku will try to open the package database in current
directory. If no file found, it will try to open
"$GOPATH/var/beku/beku.db". When both locations does not provide
package database, beku will scan entire "$GOPATH/src" and write the
package database into "$GOPATH/var/beku/beku.db".


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
installation, or removal.


## Freeze Operation

     -B, --freeze

Operate on the package database and GOPATH. This operation will ensure that
all packages listed on database file is installed with their specific
version on GOPATH.  Also, all packages that are not registered will be
removed from GOPATH "src" and "pkg" directories.


## Query Operation

	-Q, --query [pkg ...]

Query the package database.

## Remove Operation

	-R, --remove [pkg]

Remove package from GOPATH, including source and installed binaries and
archives.

### Options

	[-s,--recursive]

Also remove all target dependencies, as long as is not required by other
packages.

### Examples

	$ beku -R github.com/shuLhan/beku

Remove package "github.com/shuLhan/beku" source in "$GOPATH/src",
their installed binaries in "$GOPATH/bin", and their installed archives on
"$GOPATH/pkg/{GOOS}_{GOARCH}".

	$ beku -R github.com/shuLhan/beku --recursive
	$ beku -Rs github.com/shuLhan/beku

Remove package "github.com/shuLhan/beku" source in "$GOPATH/src",
their installed binaries in "$GOPATH/bin", their installed archives on
"$GOPATH/pkg/{GOOS}_{GOARCH}", and all their dependencies.


## Sync Operation

	-S, --sync <pkg[@version]>

Synchronizes package. Given a package import path, beku will try to clone
the package into GOPATH source directory and set the package version to
latest the tag. If no tag found, it will use the latest commit on master
branch. A specific version can be set using "@version" suffix.

If package already exist, it will reset the HEAD to the version that is set
on database file.

Sync operation will not install missing dependencies.

If no parameter is given, beku will rescan GOPATH, checking for new
packages.

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

Download package `golang.org/x/text` into `$GOPATH/src/golang.org/x/text`,
and set their version to the latest commit on branch master.

	$ beku -S github.com/golang/text --into golang.org/x/text

Download package `github.com/golang/text` into
`$GOPATH/src/golang.org/x/text`, and set their version to the latest commit
on branch master.

	$ beku -S golang.org/x/text@v0.3.0

Download package `golang.org/x/text` into `$GOPATH/src/golang.org/x/text`
and checkout the tag `v0.3.0` as the working version.

	$ beku -S golang.org/x/text@5c1cf69

Download package `golang.org/x/text` into `$GOPATH/src/golang.org/x/text`
and checkout the commit `5c1cf69` as the working version.

        $ beku -Su

Update all packages in database to new tag or commits with approval from
user.


# Known Limitations

* Only work with package hosted with Git on HTTPS or SSH.

* Tested only on package hosted on Github.

* Tested only on Git v2.17 or greater


# References

[1] https://www.archlinux.org/pacman/
