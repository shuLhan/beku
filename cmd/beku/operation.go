package main

type operation uint

const (
	opNone operation = 0
	opSync operation = 1 << iota
	opSyncInto
)
