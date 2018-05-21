package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/shuLhan/beku"
)

const (
	emptyString = ""

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
	help      bool
	query     bool
	recursive bool
	queryPkg  []string
	rmPkg     string
	syncPkg   string
	syncInto  string
}

func (cmd *command) usage() {
	help := `usage: beku <operation> [...]
operations:
	beku {-h|--help}
		` + flagUsageHelp + `
	beku {-Q|--query} [pkg ...]
		` + flagUsageQuery + `
	beku {-R|--remove} [pkg ...] [-s|--recursive]
		` + flagUsageRemove + `
	beku {-S|--sync} <pkg@version> [--into <directory>]
		` + flagUsageSync + `
`

	fmt.Print(help)

	os.Exit(1)
}

func (cmd *command) setFlags() {
	flag.Usage = cmd.usage

	flag.BoolVar(&cmd.help, "h", false, flagUsageHelp)
	flag.BoolVar(&cmd.help, "help", false, flagUsageHelp)

	flag.BoolVar(&cmd.query, "Q", false, flagUsageQuery)
	flag.BoolVar(&cmd.query, "query", false, flagUsageQuery)

	flag.StringVar(&cmd.rmPkg, "R", emptyString, flagUsageRemove)
	flag.StringVar(&cmd.rmPkg, "remove", emptyString, flagUsageRemove)

	flag.BoolVar(&cmd.recursive, "s", false, flagUsageRecursive)
	flag.BoolVar(&cmd.recursive, "recursive", false, flagUsageRecursive)

	flag.StringVar(&cmd.syncPkg, "S", emptyString, flagUsageSync)
	flag.StringVar(&cmd.syncPkg, "sync", emptyString, flagUsageSync)
	flag.StringVar(&cmd.syncInto, "into", emptyString, flagUsageSyncInto)

	flag.Parse()
}

//
// checkFlags
//
// (0) "-h" or "--help" is always the primary flag.
//
func (cmd *command) checkFlags() {
	// (0)
	if cmd.help {
		cmd.usage()
	}

	args := flag.Args()

	if cmd.recursive {
		cmd.op |= opRecursive
	}

	if cmd.query {
		cmd.op |= opQuery
		cmd.queryPkg = args
		return
	}

	if len(cmd.syncPkg) > 0 {
		cmd.op = opSync

		if len(cmd.syncInto) > 0 {
			cmd.op |= opSyncInto
		}
	}

	if len(cmd.rmPkg) > 0 {
		cmd.op |= opRemove
	}

	// Invalid command parameters
	if cmd.op == opNone || cmd.op == opRecursive || cmd.op == opSyncInto {
		cmd.usage()
	}
}

func (cmd *command) loadDatabase() (err error) {
	err = cmd.env.Load(beku.DefDBName)
	if err == nil {
		return
	}

	err = cmd.env.Load("")

	return
}

func (cmd *command) firstTime() {
	err := cmd.env.Scan()
	if err != nil {
		log.Fatal("Scan:", err)
	}

	err = cmd.env.Save("")
	if err != nil {
		log.Fatal("Save:", err)
	}

	log.Println("Initialization complete.")
}

func newCommand() (err error) {
	cmd = &command{}

	cmd.setFlags()
	cmd.checkFlags()

	cmd.env, err = beku.NewEnvironment()
	if err != nil {
		return
	}

	err = cmd.loadDatabase()
	if err != nil {
		log.Println("No database found.")
		log.Println("Initializing database for the first time...")
		cmd.firstTime()
	}

	if beku.Debug >= beku.DebugL2 {
		log.Printf("Environment: %s", cmd.env)
	}

	return
}
