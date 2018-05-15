package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/shuLhan/beku"
)

type command struct {
	op       operation
	env      *beku.Env
	help     bool
	syncPkg  string
	syncInto string
}

func (cmd *command) usage() {
	help := `usage: beku <operation> [...]
operations:
	beku {-h|--help}
		Show the short usage.
	beku {-S|--sync} <pkg@version> [--into <directory>]
		Synchronize package. Install a package into "$GOPATH/src".
`

	fmt.Print(help)

	os.Exit(1)
}

func (cmd *command) setFlags() {
	flag.Usage = cmd.usage

	flag.BoolVar(&cmd.help, "h", false, "Show the short usage.")
	flag.BoolVar(&cmd.help, "help", false, "Show the short usage.")

	flag.StringVar(&cmd.syncPkg, "sync", emptyString, "Synchronize `package`.")
	flag.StringVar(&cmd.syncPkg, "S", emptyString, "Synchronize `package`.")

	flag.StringVar(&cmd.syncInto, "into", emptyString, "Package download `directory`.")

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

	if len(cmd.syncPkg) > 0 {
		cmd.op = opSync

		if len(cmd.syncInto) > 0 {
			cmd.op |= opSyncInto
		}
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
