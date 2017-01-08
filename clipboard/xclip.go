package clipboard

import (
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
	return StdinCommand(cmdline, s)
}

func (me *XClip) GetString() (string, error) {
	cmdline := []string{me.path, "-o"}
	out, err := StdoutCommand(cmdline)
	if err != nil {
		return out, err
	}
	log.Debugf("GetString: %q", out)
	return out, nil
}
