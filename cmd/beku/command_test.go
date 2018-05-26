package main

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestParseFlags(t *testing.T) {
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
		args:   []string{"-Q", "package", "--into", "directory"},
		expErr: errInvalidOptions.Error(),
	}, {
		args:   []string{"-R", "package", "--into", "directory"},
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
			op: opQuery | opHelp,
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
