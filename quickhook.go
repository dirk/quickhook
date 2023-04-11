package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/fatih/color"

	"github.com/dirk/quickhook/hooks"
	"github.com/dirk/quickhook/repo"
	"github.com/dirk/quickhook/tracing"
)

const VERSION = "1.5.0"

var cli struct {
	Install struct {
		Yes bool   `short:"y" help:"Assume yes for all prompts"`
		Bin string `help:"Path to Quickhook executable to use in the shim (if it's not on $PATH)"`
	} `cmd:"" help:"Install Quickhook shims into .git/hooks"`
	Hook struct {
		PreCommit struct {
			Files []string `help:"For testing, supply list of files as changed files"`
		} `cmd:"" help:"Run pre-commit hooks"`
		CommitMsg struct {
			MessageFile string `arg:"" help:"Temp file containing the commit message"`
		} `cmd:"" help:"Run commit-msg hooks"`
	} `cmd:""`
	NoColor bool             `env:"NO_COLOR" help:"Don't colorize output"`
	Trace   bool             `env:"QUICKHOOK_TRACE" help:"Enable tracing, writes to trace.out"`
	Version kong.VersionFlag `help:"Show version information"`
}

func main() {
	parser, err := kong.New(&cli,
		kong.Vars{
			"version": VERSION,
		})
	if err != nil {
		panic(err)
	}

	args := os.Args[1:]
	// Print the help if there are no args.
	if len(args) == 0 {
		parsed := kong.Context{
			Kong: parser,
		}
		parsed.PrintUsage(false)
		parsed.Exit(1)
	}

	parsed, err := parser.Parse(args)
	parser.FatalIfErrorf(err)

	if cli.Trace {
		finish := tracing.Start()
		defer finish()
	}

	if cli.NoColor {
		color.NoColor = true
	}

	switch parsed.Command() {
	case "install":
		repo, err := repo.NewRepo()
		if err != nil {
			panic(err)
		}

		// TODO: Dry run option.
		prompt := !cli.Install.Yes
		quickhook := strings.TrimSpace(cli.Install.Bin)
		if quickhook == "" {
			quickhook = "quickhook"
		}
		err = install(repo, quickhook, prompt)
		if err != nil {
			panic(err)
		}

	case "hook commit-msg <message-file>":
		repo, err := repo.NewRepo()
		if err != nil {
			panic(err)
		}

		hook := hooks.CommitMsg{
			Repo: repo,
		}
		err = hook.Run(cli.Hook.CommitMsg.MessageFile)
		if err != nil {
			panic(err)
		}

	case "hook pre-commit":
		repo, err := repo.NewRepo()
		if err != nil {
			panic(err)
		}

		hook := hooks.PreCommit{Repo: repo}
		err = hook.Run(cli.Hook.PreCommit.Files)
		if err != nil {
			panic(err)
		}

	default:
		panic(fmt.Sprintf("Unrecognized command: %v", parsed.Command()))
	}
}
