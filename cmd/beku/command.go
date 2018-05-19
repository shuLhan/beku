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

	flagHelpUsage     = "Show the short usage."
	flagQueryUsage    = "Query the package database."
	flagRemoveUsage   = "Remove package from GOPATH."
	flagSyncUsage     = "Synchronize `package`."
	flagSyncIntoUsage = "Package download `directory`."
)

type command struct {
	op       operation
	env      *beku.Env
	help     bool
	query    bool
	queryPkg []string
	rmPkg    string
	syncPkg  string
	syncInto string
}

func (cmd *command) usage() {
	help := `usage: beku <operation> [...]
operations:
	beku {-h|--help}
		` + flagHelpUsage + `
	beku {-Q|--query} [pkg ...]
		` + flagQueryUsage + `
	beku {-R|--remove} [pkg ...]
		` + flagRemoveUsage + `
	beku {-S|--sync} <pkg@version> [--into <directory>]
		` + flagSyncUsage + `
`

	fmt.Print(help)

	os.Exit(1)
}

func (cmd *command) setFlags() {
	flag.Usage = cmd.usage

	flag.BoolVar(&cmd.help, "h", false, flagHelpUsage)
	flag.BoolVar(&cmd.help, "help", false, flagHelpUsage)

	flag.BoolVar(&cmd.query, "Q", false, flagQueryUsage)
	flag.BoolVar(&cmd.query, "query", false, flagQueryUsage)

	flag.StringVar(&cmd.rmPkg, "R", emptyString, flagRemoveUsage)
	flag.StringVar(&cmd.rmPkg, "remove", emptyString, flagRemoveUsage)

	flag.StringVar(&cmd.syncPkg, "S", emptyString, flagSyncUsage)
	flag.StringVar(&cmd.syncPkg, "sync", emptyString, flagSyncUsage)
	flag.StringVar(&cmd.syncInto, "into", emptyString, flagSyncIntoUsage)

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

	if cmd.query {
		cmd.op = opQuery
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
	if cmd.op == opNone || cmd.op == opSyncInto {
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
