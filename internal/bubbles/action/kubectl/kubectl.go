package kubectl

import (
	tea "github.com/charmbracelet/bubbletea"
)

type shell interface {
	Exec(c string, args ...string) tea.Cmd
	Pager(c string, args ...string) tea.Cmd
}

type Cmd struct {
	kubectx string
	shell   shell
}

func New(kubectx string, s shell) *Cmd {
	return &Cmd{kubectx: kubectx, shell: s}
}

func (k *Cmd) Edit(ns, resource string) tea.Cmd {
	args := []string{"edit", resource, "--context", k.kubectx}
	if ns != "" {
		args = append(args, "-n", ns)
	}
	return k.shell.Exec("kubectl", args...)
}

func (k *Cmd) Describe(ns, resource string) tea.Cmd {
	args := []string{"describe", resource, "--context", k.kubectx}
	if ns != "" {
		args = append(args, "-n", ns)
	}
	return k.shell.Pager("kubectl", args...)
}

func (k *Cmd) Get(ns, resource string) tea.Cmd {
	args := []string{"get", resource, "-o", "yaml", "--context", k.kubectx}
	if ns != "" {
		args = append(args, "-n", ns)
	}
	return k.shell.Pager("kubectl", args...)
}

func (k *Cmd) Delete(ns, resource string) tea.Cmd {
	args := []string{"delete", resource, "--context", k.kubectx}
	if ns != "" {
		args = append(args, "-n", ns)
	}
	return k.shell.Exec("kubectl", args...)
}
