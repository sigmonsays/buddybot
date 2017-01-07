package main

import (
	"bufio"
	"io"
	"os/exec"
	"sync"
)

var NL = byte('\n')

func NewShellExec() *ShellExec {
	se := &ShellExec{}
	return se
}

type ShellExec struct {
}

type ExecResult struct {
	cmd *exec.Cmd

	Pid      int
	ExitCode int

	errResult chan error
	Stdout    chan string
	Stderr    chan string

	wg sync.WaitGroup
}

func (me *ExecResult) process() {
	me.wg.Add(1)
	defer me.wg.Done()
	err := me.cmd.Wait()

	me.errResult <- err
}

func (me *ExecResult) Wait() error {
	return <-me.errResult
}

func stream(input io.Reader, out chan string, wg sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	bio := bufio.NewReader(input)
	for {
		line, err := bio.ReadBytes(NL)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Warnf("stream: ReadBytes: %s", err)
			break
		}
		out <- string(line)
	}
	close(out)
}

// handles a exec command from the CLI

func (me *ShellExec) ExecMessage(args []string) (*ExecResult, error) {

	log.Debugf("exec line %q", args)

	var cmd *exec.Cmd

	if len(args) > 1 {
		cmd = exec.Command(args[0], args[1:]...)
	} else {
		cmd = exec.Command(args[0])
	}

	res := &ExecResult{
		cmd:       cmd,
		errResult: make(chan error, 1),
		Stdout:    make(chan string, 0),
		Stderr:    make(chan string, 0),
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return res, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return res, err
	}

	go stream(stdout, res.Stdout, res.wg)
	go stream(stderr, res.Stderr, res.wg)

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	go res.process()

	res.Pid = cmd.Process.Pid

	// res.ExitCode todo

	return res, nil
}
