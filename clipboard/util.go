package clipboard

import (
	"bytes"
	"os/exec"
)

func StdinCommand(cmdline []string, stdin string) error {
	cmd := exec.Command(cmdline[0], cmdline[1:]...)
	cmd.Stdin = bytes.NewBufferString(stdin)
	err := cmd.Start()
	if err != nil {
		return err
	}
	err = cmd.Wait()
	if err != nil {
		return err
	}
	return nil
}

func StdoutCommand(cmdline []string) (string, error) {
	cmd := exec.Command(cmdline[0], cmdline[1:]...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	s := string(out)
	return s, nil
}
