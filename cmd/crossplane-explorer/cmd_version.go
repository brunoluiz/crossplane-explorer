package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/urfave/cli/v3"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func cmdVersion() *cli.Command {
	return &cli.Command{
		Name:  "version",
		Usage: "prints the cli version",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			s := strings.Join([]string{
				fmt.Sprintf("app: crossplane-explorer"),
				fmt.Sprintf("version: %s", version),
				fmt.Sprintf("commit: %s", commit),
				fmt.Sprintf("date: %s", date),
			}, "\n")
			fmt.Println(s)
			return nil
		},
	}
}
