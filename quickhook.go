package main

import (
	"os"

	"github.com/urfave/cli"

	"github.com/dirk/quickhook/context"
	"github.com/dirk/quickhook/hooks"
)

func main() {
	context, err := setupContextInWd()
	if err != nil {
		panic(err)
	}

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
						err := hooks.PreCommit(context)
						if err != nil { panic(err) }
						return nil
					},
				},
			},
		},
	}

	app.Run(os.Args)
}

// Set up `Context` in current working directory
func setupContextInWd() (*context.Context, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	return context.NewContext(wd)
}
