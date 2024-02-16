package main

import "github.com/urfave/cli/v2"

var (
	//main commands
	versionCmd *cli.Command = &cli.Command{
		Name:    "version",
		Aliases: []string{"v"},
		Usage:   "Print cbctl's version",
		Action: func(c *cli.Context) error {
			println(version)
			return nil
		},
	}
	nodeRootCmd *cli.Command = &cli.Command{
		Name:    "node",
		Aliases: []string{"p"},
		Usage:   "Peer management",
		Subcommands: []*cli.Command{
			nodeJoinToCmd,
			nodeListCmd,
			nodeQuitFromCmd,
		},
	}
)

var (
	//node commands
	nodeJoinToCmd *cli.Command = &cli.Command{
		Name:  "join",
		Usage: "Join to node",
		Action: func(c *cli.Context) error {
			return nil
		},
	}
	nodeListCmd *cli.Command = &cli.Command{
		Name:  "list",
		Usage: "List nodes",
		Action: func(c *cli.Context) error {
			return nil
		},
	}
	nodeQuitFromCmd *cli.Command = &cli.Command{
		Name:  "quit",
		Usage: "Quit from node",
		Action: func(c *cli.Context) error {
			return nil
		},
	}
)
