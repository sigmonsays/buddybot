package buddybot

import (
	"bytes"
	"encoding/json"
)

type ClientList struct {
	List []*Client
}
type Client struct {
	Id       int64
	Identity string
}

func NewClientList() *ClientList {
	cl := &ClientList{
		List: make([]*Client, 0),
	}
	return cl
}

func (me *ClientList) AddClient(c *Connection) {
	cl := &Client{
		Id:       c.id,
		Identity: c.Identity,
	}
	me.List = append(me.List, cl)
}

func (me *ClientList) ToJson() []byte {
	buf, err := json.Marshal(me.List)
	if err != nil {
		log.Warnf("ToJson: %s", err)
	}
	return buf
}
func (me *ClientList) FromJsonString(buf string) error {
	bin := bytes.NewBufferString(buf).Bytes()
	err := json.Unmarshal(bin, &me.List)
	return err
}
