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
	me.List[c.id] = c.Name
}

func (me *ClientList) ToJson() []byte {
	buf, _ := json.Marshal(me)
	return buf
}
