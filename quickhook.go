package main

import (
	"os"

	"github.com/urfave/cli"

	"github.com/dirk/quickhook/context"
	"github.com/dirk/quickhook/hooks"
)

const version = "0.1.2"

func main() {
	context, err := setupContextInWd()
	if err != nil {
		panic(err)
	}

	app := cli.NewApp()
	app.Name = "quickhook"
	app.Version = version
	app.Usage = "Git hook runner"

	app.Commands = []cli.Command{
		{
			Name:  "hook",
			Usage: "Run a hook",
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
