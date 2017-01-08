package keyboard

import (
	"github.com/sigmonsays/buddybot/util"
)

type Keyboard struct {
}

func (me *Keyboard) ScrollLock(enable bool) error {
	var cmdline []string

	if enable {
		cmdline = []string{"xset", "led", "named", "Scroll Lock"}
	} else {
		cmdline = []string{"xset", "-led", "named", "Scroll Lock"}
	}

	out, err := util.StdoutCommand(cmdline)
	if err != nil {
		log.Debugf("ScrollLock %s: %s (error %s)", cmdline, out, err)
	}
	return err
}
