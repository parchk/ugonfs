package main

import (
	"NfsAgent/server"

	_ "NfsAgent/gm"
	"NfsAgent/mlog"

	"github.com/prashanthpai/sunrpc"
)

func main() {

	mlog.Debug(sunrpc.DumpProcedureRegistry())

	server.Run()
}
