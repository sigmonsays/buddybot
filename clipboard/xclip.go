package clipboard

import (
	"os/exec"

	"github.com/sigmonsays/buddybot/util"
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
	return util.StdinCommand(cmdline, s)
}

func (me *XClip) GetString() (string, error) {
	cmdline := []string{me.path, "-o"}
	out, err := util.StdoutCommand(cmdline)
	if err != nil {
		return out, err
	}
	log.Debugf("GetString: %q", out)
	return out, nil
}
