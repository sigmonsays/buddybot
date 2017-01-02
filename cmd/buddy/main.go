package main

//go:generate stringer -type=ConnState

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sigmonsays/buddybot"

	gologging "github.com/sigmonsays/go-logging"
)

type ConnState int

const (
	Unknown ConnState = iota
	Connected
	Disconnected
)

func main() {

	addr := "localhost"
	path := "/chat/ws"
	nick := "default"
	loglevel := "info"

	flag.StringVar(&addr, "addr", addr, "address")
	flag.StringVar(&path, "path", path, "path")
	flag.StringVar(&nick, "nick", nick, "nick name")
	flag.StringVar(&loglevel, "log", loglevel, "log level")
	flag.Parse()

	gologging.SetLogLevel(loglevel)

	state := &state{
		addr:      addr,
		path:      path,
		identity:  buddybot.NewIdentity(),
		connstate: make(chan ConnState, 10),
	}

	state.identity.Nick = nick
	log.Infof("Identity %s", state.identity.String())

	// signal to interrupt
	state.interrupt = make(chan os.Signal, 1)
	signal.Notify(state.interrupt, os.Interrupt)

	// read from stdin for commands
	state.lines = make(chan string, 2)
	go func() {
		rdr := bufio.NewReader(os.Stdin)
		for {
			line, err := rdr.ReadString('\n')
			if err != nil {
				log.Infof("Read: %s", err)
				break
			}
			state.lines <- line
		}
	}()

	state.loop()
}

type state struct {
	addr      string
	path      string
	identity  *buddybot.Identity
	interrupt chan os.Signal
	connstate chan ConnState
	lines     chan string

	c *websocket.Conn
}

func (me *state) NewMessage() *buddybot.Message {
	m := &buddybot.Message{}
	m.Id = 0
	m.Op = buddybot.MessageOp
	m.From = me.identity.String()
	return m
}

func (me *state) loop() error {
	var err error

	for {
		err = me.ioloop()
		if err == io.EOF {
			break
		}
		if err != nil {
			time.Sleep(time.Duration(1) * time.Second)
		}
	}

	return err
}

func (me *state) ioloop() error {

	// establish connection
	u := url.URL{Scheme: "ws", Host: me.addr, Path: me.path}

	log.Infof("Connecting to %s", u.String())
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Errorf("dial: %s", err)
		return err
	}

	me.c = c

	log.Infof("Connecton established")

	done := make(chan struct{})

	// startup the receive loop
	go func() {
		log.Tracef("receive loop started")
		defer c.Close()
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Warnf("read: %s", err)
				me.connstate <- Disconnected
				return
			}
			err = me.receiveMessage(message)
			if err != nil {
				log.Warnf("receiveMessage: %s", err)
			}
		}
		log.Tracef("receive loop exited")
	}()

	// send a join message
	j := me.NewMessage()
	j.Op = buddybot.JoinOp
	j.Message = "JOIN EVENT"
	err = c.WriteMessage(websocket.TextMessage, j.Json())
	if err != nil {
		log.Infof("join: write: %s", err)
	}
	log.Debugf("join message sent - %s", j)

	// see who is online
	l := me.NewMessage()
	l.Op = buddybot.ClientListOp
	err = c.WriteMessage(websocket.TextMessage, l.Json())
	if err != nil {
		log.Infof("clientList: write: %s", err)
	}

	// start a ping loop
	go func() {
		log.Debugf("ping loop started")
		d := time.Duration(10) * time.Second
		t := time.NewTicker(d)
		defer t.Stop()
	Loop:
		for {
			select {
			case <-t.C:
				p := me.NewMessage()
				p.Op = buddybot.PingOp
				err := c.WriteMessage(websocket.TextMessage, p.Json())
				if err != nil {
					log.Infof("ping: write: %s", err)
					me.connstate <- Disconnected
					break Loop
				}
			}
			me.connstate <- Connected
		}
		log.Debugf("ping loop exited")
	}()

	// just wait on interrupt
	for {
		select {
		case cstate := <-me.connstate:
			if cstate == Disconnected {
				log.Tracef("disconnected.")
				return fmt.Errorf("Disconnected")
			}

		case line := <-me.lines:

			line = strings.TrimRight(line, "\n")
			if len(line) == 0 {
				continue
			}

			msg := me.NewMessage()
			msg.Message = line
			buf, err := json.Marshal(msg)
			if err != nil {
				log.Infof("Marshal: %s", err)
				continue
			}
			log.Tracef("sendMessage: %s", msg)
			me.sendMessage(buf)

		case <-me.interrupt:
			log.Infof("Interrupt received..")
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Infof("message: write close: %s", err)
				return err
			}
			select {
			case <-done:
				log.Infof("done received from interrupt")
			case <-time.After(time.Second):
				log.Infof("timeout from interrupt")
			}
			return io.EOF
		}
	}

	log.Infof("Why here?")
	return nil
}

func (me *state) receiveMessage(msg []byte) error {
	log.Tracef("receiveMessage bytes %s", msg)
	m := me.NewMessage()
	err := json.Unmarshal(msg, m)
	if err != nil {
		return err
	}

	if m.Op == buddybot.JoinOp {
		log.Infof("JOIN from=%s", m.From)

	} else if m.Op == buddybot.MessageOp {
		fmt.Printf("MESSAGE <%s> %s\n", m.From, m.Message)

	} else {
		log.Tracef("receiveMessage: %s", m)
	}

	return nil
}

func (me *state) sendMessage(msg []byte) error {
	err := me.c.WriteMessage(websocket.TextMessage, msg)
	return err
}
