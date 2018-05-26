package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/shuLhan/beku"
)

var (
	errInvalidOptions  = errors.New("error: invalid options")
	errMultiOperations = errors.New("error: only at operation may be used at a time")
	errNoDB            = errors.New("error: no database found")
	errNoOperation     = errors.New("error: no operation specified")
	errNoTarget        = errors.New("error: no package specified")
)

const (
	flagOperationHelp   = "Show the short usage."
	flagOperationFreeze = "Install all packages on database."
	flagOperationQuery  = "Query the package database."
	flagOperationRemove = "Remove package."
	flagOperationSync   = "Synchronize package. If no package is given, it will do rescan."

	flagOptionRecursive = "Remove package including their dependencies."
	flagOptionSyncInto  = "Download package into `directory`."
	flagOptionUpdate    = "Update all packages to latest version."
)

type command struct {
	op        operation
	env       *beku.Env
	pkgs      []string
	syncInto  string
	firstTime bool
}

func (cmd *command) usage() {
	help := `usage: beku <operation> [...]
operations:
	beku {-h|--help}
		` + flagOperationHelp + `

	beku {-B|--freeze}
		` + flagOperationFreeze + `

	beku {-Q|--query} [pkg ...]
		` + flagOperationQuery + `

	beku {-R|--remove} <pkg> [options]
		` + flagOperationRemove + `

	options:
		[-s|--recursive]
			` + flagOptionRecursive + `

	beku {-S|--sync} <pkg[@version]> [options]
		` + flagOperationSync + `

	option:
		[-u|--update]
			` + flagOptionUpdate + `

	options:
		[--into <directory>]
			` + flagOptionSyncInto + `
`

	fmt.Fprint(os.Stderr, help)
}

func (cmd *command) parseSyncFlags(arg string) (operation, error) {
	if len(arg) == 0 {
		return opNone, nil
	}

	switch arg[0] {
	case 'u':
		return opUpdate, nil
	}

	return opNone, errInvalidOptions
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
	case 's':
		op = opRecursive
		if len(arg) > 1 {
			return opNone, errInvalidOptions
		}
	case 'h':
		op = opHelp
		if len(arg) > 1 {
			return opNone, errInvalidOptions
		}
	case 'B':
		op = opFreeze
		if len(arg) > 1 {
			return opNone, errInvalidOptions
		}
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
	case "freeze":
		op = opFreeze
	case "into":
		op = opSyncInto
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
	default:
		return opNone, errInvalidOptions
	}

	cmd.op |= op

	return
}

//
// parseFlags for multiple operations, invalid options, or empty targets.
//
// (0) "-h" or "--help" flag is a stopper.
// (1) Only one operation is allowed.
// (2) "-R" or "-S" must have target
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
		// (0)
		if op == opHelp {
			return
		}
	}
	if cmd.op == opRecursive || cmd.op == opSyncInto || cmd.op == opUpdate {
		return errInvalidOptions
	}
	if cmd.op&opSyncInto == opSyncInto {
		if cmd.op&opSync != opSync {
			return errInvalidOptions
		}
	}

	// (1)
	op = cmd.op & (opFreeze | opQuery | opRemove | opSync)
	if op != opFreeze && op != opQuery && op != opRemove && op != opSync {
		return errMultiOperations
	}

	// (2)
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

	err = cmd.env.Load("")

	return
}

func (cmd *command) sync() (err error) {
	if len(cmd.pkgs) > 1 && len(cmd.syncInto) > 0 {
		return errInvalidOptions
	}

	var ok bool

	if cmd.firstTime {
		ok, err = cmd.env.Rescan()
		if !ok || err != nil {
			return
		}
	}

	switch len(cmd.pkgs) {
	case 0:
		if cmd.op&opUpdate > 0 {
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

	cmd.env, err = beku.NewEnvironment()
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

	if beku.Debug >= beku.DebugL2 {
		fmt.Printf("Environment: %s", cmd.env.String())
	}

	return
}
