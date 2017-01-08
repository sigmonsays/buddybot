package clipboard

import (
	"os/exec"

	"github.com/sigmonsays/buddybot/util"
)

func NewPBClip() *PBClip {
	pbcopy, err := exec.LookPath("pbcopy")
	if err != nil {
		log.Warnf("clipboard functionality probably wont work")
	}
	pbpaste, err := exec.LookPath("pbpaste")
	if err != nil {
		log.Warnf("clipboard functionality probably wont work")
	}

	return &PBClip{
		pbcopy:  pbcopy,
		pbpaste: pbpaste,
	}
}

type PBClip struct {
	pbcopy  string
	pbpaste string
}

func (me *PBClip) SetString(s string) error {
	log.Debugf("SetString: %q", s)
	cmdline := []string{me.pbcopy, "-i"}
	return util.StdinCommand(cmdline, s)
}

func (me *PBClip) GetString() (string, error) {
	cmdline := []string{me.pbpaste, "-o"}
	out, err := util.StdoutCommand(cmdline)
	if err != nil {
		return out, err
	}
	log.Debugf("GetString: %q", out)
	return out, nil
}
