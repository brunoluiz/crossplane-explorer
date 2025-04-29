package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/urfave/cli/v3"
)

func cmdMain(cmds ...*cli.Command) *cli.Command {
	return &cli.Command{
		Name:     "crossplane-explorer",
		Usage:    "Set of tools to explore your crossplane resources",
		Commands: cmds,
	}
}

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		os.Interrupt,
	)
	defer stop()

	if err := cmdMain(
		cmdTrace(),
	).Run(ctx, os.Args); err != nil {
		log.Println(err)
	}
}
