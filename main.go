package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	flag.StringVar(&TOKEN, "t", "", "telegram bot token")
	flag.StringVar(&DBCONN, "d", "", "mysql or its compatible database connection URL")
	flag.BoolVar(&VerboseMode, "e", false, "display all logs")

	version := flag.Bool("v", false, "display current version and exit")
	setadmin := flag.Int64("setadmin", 0, "set admin and delete all the other existing admins")

	flag.Parse()

	if *version {
		fmt.Println(VERSION)
		os.Exit(0)
	}

	if err := InitDatabase(); err != nil {
		DErrorE(err, "Database Error | Cannot initialize database.")
		os.Exit(1)
	}

	InitTables()
	ReadConfigs()

	if *setadmin > 0 {
		UpdateAdmin(*setadmin, UMSet)
		os.Exit(0)
	}

	InitTelegram()

	<-MakeSysChan()
	DInfo("shutting down.")
}
