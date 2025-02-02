package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/explorer"
	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/explorer/viewer"
	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/table"
	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/tree"
	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/tree/statusbar"
	"github.com/brunoluiz/crossplane-explorer/internal/xplane"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/urfave/cli/v3"
)

func cmdTrace() *cli.Command {
	return &cli.Command{
		Usage: `Explore tracing from Crossplane. Usage is available through arguments or data stream
1. To load it straight from a live resource using the crossplane CLI, do 'crossplane-explorer trace <object name>'
2. To load it from a trace JSON file, do 'crossplane beta trace -o json <> | crossplane-explorer trace --stdin'

Live mode is only available for (1) through the use of --watch / --watch-interval (see flag usage below)`,
		Name:    "trace",
		Aliases: []string{"t"},
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "log", Aliases: []string{"l"}, Usage: "Log destination", Value: "crossplane-explorer.trace.log"},
			&cli.StringFlag{Name: "cmd", Usage: "Which binary should it use to generate the JSON trace", Value: "crossplane beta trace -o json"},
			&cli.StringFlag{Name: "context", Aliases: []string{"ctx"}, Usage: "Kubernetes context to be used"},
			&cli.StringFlag{Name: "namespace", Aliases: []string{"n", "ns"}, Usage: "Kubernetes namespace to be used"},
			&cli.BoolFlag{Name: "stdin", Aliases: []string{"in"}, Usage: "Specify in case file is piped into stdin"},
			&cli.BoolFlag{Name: "short", Usage: "Return short result columns for small screens"},
			&cli.BoolFlag{Name: "watch", Aliases: []string{"w"}, Usage: "Refresh trace every 10 seconds"},
			&cli.DurationFlag{Name: "watch-interval", Aliases: []string{"wi"}, Usage: "Refresh interval for the watcher feature", Value: 5 * time.Second},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			f, err := os.Create(c.String("log"))
			if err != nil {
				return err
			}

			logger := slog.New(slog.NewTextHandler(f, &slog.HandlerOptions{}))

			app := tea.NewProgram(
				explorer.New(
					logger,
					tree.New(
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
						statusbar.New(),
					),
					viewer.New(),
					getTracer(c),
					explorer.WithWatch(c.Bool("watch")),
					explorer.WithWatchInterval(c.Duration("watch-interval")),
					explorer.WithShortColumns(c.Bool("short")),
				),
				tea.WithAltScreen(),
				tea.WithContext(ctx),
			)

			_, err = app.Run()
			return err
		},
	}
}

func getTracer(c *cli.Command) explorer.Tracer {
	if c.Bool("stdin") {
		return xplane.NewReaderTraceQuerier(os.Stdin)
	}

	return xplane.NewCLITraceQuerier(
		c.String("cmd"),
		c.String("namespace"),
		c.String("context"),
		c.Args().First(),
	)
}
