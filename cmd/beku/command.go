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
	errNoTarget        = errors.New("error: no targets specified")
)

const (
	emptyValue = ""

	flagUsageHelp      = "Show the short usage."
	flagUsageQuery     = "Query the package database."
	flagUsageRecursive = "Remove target include their dependencies."
	flagUsageRemove    = "Remove package from GOPATH."
	flagUsageSync      = "Synchronize `package`."
	flagUsageSyncInto  = "Package download `directory`."
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
		` + flagUsageHelp + `
	beku {-Q|--query} [pkg ...]
		` + flagUsageQuery + `
	beku {-R|--remove} <pkg> [-s|--recursive]
		` + flagUsageRemove + `
	beku {-S|--sync} <pkg[@version]> [--into <directory>]
		` + flagUsageSync + `
`

	fmt.Fprint(os.Stderr, help)
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
	case 'Q':
		op = opQuery
		if len(arg) > 1 {
			return opNone, errInvalidOptions
		}
	case 'S':
		op = opSync
		if len(arg) > 1 {
			return opNone, errInvalidOptions
		}
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
	if cmd.op == opRecursive || cmd.op == opSyncInto {
		return errInvalidOptions
	}
	if cmd.op&opSyncInto == opSyncInto {
		if cmd.op&opSync != opSync {
			return errInvalidOptions
		}
	}

	// (1)
	op = cmd.op & (opQuery | opRemove | opSync)
	if op == opQuery|opRemove || op == opQuery|opSync || op == opRemove|opSync {
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
	if cmd.firstTime || len(cmd.pkgs) == 0 {
		err = cmd.env.Rescan()
		if err != nil {
			return
		}
	}
	if len(cmd.pkgs) > 0 {
		err = cmd.env.Sync(cmd.pkgs[0], cmd.syncInto)
	}
	return
}

func newCommand() (err error) {
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
		fmt.Printf("Environment: %s", cmd.env)
	}

	return
}
