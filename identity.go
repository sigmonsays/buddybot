package buddybot

import (
	"fmt"
	"os"
	"os/user"
)

type Identity struct {
	IP       string
	Hostname string
	Username string
}

func NewIdentity() *Identity {

	ip, err := ExternalIP()
	if err != nil {
		log.Warnf("ExternalIP: %s", err)
	}

	hostname, err := os.Hostname()
	if err != nil {
		log.Warnf("Hostname: %s", err)
		hostname = "uknown"
	}

	ident := &Identity{
		IP:       ip,
		Hostname: hostname,
		Username: hostname,
	}

	u, err := user.Current()
	if err == nil {
		ident.Username = u.Username
	} else {
		log.Warnf("User: %s", err)
	}
	return ident
}

func (me *Identity) String() string {
	return fmt.Sprintf("%s@%s/%s", me.Username, me.Hostname, me.IP)
}
