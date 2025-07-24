package kubectl

import (
	tea "github.com/charmbracelet/bubbletea"
)

type shell interface {
	Exec(c string, args ...string) tea.Cmd
	Pager(c string, args ...string) tea.Cmd
}

type Cmd struct {
	shell shell
}

func New(s shell) *Cmd {
	return &Cmd{shell: s}
}

func (k *Cmd) Edit(ns, resource string) tea.Cmd {
	args := []string{"edit", resource}
	if ns != "" {
		args = append(args, "-n", ns)
	}
	return k.shell.Exec("kubectl", args...)
}

func (k *Cmd) Describe(ns, resource string) tea.Cmd {
	args := []string{"describe", resource}
	if ns != "" {
		args = append(args, "-n", ns)
	}
	return k.shell.Pager("kubectl", args...)
}

func (k *Cmd) Get(ns, resource string) tea.Cmd {
	args := []string{"get", resource, "-o", "yaml"}
	if ns != "" {
		args = append(args, "-n", ns)
	}
	return k.shell.Pager("kubectl", args...)
}

func (k *Cmd) Delete(ns, resource string) tea.Cmd {
	args := []string{"delete", resource}
	if ns != "" {
		args = append(args, "-n", ns)
	}
	return k.shell.Exec("kubectl", args...)
}
