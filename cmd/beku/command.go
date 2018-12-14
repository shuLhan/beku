// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/shuLhan/beku"
	"github.com/shuLhan/share/lib/debug"
)

var (
	errInvalidOptions  = errors.New("error: invalid options")
	errMultiOperations = errors.New("error: only at operation may be used at a time")
	errNoDB            = errors.New("error: no database found")
	errNoOperation     = errors.New("error: no operation specified")
	errNoTarget        = errors.New("error: no package specified")
)

const (
	flagOperationHelp     = "Show the short usage."
	flagOperationDatabase = "Operate on the package database."
	flagOperationFreeze   = "Install all packages on database."
	flagOperationQuery    = "Query the package database."
	flagOperationRemove   = "Remove package."
	flagOperationSync     = "Synchronize package. If no package is given, it will do rescan."
	flagOperationVersion  = "Print beku version."

	flagOptionExclude   = "Exclude package from further operation"
	flagOptionNoConfirm = "No confirmation will be asked on any operation."
	flagOptionNoDeps    = "Do not install any missing dependencies."
	flagOptionRecursive = "Remove package including their dependencies."
	flagOptionSyncInto  = "Download package into `directory`."
	flagOptionUpdate    = "Update all packages to latest version."
	flagOptionVendor    = "Operate in vendor mode."
)

type command struct {
	op        operation
	env       *beku.Env
	pkgs      []string
	syncInto  string
	firstTime bool
	noConfirm bool
	noDeps    bool
	vendor    bool
}

func (cmd *command) usage() {
	help := `usage: beku <operation> [...]
common options:
	--noconfirm
		` + flagOptionNoConfirm + `
	-d,--nodeps
		` + flagOptionNoDeps + `
	-V,--vendor
		` + flagOptionVendor + `
operations:
	beku {-h|--help}
		` + flagOperationHelp + `

	beku {--version}
		` + flagOperationVersion + `

	beku {-B|--freeze}
		` + flagOperationFreeze + `

	beku {-D|--database}
		` + flagOperationDatabase + `

	options:
		[-e|--exclude]
			` + flagOptionExclude + `

	beku {-Q|--query} [pkg ...]
		` + flagOperationQuery + `

	beku {-R|--remove} <pkg> [options]
		` + flagOperationRemove + `

	options:
		[-s|--recursive]
			` + flagOptionRecursive + `

	beku {-S|--sync} <pkg[@version]> [options]
		` + flagOperationSync + `

	options:
		[-u|--update]
			` + flagOptionUpdate + `

		[--into <directory>]
			` + flagOptionSyncInto + `
`

	fmt.Fprint(os.Stderr, help)
}

func (cmd *command) version() {
	fmt.Printf("beku v%d.%d.%d%s\n", verMajor, verMinor, verPatch, verMetadata)
}

func (cmd *command) parseDatabaseFlags(arg string) (operation, error) {
	if len(arg) == 0 {
		return opNone, nil
	}

	switch arg[0] {
	case 'e':
		return opExclude, nil
	}

	return opNone, errInvalidOptions
}

func (cmd *command) parseFreezeFlags(arg string) error {
	if len(arg) == 0 {
		return nil
	}

	switch arg[0] {
	case 'd':
		cmd.noDeps = true
		return nil
	}

	return errInvalidOptions
}

func (cmd *command) parseSyncFlags(arg string) (operation, error) {
	if len(arg) == 0 {
		return opNone, nil
	}

	var op operation

	for _, c := range arg {
		switch c {
		case 'u':
			op |= opUpdate
		case 'd':
			cmd.noDeps = true
		default:
			return opNone, errInvalidOptions
		}
	}

	return op, nil
}

func (cmd *command) parseRemoveFlags(arg string) (operation, error) {
	if len(arg) == 0 {
		return opNone, nil
	}

	var op operation

	switch arg[0] {
	case 's':
		op = opRecursive
		return op, nil
	}

	return opNone, errInvalidOptions
}

func (cmd *command) parseShortFlags(arg string) (operation, error) {
	if len(arg) == 0 {
		return opNone, errInvalidOptions
	}

	var (
		op  operation
		err error
	)

	switch arg[0] {
	case 'd':
		if len(arg) > 1 {
			return opNone, errInvalidOptions
		}
		cmd.noDeps = true
	case 's':
		if len(arg) > 1 {
			return opNone, errInvalidOptions
		}
		op = opRecursive
	case 'h':
		if len(arg) > 1 {
			return opNone, errInvalidOptions
		}
		op = opHelp
	case 'B':
		err = cmd.parseFreezeFlags(arg[1:])
		if err != nil {
			return opNone, err
		}
		op |= opFreeze
	case 'D':
		op, err = cmd.parseDatabaseFlags(arg[1:])
		if err != nil {
			return opNone, err
		}
		op |= opDatabase
	case 'Q':
		op = opQuery
		if len(arg) > 1 {
			return opNone, errInvalidOptions
		}
	case 'S':
		op, err = cmd.parseSyncFlags(arg[1:])
		if err != nil {
			return opNone, err
		}
		op |= opSync
	case 'R':
		op, err = cmd.parseRemoveFlags(arg[1:])
		if err != nil {
			return opNone, err
		}
		op |= opRemove
	case 'V':
		if len(arg) > 1 {
			return opNone, errInvalidOptions
		}
		cmd.vendor = true
	default:
		return opNone, errInvalidOptions
	}

	cmd.op |= op

	return op, nil
}

