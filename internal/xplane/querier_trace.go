package xplane

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// CLITraceQuerier defines a trace querier using the crossplane CLI.
type CLITraceQuerier struct {
	app  string
	args []string
}

func NewCLITraceQuerier(
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
		app:  app,
		args: args,
	}
}

func (q *CLITraceQuerier) GetTrace() (*Resource, error) {
	//nolint // trust the user input
	stdout, err := exec.Command(q.app, q.args...).Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get trace from CLI: %w", err)
	}

	return Parse(bytes.NewReader(stdout))
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
