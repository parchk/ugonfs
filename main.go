package main

import (
	"ugonfs/server"

	_ "ugonfs/gm"
	"ugonfs/mlog"

	"github.com/parchk/sunrpc"
)

func main() {

	mlog.Debug(sunrpc.DumpProcedureRegistry())

	server.Run()
}
