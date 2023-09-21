package main

import (
	"os"

	"github.com/urfave/cli/v2"
)

const (
	version = "v0.0.1"
)

func main() {
	app := cli.NewApp()
	app.Commands = []*cli.Command{
		versionCmd,
		peerRootCmd,
	}
	app.Flags = []cli.Flag{
		connectFlag,
	}
	err := app.Run(os.Args)
	if err != nil {
		println("cbctl exit: " + err.Error())
	}
}
