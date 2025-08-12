package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/urfave/cli/v3"
)

var logger *slog.Logger = slog.New(slog.DiscardHandler)

func cmdMain(cmds ...*cli.Command) *cli.Command {
	return &cli.Command{
		Name:  "xpdig",
		Usage: "Set of tools to explore your crossplane resources",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "log",
				Aliases: []string{"l"},
				Usage:   "Log destination (default: none, but example: /tmp/logs.txt)",
				Value:   "",
			},
			&cli.StringFlag{
				Name:    "log-level",
				Aliases: []string{"ll"},
				Usage:   "Log level (default: INFO, available: DEBUG, INFO, WARN, ERROR)",
				Value:   slog.LevelInfo.String(),
			},
		},
		Commands: cmds,
		Before: func(ctx context.Context, c *cli.Command) (context.Context, error) {
			if c.String("log") != "" {
				f, err := os.Create(c.String("log"))
				if err != nil {
					return ctx, err
				}

				var level slog.Level
				if err = level.UnmarshalText([]byte(c.String("log-level"))); err != nil {
					return ctx, err
				}

				logger = slog.New(slog.NewJSONHandler(f, &slog.HandlerOptions{
					Level: level,
				}))
			}

			return ctx, nil
		},
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
		cmdVersion(),
	).Run(ctx, os.Args); err != nil {
		log.Println(err)
	}
}
