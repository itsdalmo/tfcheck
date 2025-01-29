package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	tfcheck "github.com/itsdalmo/tfcheck"
)

var version = "dev"

func main() {
	app := &cli.App{
		Name:      "tfcheck",
		Usage:     "Runner for terraform checks (fmt, validate and tflint)",
		UsageText: "tfcheck [--no-tui] [--max-in-parallel=<number>] [--tflint-config=<path>] [DIRECTORY]",
		Version:   version,
		Action:    run,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "no-tui",
				Usage: "Disable the TUI even if we are running in a TTY",
			},
			&cli.IntFlag{
				Name:    "max-in-parallel",
				Usage:   "Limit the number of jobs executed in parallel",
				Aliases: []string{"p"},
			},
			&cli.StringFlag{
				Name:  "tflint-config",
				Usage: "Optional config to use for tflint",
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		os.Exit(1)
	}
}

func run(c *cli.Context) error {
	if c.NArg() > 1 {
		err := fmt.Errorf("unknown argument(s): %v", c.Args().Tail())
		return cli.Exit(err, 1)
	}

	rootDirectory := c.Args().Get(0)
	if rootDirectory == "" {
		rootDirectory = "."
	}

	dirs, err := tfcheck.FindTerraformDirectories(rootDirectory)
	if err != nil {
		return cli.Exit(err, 1)
	}

	cfg := tfcheck.Config{
		Directories:   dirs,
		MaxInParallel: c.Int("max-in-parallel"),
		NoTUI:         c.Bool("no-tui"),
		TFLintConfig:  c.String("tflint-config"),
	}

	if err := tfcheck.Run(cfg); err != nil {
		return cli.Exit(err, 1)
	}

	return nil
}
