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
)
