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
		Action: func(_ context.Context, _ *cli.Command) error {
			s := strings.Join([]string{
				"app: xpdig",
				fmt.Sprintf("version: %s", version),
				fmt.Sprintf("commit: %s", commit),
				fmt.Sprintf("date: %s", date),
			}, "\n")
			fmt.Println(s)
			return nil
		},
	}
}
