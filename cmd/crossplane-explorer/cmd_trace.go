package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/app"
	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/components/navigator"
	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/components/statusbar"
	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/components/table"
	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/layout/xpnavigator"
	"github.com/brunoluiz/crossplane-explorer/internal/bubbles/layout/xpsummary"
	"github.com/brunoluiz/crossplane-explorer/internal/xplane"
	"github.com/charmbracelet/bubbles/textinput"
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
			&cli.StringFlag{Name: "log", Aliases: []string{"l"}, Usage: "Log destination (eg: /tmp/logs.txt", Value: ""},
			&cli.StringFlag{Name: "cmd", Usage: "Which binary should it use to generate the JSON trace", Value: "crossplane beta trace -o json"},
			&cli.StringFlag{Name: "context", Aliases: []string{"ctx"}, Usage: "Kubernetes context to be used"},
			&cli.StringFlag{Name: "namespace", Aliases: []string{"n", "ns"}, Usage: "Kubernetes namespace to be used"},
			&cli.BoolFlag{Name: "stdin", Aliases: []string{"in"}, Usage: "Specify in case file is piped into stdin"},
			&cli.BoolFlag{Name: "short", Usage: "Return short result columns for small screens"},
			&cli.BoolFlag{Name: "watch", Aliases: []string{"w"}, Usage: "Refresh trace every 10 seconds"},
			&cli.DurationFlag{Name: "watch-interval", Aliases: []string{"wi"}, Usage: "Refresh interval for the watcher feature", Value: 5 * time.Second},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			logger := slog.New(slog.DiscardHandler)
			if c.String("log") != "" {
				f, err := os.Create(c.String("log"))
				if err != nil {
					return err
				}

				logger = slog.New(slog.NewTextHandler(f, &slog.HandlerOptions{}))
			}

			nav := navigator.New(
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
			)

			program := tea.NewProgram(
				app.New(
					logger,
					xpnavigator.New(
						logger,
						nav,
						statusbar.New(),
						getTracer(c),
						xpnavigator.WithWatch(c.Bool("watch")),
						xpnavigator.WithWatchInterval(c.Duration("watch-interval")),
						xpnavigator.WithShortColumns(c.Bool("short")),
					),
					xpsummary.New(),
				),
				tea.WithAltScreen(),
				tea.WithContext(ctx),
			)

			_, err := program.Run()
			return err
		},
	}
}

func getTracer(c *cli.Command) xpnavigator.Tracer {
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
