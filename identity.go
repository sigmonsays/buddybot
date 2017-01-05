package buddybot

import (
	"fmt"
	"os"
	"os/user"
	"strings"
)

type Identity struct {
	IP       string
	Hostname string
	Username string
	Nick     string
}

func ParseIdentity(identity string) (*Identity, error) {
	i := NewIdentity()

	tmp1 := strings.Split(identity, "@") // username @ host / ip / nick
	if len(i.Username) > 0 {
		i.Username = tmp1[0]
	}
	if len(tmp1) > 1 {
		tmp2 := strings.Split(tmp1[1], "/") // host / ip / nick
		if len(tmp2) > 0 {
			i.Hostname = tmp2[0]
		}
		if len(tmp2) > 1 {
			i.IP = tmp2[1]
		}
		if len(tmp2) > 2 {
			i.Nick = tmp2[2]
		}
	}

	return i, nil
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
	return fmt.Sprintf("%s@%s/%s/%s", me.Username, me.Hostname, me.IP, me.Nick)
}
