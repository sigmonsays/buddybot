package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sigmonsays/buddybot"
)

func main() {

	addr := "chat.grepped.org"
	path := "/chat/ws"

	flag.StringVar(&addr, "addr", addr, "address")
	flag.StringVar(&path, "path", path, "path")
	flag.Parse()

	state := &state{
		addr:     addr,
		path:     path,
		identity: buddybot.NewIdentity(),
	}
	log.Infof("Identity %s", state.identity.String())

	u := url.URL{Scheme: "ws", Host: addr, Path: path}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Errorf("dial: %s", err)
	}

	state.c = c
	defer c.Close()

	state.loop()
}

type state struct {
	addr      string
	path      string
	identity  *buddybot.Identity
	interrupt chan os.Signal

	c *websocket.Conn
}

func (me *state) NewMessage() *buddybot.Message {
	m := &buddybot.Message{}
	m.Id = 1
	m.Op = buddybot.MessageOp
	m.From = me.identity.String()
	return m
}

func (me *state) loop() error {

	c := me.c
	done := make(chan struct{})

	// startup the receive loop
	go func() {
		defer c.Close()
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Infof("read: %s", err)
				return
			}
			me.receiveMessage(message)
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	// signal to interrupt
	me.interrupt = make(chan os.Signal, 1)
	signal.Notify(me.interrupt, os.Interrupt)

	// read from stdin for commands
	lines := make(chan string, 2)
	go func() {
		rdr := bufio.NewReader(os.Stdin)
		for {
			line, err := rdr.ReadString('\n')
			if err != nil {
				log.Infof("Read: %s", err)
				break
			}
			lines <- line
		}
	}()

	// send a join message
	j := me.NewMessage()
	j.Op = buddybot.JoinOp
	err := c.WriteMessage(websocket.TextMessage, j.Json())
	if err != nil {
		log.Infof("join: write: %s", err)
	}

	// start a ping loop
	go func() {
		d := time.Duration(10) * time.Second
		t := time.NewTicker(d)
		for {
			select {
			case <-t.C:
				p := me.NewMessage()
				p.Op = buddybot.PingOp
				err := c.WriteMessage(websocket.TextMessage, p.Json())
				if err != nil {
					log.Infof("ping: write: %s", err)
				}
			}
		}
	}()

	// just wait on interrupt
	for {
		select {
		case line := <-lines:

			msg := me.NewMessage()
			msg.Message = line
			buf, err := json.Marshal(msg)
			if err != nil {
				log.Infof("Marshal: %s", err)
				continue
			}
			log.Infof(">> %s", buf)
			me.sendMessage(buf)

		case <-me.interrupt:
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Infof("message: write close: %s", err)
				return err
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			c.Close()
			return nil
		}
	}
}

func (me *state) receiveMessage(msg []byte) error {
	log.Infof("receiveMessage: %s", msg)
	return nil
}

func (me *state) sendMessage(msg []byte) error {
	err := me.c.WriteMessage(websocket.TextMessage, msg)
	return err
}
