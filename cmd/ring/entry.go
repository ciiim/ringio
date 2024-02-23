package ring

import (
	"fmt"
	"log"
	"os"

	"github.com/ciiim/cloudborad/cmd/ring/ringapi"
	"github.com/ciiim/cloudborad/cmd/ring/router"
	"github.com/ciiim/cloudborad/ringio"
	"github.com/urfave/cli/v2"
)

var (
	App *cli.App

	name string = func() string {
		hostname, err := os.Hostname()
		if err != nil {
			hostname = ringio.DefaultName
		}
		return hostname
	}()

	port int = ringio.DefaultPort

	replica int = ringio.DefualtReplica

	rootPath string = "./ring"

	httpPort int = 0
)

func init() {

	App = cli.NewApp()

	App.Description = "RingIO - A distributed storage system"

	App.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "hostname",
			Usage:       "ring hostname",
			Value:       name,
			Destination: &name,
		},
		&cli.IntFlag{
			Name:        "port",
			Aliases:     []string{"p"},
			Usage:       "ring port",
			Value:       port,
			Destination: &port,
		},
		&cli.StringFlag{
			Name:        "root",
			Usage:       "ring root path",
			Value:       rootPath,
			Destination: &rootPath,
		},
		&cli.IntFlag{
			Name:        "replica",
			Usage:       "ring replica",
			Value:       replica,
			Destination: &replica,
		},

		&cli.IntFlag{
			Name:        "http",
			Usage:       "ring http service",
			Value:       httpPort,
			Destination: &httpPort,
		},
		&cli.StringSliceFlag{
			Name:  "nodes",
			Usage: "join nodes",
		},
	}

	App.Action = func(ctx *cli.Context) error {
		run(ctx)
		return nil
	}

}

func run(ctx *cli.Context) {
	ringapi.InitRingAPI(ctx)

	if httpPort > 0 {
		api := router.Router()

		go func() {
			if err := api.Run(fmt.Sprintf(":%d", httpPort)); err != nil {
				log.Fatalln(err)
			}
		}()
	}
	nodes := ctx.StringSlice("nodes")
	for _, node := range nodes {
		if err := ringapi.Ring.Join(node); err != nil {
			log.Println("Join Failed:", err)
		}
	}
	ringapi.Ring.Serve()
}
