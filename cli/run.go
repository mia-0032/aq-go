package cli

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mia-0032/aq-go/cmd"
	"github.com/Bowery/prompt"
	"github.com/urfave/cli"
)

var BucketFlag = cli.StringFlag{
	Name: "bucket, b",
	Usage: "S3 bucket where the query result is stored.",
	EnvVar: "AQ_DEFAULT_BUCKET",
}

var ObjectPrefixFlag = cli.StringFlag{
	Name: "object_prefix, o",
	Value: "Unsaved/" + time.Now().Format("2006/01/02"),
	Usage: "S3 object prefix where the query result is stored.",
}

var Commands = []cli.Command{
	{
		Name:   "query",
		Usage:  "Run query",
		Action: cmd.Query,
		ArgsUsage:   "QUERY",
		Flags: []cli.Flag{
			BucketFlag,
			ObjectPrefixFlag,
			cli.IntFlag{
				Name: "timeout, t",
				Value: 0,
				Usage: "Wait for execution of the query for this number of seconds. If this is set to 0, timeout is disabled.",
			},
		},
		Before: func(c *cli.Context) error {
			if c.NArg() == 0 {
				return cli.NewExitError("QUERY must be specified.", 1)
			}
			if c.String("bucket") == "" {
				return cli.NewExitError("bucket must be specified.", 1)
			}
			return nil
		},
	},
	{
		Name:   "ls",
		Usage:  "Show databases or tables in specified database",
		Action: cmd.Ls,
		ArgsUsage:   "[DATABASE]",
		Flags: []cli.Flag{
			BucketFlag,
			ObjectPrefixFlag,
		},
		Before: func(c *cli.Context) error {
			if c.String("bucket") == "" {
				return cli.NewExitError("bucket must be specified.", 1)
			}
			return nil
		},
	},
	{
		Name:   "head",
		Usage:  "Show records in specified table",
		Action: cmd.Head,
		ArgsUsage:   "DATABASE.TABLE",
		Flags: []cli.Flag{
			BucketFlag,
			ObjectPrefixFlag,
			cli.IntFlag{
				Name: "max_rows, n",
				Value: 100,
				Usage: "This number of rows are printed.",
			},
		},
		Before: func(c *cli.Context) error {
			if c.String("bucket") == "" {
				return cli.NewExitError("bucket must be specified.", 1)
			}
			if c.NArg() == 0 {
				return cli.NewExitError("DATABASE and TABLE must be specified.", 1)
			}
			if len(strings.Split(c.Args().First(), ".")) != 2 {
				return cli.NewExitError("[DATABASE].[TABLE] must contain `.`.", 1)
			}
			return nil
		},
	},
	{
		Name:   "mk",
		Usage:  "Create database",
		Action: cmd.Mk,
		ArgsUsage:   "DATABASE",
		Flags: []cli.Flag{
			BucketFlag,
			ObjectPrefixFlag,
		},
		Before: func(c *cli.Context) error {
			if c.String("bucket") == "" {
				return cli.NewExitError("bucket must be specified.", 1)
			}
			if c.NArg() == 0 {
				return cli.NewExitError("DATABASE must be specified.", 1)
			}
			if len(strings.Split(c.Args().First(), ".")) >= 2 {
				return cli.NewExitError("If you want to create table, use `load` subcommand.", 1)
			}
			return nil
		},
	},
	{
		Name:   "rm",
		Usage:  "Drop database or table",
		Action: cmd.Rm,
		ArgsUsage:   "NAME",
		Flags: []cli.Flag{
			BucketFlag,
			ObjectPrefixFlag,
			cli.BoolFlag{
				Name: "force, f",
				Usage: "Skip confirmation if this is set.",
			},
		},
		Before: func(c *cli.Context) error {
			if c.String("bucket") == "" {
				return cli.NewExitError("bucket must be specified.", 1)
			}
			if c.NArg() == 0 {
				return cli.NewExitError("NAME must be specified.", 1)
			}

			var answer bool
			if c.Bool("force") {
				answer = true
			} else {
				answer, _ = prompt.Ask("Would you remove " + c.Args().First())
			}
			if !answer {
				return cli.NewExitError("Canceled.", 1)
			}

			return nil
		},
	},
	{
		Name:   "load",
		Usage:  "Create table and load data",
		Action: cmd.Load,
		ArgsUsage:   "DATABASE.TABLE SOURCE SCHEMA",
		Flags: []cli.Flag{
			BucketFlag,
			ObjectPrefixFlag,
			cli.StringFlag{
				Name: "source_format, s",
				Value: "NEWLINE_DELIMITED_JSON",
				Usage: "Specify source file data format. Now aq support only NEWLINE_DELIMITED_JSON.",
			},
			cli.StringFlag{
				Name: "partitioning, p",
				Value: "",
				Usage: "Specify partition key and type. ex. key1:type1,key2:type2,...",
			},
		},
		Before: func(c *cli.Context) error {
			if c.String("bucket") == "" {
				return cli.NewExitError("bucket must be specified.", 1)
			}
			if !strings.HasPrefix(c.Args().Get(1), "s3://") {
				return cli.NewExitError("`SOURCE` must start with 's3://'", 1)
			}
			if c.String("source_format") != "NEWLINE_DELIMITED_JSON" {
				return cli.NewExitError("Now aq support only NEWLINE_DELIMITED_JSON.", 1)
			}
			return nil
		},
	},
}

func Run() int {
	app := cli.NewApp()
	app.Name = "aq"
	app.Usage = "Command Line Tool for AWS Athena (bq command like)"
	app.Version = "0.2.0"
	app.EnableBashCompletion = true
	app.Commands = Commands
	app.Action = func (_ *cli.Context) {
		var subcmds []string
		for _, subcmd := range Commands {
			subcmds = append(subcmds, subcmd.Name)
		}
		fmt.Fprintf(os.Stderr, "%s: %s\n", "Subcommands", strings.Join(subcmds, ", "))
	}

	return msg(app.Run(os.Args))
}

func msg(err error) int {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", os.Args[0], err)
		return 1
	}
	return 0
}
