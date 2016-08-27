package main

import (
	"os"

	"github.com/dirk/quickhook/hooks"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()

	app.Commands = []cli.Command{
		{
			Name:  "hook",
			Usage: "run a hook",
			Action: func(c *cli.Context) error {
				return cli.ShowSubcommandHelp(c)
			},
			Subcommands: []cli.Command{
				cli.Command{
					Name: "pre-commit",
					Action: func(c *cli.Context) error {
						return hooks.PreCommit()
					},
				},
			},
		},
	}

	app.Run(os.Args)
}
