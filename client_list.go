package buddybot

import (
	"encoding/json"
)

type ClientList struct {
	List map[int64]string
}

func NewClientList() *ClientList {
	cl := &ClientList{
		List: make(map[int64]string, 0),
	}
	return cl
}

func (me *ClientList) AddClient(c *Connection) {
	me.List[c.id] = c.Identity
}

func (me *ClientList) ToJson() []byte {
	buf, err := json.Marshal(me)
	if err != nil {
		log.Warnf("ToJson: %s", err)
	}
	return buf
}