func (cmd *command) parseLongFlags(arg string) (op operation, err error) {
	if len(arg) == 0 {
		return opNone, errInvalidOptions
	}
	switch arg {
	case "help":
		op = opHelp
	case "database":
		op = opDatabase
	case "exclude":
		op = opExclude
	case "freeze":
		op = opFreeze
	case "into":
		op = opSyncInto
	case "noconfirm":
		cmd.noConfirm = true
	case "nodeps":
		cmd.noDeps = true
	case "query":
		op = opQuery
	case "recursive":
		op = opRecursive
	case "remove":
		op = opRemove
	case "sync":
		op = opSync
	case "update":
		op = opUpdate
	case "vendor":
		cmd.vendor = true
	case "version":
		op = opVersion
	default:
		return opNone, errInvalidOptions
	}

	cmd.op |= op

	return
}

//
// parseFlags for multiple operations, invalid options, or empty targets.
//
func (cmd *command) parseFlags(args []string) (err error) {
	if len(args) == 0 {
		return errNoOperation
	}

	var (
		fl int
		op operation
	)
	for _, arg := range args {
		fl = 0
		for y, r := range arg {
			if fl == 1 {
				if r == '-' {
					fl++
					continue
				}
				op, err = cmd.parseShortFlags(arg[y:])
				if err != nil {
					return
				}
				break
			}
			if fl == 2 {
				op, err = cmd.parseLongFlags(arg[y:])
				if err != nil {
					return
				}
				break
			}
			if y == 0 && r == '-' {
				fl++
				continue
			}

			if op == opSyncInto {
				cmd.syncInto = arg
			} else {
				cmd.pkgs = append(cmd.pkgs, arg)
			}
			break
		}

		// "-h", "--help", "--version" flag is a stopper.
		if op == opHelp || op == opVersion {
			cmd.op = op
			return
		}
	}

	switch cmd.op {
	case opNone, opExclude, opRecursive, opSyncInto, opUpdate:
		return errInvalidOptions
	}

	if cmd.op&opSyncInto == opSyncInto {
		if cmd.op&opSync != opSync {
			return errInvalidOptions
		}
	}

	// Only one operation is allowed.
	op = cmd.op & (opDatabase | opFreeze | opQuery | opRemove | opSync)
	if op != opDatabase && op != opFreeze && op != opQuery &&
		op != opRemove && op != opSync {
		return errMultiOperations
	}

	// "-R" or "-S" must have target
	if op == opRemove {
		if len(cmd.pkgs) == 0 {
			return errNoTarget
		}
	}

	return nil
}

func (cmd *command) loadDatabase() (err error) {
	err = cmd.env.Load(beku.DefDBName)
	if err == nil {
		return
	}

	if !cmd.vendor {
		err = cmd.env.Load("")
	}

	return
}

func (cmd *command) sync() (err error) {
	if len(cmd.pkgs) > 1 && len(cmd.syncInto) > 0 {
		return errInvalidOptions
	}

	var ok bool

	if cmd.firstTime {
		ok, err = cmd.env.Rescan(true)
		if !ok || err != nil {
			return
		}
	}

	switch len(cmd.pkgs) {
	case 0:
		if cmd.op&opUpdate == 0 {
			if !cmd.firstTime {
				_, err = cmd.env.Rescan(false)
			}
		} else {
			err = cmd.env.SyncAll()
		}
	case 1:
		err = cmd.env.Sync(cmd.pkgs[0], cmd.syncInto)
	default:
		err = cmd.env.SyncMany(cmd.pkgs)
	}

	return
}

func newCommand() (cmd *command, err error) {
	cmd = &command{}

	err = cmd.parseFlags(os.Args[1:])
	if err != nil {
		return
	}

	switch cmd.op {
	case opHelp:
		cmd.usage()
		os.Exit(1)
	case opVersion:
		cmd.version()
		os.Exit(0)
	}

	cmd.env, err = beku.NewEnvironment(cmd.vendor, cmd.noDeps)
	if err != nil {
		return
	}

	err = cmd.loadDatabase()
	if err != nil {
		if os.IsNotExist(err) {
			if cmd.op&opSync > 0 {
				cmd.firstTime = true
				err = nil
			} else {
				err = errNoDB
			}
		}
	}

	if debug.Value >= 2 {
		fmt.Printf("Environment: %s", cmd.env.String())
	}

	return
}
