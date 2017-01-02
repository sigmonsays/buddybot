package main

import (
	"bufio"
	"io"
	"os/exec"
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
	Done     chan error

	Stdout   chan string
	Stderr   chan string
	finished chan bool

	pending int
}

func (me *ExecResult) wait() error {
	err := me.cmd.Wait()
	if err == nil {
		log.Debugf("Wait returned success")
	} else {
		log.Debugf("Wait returned %s", err)
	}
	return err
}

func (me *ExecResult) process() {
	var err error

Dance:
	for {
		select {

		case <-me.finished:
			me.pending--
			log.Debugf("finished: pending=%d", me.pending)
			if me.pending == 0 {
				break Dance
			}
		}
	}

	err = me.wait()

	log.Tracef("completly finished: sending Done signal")
	me.Done <- err

	return
}

func stream(input io.Reader, out chan string, finished chan bool) {

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
		log.Tracef("line: (%d) %q", len(line), line)
		out <- string(line)
	}

	log.Tracef("stream finished")
	finished <- true
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
		cmd:      cmd,
		Stdout:   make(chan string, 0),
		Stderr:   make(chan string, 0),
		finished: make(chan bool, 3),
		pending:  2,
		Done:     make(chan error, 0),
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return res, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return res, err
	}

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	go stream(stdout, res.Stdout, res.finished)
	go stream(stderr, res.Stderr, res.finished)

	res.Pid = cmd.Process.Pid

	go res.process()

	return res, nil
}
