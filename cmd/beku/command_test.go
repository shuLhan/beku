// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"go/build"
	"os"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func testParseFlags(t *testing.T) {
	cases := []struct {
		args   []string
		expErr string
		expCmd *command
	}{{
		expErr: errNoOperation.Error(),
	}, {
		args:   []string{"-s", "-"},
		expErr: errInvalidOptions.Error(),
	}, {
		args:   []string{"-su", "-"},
		expErr: errInvalidOptions.Error(),
	}, {
		args:   []string{"-hs"},
		expErr: errInvalidOptions.Error(),
	}, {
		args: []string{"-h"},
		expCmd: &command{
			op: opHelp,
		},
	}, {
		args: []string{"--help"},
		expCmd: &command{
			op: opHelp,
		},
	}, {
		args:   []string{"--into"},
		expErr: errInvalidOptions.Error(),
	}, {
		args:   []string{"--noconfirm"},
		expErr: errInvalidOptions.Error(),
	}, {
		args:   []string{"--update"},
		expErr: errInvalidOptions.Error(),
	}, {
		args:   []string{"-s"},
		expErr: errInvalidOptions.Error(),
	}, {
		args:   []string{"--recursive"},
		expErr: errInvalidOptions.Error(),
	}, {
		args:   []string{"--into", "directory"},
		expErr: errInvalidOptions.Error(),
	}, {
		args: []string{"-B"},
		expCmd: &command{
			op: opFreeze,
		},
	}, {
		args: []string{"--freeze"},
		expCmd: &command{
			op: opFreeze,
		},
	}, {
		args:   []string{"-Bs"},
		expErr: errInvalidOptions.Error(),
	}, {
		args: []string{"-D"},
		expCmd: &command{
			op: opDatabase,
		},
	}, {
		args:   []string{"-Ds"},
		expErr: errInvalidOptions.Error(),
	}, {
		args: []string{"-De"},
		expCmd: &command{
			op: opDatabase | opExclude,
		},
	}, {
		args: []string{"--database"},
		expCmd: &command{
			op: opDatabase,
		},
	}, {
		args: []string{"--database", "--exclude", "A"},
		expCmd: &command{
			op: opDatabase | opExclude,
			pkgs: []string{
				"A",
			},
		},
	}, {
		args:   []string{"-Qs", "A"},
		expErr: errInvalidOptions.Error(),
	}, {
		args:   []string{"-Q", "package", "--into", "directory"},
		expErr: errInvalidOptions.Error(),
	}, {
		args:   []string{"-Q", "query", "-R", "remove"},
		expErr: errMultiOperations.Error(),
	}, {
		args:   []string{"-Q", "-R", "remove"},
		expErr: errMultiOperations.Error(),
	}, {
		args:   []string{"-Q", "query", "-S", "sync"},
		expErr: errMultiOperations.Error(),
	}, {
		args:   []string{"-S", "sync", "-R", "remove"},
		expErr: errMultiOperations.Error(),
	}, {
		args: []string{"-Q"},
		expCmd: &command{
			op: opQuery,
		},
	}, {
		args: []string{"--query"},
		expCmd: &command{
			op: opQuery,
		},
	}, {
		args: []string{"-Q", "-h"},
		expCmd: &command{
			op: opHelp,
		},
	}, {
		args: []string{"A", "-Q", "B"},
		expCmd: &command{
			op:   opQuery,
			pkgs: []string{"A", "B"},
		},
	}, {
		args: []string{"-S"},
		expCmd: &command{
			op: opSync,
		},
	}, {
		args: []string{"-Su"},
		expCmd: &command{
			op: opSync | opUpdate,
		},
	}, {
		args:   []string{"-Sh"},
		expErr: errInvalidOptions.Error(),
	}, {
		args: []string{"--sync"},
		expCmd: &command{
			op: opSync,
		},
	}, {
		args: []string{"-S", "package", "another"},
		expCmd: &command{
			op:   opSync,
			pkgs: []string{"package", "another"},
		},
	}, {
		args: []string{"-S", "package", "--into", "directory"},
		expCmd: &command{
			op:       opSync | opSyncInto,
			pkgs:     []string{"package"},
			syncInto: "directory",
		},
	}, {
		args: []string{"--sync", "A"},
		expCmd: &command{
			op:   opSync,
			pkgs: []string{"A"},
		},
	}, {
		args:   []string{"-R"},
		expErr: errNoTarget.Error(),
	}, {
		args:   []string{"--remove"},
		expErr: errNoTarget.Error(),
	}, {
		args:   []string{"-Rs"},
		expErr: errNoTarget.Error(),
	}, {
		args:   []string{"-R", "package", "--into", "directory"},
		expErr: errInvalidOptions.Error(),
	}, {
		args: []string{"-R", "A"},
		expCmd: &command{
			op:   opRemove,
			pkgs: []string{"A"},
		},
	}, {
		args: []string{"--remove", "A"},
		expCmd: &command{
			op:   opRemove,
			pkgs: []string{"A"},
		},
	}, {
		args: []string{"--remove", "-s", "A"},
		expCmd: &command{
			op:   opRemove | opRecursive,
			pkgs: []string{"A"},
		},
	}, {
		args: []string{"--remove", "--recursive", "A"},
		expCmd: &command{
			op:   opRemove | opRecursive,
			pkgs: []string{"A"},
		},
	}, {
		args: []string{"--remove", "A", "--recursive"},
		expCmd: &command{
			op:   opRemove | opRecursive,
			pkgs: []string{"A"},
		},
	}, {
		args:   []string{"--remove", "A", "---recursive"},
		expErr: errInvalidOptions.Error(),
	}, {
		args: []string{"-R", "A", "-s"},
		expCmd: &command{
			op:   opRemove | opRecursive,
			pkgs: []string{"A"},
		},
	}, {
		args: []string{"-R", "A", "--recursive"},
		expCmd: &command{
			op:   opRemove | opRecursive,
			pkgs: []string{"A"},
		},
	}, {
		args: []string{"-Rs", "A"},
		expCmd: &command{
			op:   opRemove | opRecursive,
			pkgs: []string{"A"},
		},
	}, {
		args: []string{"-Rs", "A", "--noconfirm"},
		expCmd: &command{
			op:        opRemove | opRecursive,
			noConfirm: true,
			pkgs:      []string{"A"},
		},
	}, {
		args:   []string{"-Rx", "A"},
		expErr: errInvalidOptions.Error(),
	}, {
		args:   []string{"-T", "A"},
		expErr: errInvalidOptions.Error(),
	}}

	for _, c := range cases {
		t.Log(c.args)

		cmd := new(command)

		err := cmd.parseFlags(c.args)
		if err != nil {
			test.Assert(t, "err", c.expErr, err.Error(), true)
			continue
		}

		test.Assert(t, "cmd", c.expCmd, cmd, true)
	}
}

