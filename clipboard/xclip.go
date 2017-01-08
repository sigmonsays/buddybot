package clipboard

import (
	"bytes"
	"os/exec"
)

func NewXClip() *XClip {

	path, err := exec.LookPath("xclip")
	if err != nil {
		log.Warnf("clipboard functionality probably wont work")
	}

	return &XClip{
		path: path,
	}
}

type XClip struct {
	path string
}

func (me *XClip) SetString(s string) error {
	log.Debugf("SetString: %q", s)
	cmdline := []string{me.path, "-i"}
	cmd := exec.Command(cmdline[0], cmdline[1:]...)
	cmd.Stdin = bytes.NewBufferString(s)
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

func (me *XClip) GetString() (string, error) {
	cmdline := []string{me.path, "-o"}
	cmd := exec.Command(cmdline[0], cmdline[1:]...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	s := string(out)
	log.Debugf("GetString: %q", s)
	return s, nil
}
