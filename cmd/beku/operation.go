// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type operation uint

const opNone operation = 0

const (
	opHelp operation = 1 << iota
	opDatabase
	opExclude
	opFreeze
	opQuery
	opRecursive
	opRemove
	opSync
	opSyncInto
	opUpdate
	opVersion
)
