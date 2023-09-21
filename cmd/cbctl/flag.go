package main

import (
	"github.com/urfave/cli/v2"
)

var (
	//main flags
	connectFlag = &cli.StringSliceFlag{
		Name:    "host",
		Aliases: []string{"c"},
		Value:   cli.NewStringSlice("localhost:9631", "123456"),
		Usage:   "Connect to peer",
		Action: func(c *cli.Context, args []string) error {
			if len(args) > 2 {
				return cli.Exit("Too many arguments for connect flag", 1)
			}
			return nil
		},
	}
)