func testNewCommand(t *testing.T) {
	cases := []struct {
		desc   string
		gopath string
		args   []string
		expCmd *command
		expErr string
	}{{
		desc: "With sync",
		args: []string{
			"beku", "-S", "A",
		},
		expCmd: &command{
			op:   opSync,
			pkgs: []string{"A"},
		},
	}, {
		desc:   "With sync operation and no database found",
		gopath: "/tmp",
		args: []string{
			"beku", "-S", "A",
		},
		expCmd: &command{
			op:        opSync,
			pkgs:      []string{"A"},
			firstTime: true,
		},
	}, {
		desc:   "With remove operation and no database found",
		gopath: "/tmp",
		args: []string{
			"beku", "-R", "A",
		},
		expCmd: &command{
			op:        opRemove,
			pkgs:      []string{"A"},
			firstTime: false,
		},
		expErr: errNoDB.Error(),
	}}

	for _, c := range cases {
		orgGOPATH := build.Default.GOPATH
		orgArgs := os.Args

		if len(c.gopath) > 0 {
			build.Default.GOPATH = c.gopath
		}
		os.Args = c.args

		cmd, err := newCommand()
		if err != nil {
			test.Assert(t, "err", c.expErr, err.Error(), true)
			build.Default.GOPATH = orgGOPATH
			os.Args = orgArgs
			continue
		}

		test.Assert(t, "command.op", c.expCmd.op, cmd.op, true)
		test.Assert(t, "command.pkgs", c.expCmd.pkgs, cmd.pkgs, true)
		test.Assert(t, "command.syncInto", c.expCmd.syncInto, cmd.syncInto, true)
		test.Assert(t, "command.firstTime", c.expCmd.firstTime, cmd.firstTime, true)

		build.Default.GOPATH = orgGOPATH
		os.Args = orgArgs
	}
}

func TestCommand(t *testing.T) {
	t.Run("parseFlags", testParseFlags)
	t.Run("newCommand", testNewCommand)
}
