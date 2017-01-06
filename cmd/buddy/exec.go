package main

import (
	"os"
	"os/exec"
)

func NewShellExec() *ShellExec {
	se := &ShellExec{}
	return se
}

type ShellExec struct {
}

type ExecResult struct {
	Pid      int
	ExitCode int
}

// handles a exec command from the CLI

func (me *ShellExec) ExecMessage(args []string) (*ExecResult, error) {

	log.Debugf("exec line %q", args)

	res := &ExecResult{}

	cmd := exec.Command(args[0], args[1:]...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		return nil, err
	}

	res.Pid = cmd.Process.Pid

	// res.ExitCode todo

	return res, nil
}
