package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/brunoluiz/xpdig/internal/bubbles/action/kubectl"
	"github.com/brunoluiz/xpdig/internal/bubbles/action/shell"
	"github.com/brunoluiz/xpdig/internal/bubbles/app"
	"github.com/brunoluiz/xpdig/internal/bubbles/component/navigator"
	"github.com/brunoluiz/xpdig/internal/bubbles/component/statusbar"
	"github.com/brunoluiz/xpdig/internal/bubbles/component/table"
	"github.com/brunoluiz/xpdig/internal/bubbles/layout/xpnavigator"
	"github.com/brunoluiz/xpdig/internal/xplane"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/davecgh/go-spew/spew"
	"github.com/urfave/cli/v3"
)

func cmdTrace() *cli.Command {
	return &cli.Command{
		Usage: `Explore tracing from Crossplane. Usage is available through arguments or data stream
1. To load it straight from a live resource using the crossplane CLI, do 'xpdig trace <object name>'
2. To load it from a trace JSON file, do 'crossplane beta trace -o json <> | xpdig trace --stdin'

Live mode is only available for (1) through the use of --watch / --watch-interval (see flag usage below)`,
		Name:    "trace",
		Aliases: []string{"t"},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "log",
				Aliases: []string{"l"},
				Usage:   "Log destination (eg: /tmp/logs.txt",
				Value:   "",
			},
			&cli.StringFlag{
				Name:    "dump",
				Aliases: []string{"du"},
				Usage:   "Message dump destination (eg: /tmp/dump.txt",
				Value:   "",
			},
			&cli.StringFlag{
				Name:  "cmd",
				Usage: "Which binary should it use to generate the JSON trace",
				Value: "crossplane beta trace -o json",
			},
			&cli.StringFlag{Name: "context", Aliases: []string{"ctx"}, Usage: "Kubernetes context to be used"},
			&cli.StringFlag{Name: "namespace", Aliases: []string{"n", "ns"}, Usage: "Kubernetes namespace to be used"},
			&cli.BoolFlag{Name: "stdin", Aliases: []string{"in"}, Usage: "Specify in case file is piped into stdin"},
			&cli.BoolFlag{Name: "short", Usage: "Return short result columns for small screens"},
			&cli.BoolFlag{Name: "watch", Aliases: []string{"w"}, Usage: "Refresh trace every 10 seconds"},
			&cli.DurationFlag{
				Name:    "watch-interval",
				Aliases: []string{"wi"},
				Usage:   "Refresh interval for the watcher feature",
				Value:   5 * time.Second,
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			logger := slog.New(slog.DiscardHandler)
			if c.String("log") != "" {
				f, err := os.Create(c.String("log"))
				if err != nil {
					return err
				}

				logger = slog.New(slog.NewJSONHandler(f, &slog.HandlerOptions{}))
			}

			dumper := func(...any) {}
			if c.String("dump") != "" {
				f, err := os.Create(c.String("dump"))
				if err != nil {
					return err
				}

				dumper = func(a ...any) {
					spew.Fdump(f, a...)
				}
			}

			tracer, err := getTracer(c, logger)
			if err != nil {
				return err
			}

			logger.Info("Starting xpdig", "info", map[string]any{
				"version": version,
				"args":    c.Args().Slice(),
				"flags": map[string]any{
					"cmd":            c.String("cmd"),
					"context":        c.String("context"),
					"namespace":      c.String("namespace"),
					"stdin":          c.Bool("stdin"),
					"short":          c.Bool("short"),
					"watch":          c.Bool("watch"),
					"watch-interval": c.Duration("watch-interval"),
				},
			})

			program := tea.NewProgram(
				app.New(
					logger,
					dumper,
					kubectl.New(c.String("context"), shell.New(logger)),
					xpnavigator.New(
						logger,
						navigator.New(
							logger,
							table.New(
								table.WithFocused(true),
								table.WithStyles(func() table.Styles {
									s := table.DefaultStyles()
									s.Selected = lipgloss.NewStyle().
										Foreground(lipgloss.ANSIColor(ansi.Black)).
										Background(lipgloss.ANSIColor(ansi.White))
									return s
								}()),
							),
							textinput.New(),
						),
						statusbar.New(),
						tracer,
						xpnavigator.WithWatch(c.Bool("watch")),
						xpnavigator.WithWatchInterval(c.Duration("watch-interval")),
						xpnavigator.WithShortColumns(c.Bool("short")),
					),
				),
				tea.WithAltScreen(),
				tea.WithContext(ctx),
			)

			_, err = program.Run()
			if err != nil {
				return err
			}

			return err
		},
	}
}

type ErrInvalidArgument struct{}

func (e *ErrInvalidArgument) Error() string {
	return fmt.Sprintf("trace for is not possible: argument must be on the format '<kind>/<name>' or '<kind> <name>'")
}

func getTracer(c *cli.Command, logger *slog.Logger) (xpnavigator.Tracer, error) {
	if c.Bool("stdin") {
		return xplane.NewReaderTraceQuerier(os.Stdin), nil
	}

	var kind, object string
	switch c.Args().Len() {
	case 1:
		n1 := c.Args().First()
		res := strings.Split(n1, "/")
		if len(res) != 2 {
			return nil, &ErrInvalidArgument{}
		}
		kind, object = res[0], res[1]
	case 2:
		kind, object = c.Args().Get(0), c.Args().Get(1)
	default:
		return nil, &ErrInvalidArgument{}
	}

	return xplane.NewCLITraceQuerier(
		logger,
		c.String("cmd"),
		c.String("namespace"),
		c.String("context"),
		kind, object,
	), nil
}
