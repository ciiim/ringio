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
	peerRootCmd *cli.Command = &cli.Command{
		Name:    "peer",
		Aliases: []string{"p"},
		Usage:   "Peer management",
		Subcommands: []*cli.Command{
			peerJoinToCmd,
			peerListCmd,
			peerQuitFromCmd,
		},
	}
)

var (
	//peer commands
	peerJoinToCmd *cli.Command = &cli.Command{
		Name:  "join",
		Usage: "Join to peer",
		Action: func(c *cli.Context) error {
			return nil
		},
	}
	peerListCmd *cli.Command = &cli.Command{
		Name:  "list",
		Usage: "List peers",
		Action: func(c *cli.Context) error {
			return nil
		},
	}
	peerQuitFromCmd *cli.Command = &cli.Command{
		Name:  "quit",
		Usage: "Quit from peer",
		Action: func(c *cli.Context) error {
			return nil
		},
	}
)
