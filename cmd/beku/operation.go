package main

type operation uint

const opNone operation = 0

const (
	opHelp operation = 1 << iota
	opQuery
	opRecursive
	opRemove
	opSync
	opSyncInto
)
