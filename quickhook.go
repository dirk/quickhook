package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli"

	"github.com/dirk/quickhook/context"
	"github.com/dirk/quickhook/hooks"
)

const VERSION = "1.4.0"

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
						files := []string{}
						if c.Bool("files") {
							files = c.Args()
						}

						err := hooks.PreCommit(context, &hooks.PreCommitOpts{
							All:     c.Bool("all"),
							Files:   files,
							NoColor: c.Bool("no-color"),
						})
						if err != nil {
							panic(err)
						}
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
							NoColor:         c.Bool("no-color"),
							MessageTempFile: messageTempFile,
						})
						if err != nil {
							panic(err)
						}
						return nil
					},
				},
			},
		},
		{
			Name:      "install",
			Usage:     "Install Quickhook shims into .git/hooks",
			ArgsUsage: " ", // Don't show "[arguments...]"
			Flags: []cli.Flag{
				yesFlag(),
			},
			Action: func(c *cli.Context) error {
				prompt := c.Bool("yes") != true

				err := Install(context, prompt)
				if err != nil {
					panic(err)
				}
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
		Name:   "no-color",
		EnvVar: "NO_COLOR,QUICKHOOK_NO_COLOR",
		Usage:  "Don't colorize output",
	}
}

func allFlag() cli.Flag {
	return cli.BoolFlag{
		Name:  "all, a",
		Usage: "Run on all Git-tracked files",
	}
}

func filesFlag() cli.Flag {
	return cli.BoolFlag{
		Name:  "files, F",
		Usage: "Run on the given comma-separated list of files",
	}
}

func yesFlag() cli.Flag {
	return cli.BoolFlag{
		Name:  "yes, y",
		Usage: "Assume yes for all prompts",
	}
}
