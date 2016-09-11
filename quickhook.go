package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli"

	"github.com/dirk/quickhook/context"
	"github.com/dirk/quickhook/hooks"
)

const VERSION = "1.1.0"

func main() {
	context, err := setupContextInWd()
	if err != nil {
		panic(err)
	}

	app := cli.NewApp()
	app.Name = "quickhook"
	app.Version = VERSION
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
					Flags: []cli.Flag{
						allFlag(),
						filesFlag(),
						noColorFlag(),
					},
					Action: func(c *cli.Context) error {
						err := hooks.PreCommit(context, &hooks.PreCommitOpts{
							All: c.Bool("all"),
							Files: c.String("files"),
							NoColor: c.Bool("no-color"),
						})
						if err != nil { panic(err) }
						return nil
					},
				},
				cli.Command{
					Name: "commit-msg",
					Flags: []cli.Flag{
						noColorFlag(),
					},
					Action: func(c *cli.Context) error {
						messageTempFile := c.Args().Get(0)
						if messageTempFile == "" {
							fmt.Println("Missing message temp file argument")
							os.Exit(1)
						}

						err := hooks.CommitMsg(context, &hooks.CommitMsgOpts{
							NoColor: c.Bool("no-color"),
							MessageTempFile: messageTempFile,
						})
						if err != nil { panic(err) }
						return nil
					},
				},
			},
		},
		{
			Name: "install",
			Usage: "install Quickhook shims into .git/hooks",
			Action: func(c *cli.Context) error {
				err := Install(context)
				if err != nil { panic(err) }
				return nil
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

func noColorFlag() cli.Flag {
	return cli.BoolFlag{
		Name: "no-color",
		EnvVar: "NO_COLOR,QUICKHOOK_NO_COLOR",
		Usage: "don't colorize output",
	}
}

func allFlag() cli.Flag {
	return cli.BoolFlag{
		Name: "all, a",
		Usage: "run on all Git-tracked files",
	}
}

func filesFlag() cli.Flag {
	return cli.StringFlag{
		Name: "files, F",
		Usage: "run on the given comma-separated list of files",
	}
}
