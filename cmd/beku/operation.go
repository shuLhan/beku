package main

type operation uint

const opNone operation = 0

const (
	opQuery operation = 1 << iota
	opSync
	opSyncInto
)
