package xplane

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"strings"
)

// CLITraceQuerier defines a trace querier using the crossplane CLI.
type CLITraceQuerier struct {
	logger *slog.Logger
	app    string
	args   []string
}

func NewCLITraceQuerier(
	logger *slog.Logger,
	cmd string,
	namespace string,
	context string,
	kind string,
	object string,
) *CLITraceQuerier {
	s := strings.Split(cmd, " ")
	app := s[0]
	args := s[1:]
	if namespace != "" && namespace != "-" {
		args = append(args, "--namespace", namespace)
	}

	if context != "" && context != "-" {
		args = append(args, "--context", context)
	}
	args = append(args, fmt.Sprintf("%s/%s", kind, object))

	return &CLITraceQuerier{
		logger: logger,
		app:    app,
		args:   args,
	}
}

func (q *CLITraceQuerier) GetTrace() (*Resource, error) {
	q.logger.Info("executing crossplane", "cmd", q.app, "args", q.args)

	//nolint // trust the user input
	out, err := exec.Command(q.app, q.args...).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failure while executing '%s %s' (%w):\n%s", q.app, strings.Join(q.args, " "), err, out)
	}

	return Parse(bytes.NewReader(out))
}

// ReaderTraceQuerier defines a trace querier using piped files through stdin.
type ReaderTraceQuerier struct {
	r io.Reader
}

func NewReaderTraceQuerier(r io.Reader) *ReaderTraceQuerier {
	return &ReaderTraceQuerier{r: r}
}

func (q *ReaderTraceQuerier) GetTrace() (*Resource, error) {
	return Parse(q.r)
}
