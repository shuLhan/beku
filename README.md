# Beku

Beku is a library and program to manage packages in `$GOPATH`.

For beku as library see the following
[![GoDoc](https://godoc.org/github.com/shuLhan/beku?status.svg)](https://godoc.org/github.com/shuLhan/beku).

For beku as program see the below documentation or at
[![GoDoc](https://godoc.org/github.com/shuLhan/beku/cmd/beku?status.svg)](https://godoc.org/github.com/shuLhan/beku/cmd/beku).


# Beku program

Beku is command line program to manage packages in $GOPATH. Beku provide
syntax like `pacman`.

Beku read and write the package database into a file named "gopath.deps".

At first execution, beku will try to open the package database in current
directory. If no file found, it will try to open
"$GOPATH/var/beku/gopath.deps". When both locations does not provide
package database, beku will scan entire "$GOPATH/src" and write the
package database into "$GOPATH/var/beku/gopath.deps".

## Sync Operation

     -S, --sync <pkg[@version]>

Synchronizes package. Given a package import path, beku will try to clone
the package into GOPATH source directory and set the package version to
latest the tag. If no tag found, it will use the latest commit on master
branch. A specific version can be set using "@version" suffix.

If package already exist, it will reset the HEAD to the version that is set
on database file.

Sync operation will not install missing dependencies.

### Options

     [--into <destination>]

This option will install the package import path into custom directory.
It is useful if you have the fork of the main package but want to install
it to the legacy directory.

### Examples

	beku -S golang.org/x/text

Download package `golang.org/x/text` into `$GOPATH/src/golang.org/x/text`,
and set their version to the latest commit on branch master.

	beku -S github.com/golang/text --into golang.org/x/text

Download package `github.com/golang/text` into
`$GOPATH/src/golang.org/x/text`, and set their version to the latest commit
on branch master.

     beku -S golang.org/x/text@v0.3.0

Download package `golang.org/x/text` into `$GOPATH/src/golang.org/x/text`
and checkout the tag `v0.3.0` as the working version.

     beku -S golang.org/x/text@5c1cf69

Download package `golang.org/x/text` into `$GOPATH/src/golang.org/x/text`
and checkout the commit `5c1cf69` as the working version.


# Known Limitations

* Only work with package hosted with Git on HTTPS or SSH.

* Tested only on package hosted on Github.


# References

[1] https://www.archlinux.org/pacman/