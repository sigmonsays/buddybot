package main

//go:generate stringer -type=ConnState

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/facebookgo/devrestarter"
	"github.com/gorilla/websocket"
	"github.com/sigmonsays/buddybot"
	"gopkg.in/natefinch/lumberjack.v2"

	gologging "github.com/sigmonsays/go-logging"
)

type ConnState int

const (
	Unknown ConnState = iota
	Connected
	Disconnected
)

func main() {

	path := "/chat/ws"
	verbose := false

	conf := GetDefaultConfig()

	flag.StringVar(&conf.ServerAddress, "addr", conf.ServerAddress, "address of buddyd websocket server")
	flag.StringVar(&path, "path", path, "path")
	flag.StringVar(&conf.Nick, "nick", conf.Nick, "nick name")
	flag.StringVar(&conf.LogLevel, "log", conf.LogLevel, "log level")
	flag.BoolVar(&verbose, "verbose", verbose, "be verbose")
	flag.Parse()

	// load home config
	datadir := filepath.Join(os.Getenv("HOME"), ".buddy")
	user_conf := filepath.Join(datadir, "buddy.yaml")
	st, err := os.Stat(user_conf)
	if err == nil && st.IsDir() == false {

		err = conf.LoadYaml(user_conf)
		if err != nil {
			StartupError("LoadYaml %s", err)
		}

	}

	// generate a nick if its empty
	if conf.Nick == "" {
		conf.Nick = fmt.Sprintf("user-%d", time.Now().Unix())
		log.Warnf("nick name not set, using generated %s", conf.Nick)
	}

	// setup lumberjack logging
	filelog_name := filepath.Join(datadir, "buddy.log")
	ljack := &lumberjack.Logger{
		Filename:   filelog_name,
		MaxSize:    500, // megabytes
		MaxBackups: 3,
		MaxAge:     28, //days
	}

	gologging.SetLogLevel(conf.LogLevel)

	if verbose {
		conf.PrintYaml()
		for name, level := range conf.VerboseLogLevels {
			gologging.SetLevel(name, level)
		}

	} else {

		gologging.SetLogOutput(ljack)

		for name, level := range conf.LogLevels {
			gologging.SetLevel(name, level)
		}
	}

	if conf.GitWatch {

		gopath := os.Getenv("GOPATH")
		if gopath == "" {
			StartupError("GOPATH must be set for git_watch to work")
		}

		git_path, err := exec.LookPath("git")
		if err != nil {
			StartupError("can not find git command: %s", err)
		}
		log.Debugf("git path %s", git_path)

		devrestarter.Init()
		go GitWatch(conf)
	}

	identity := buddybot.NewIdentity()
	handler := NewHandler(identity)
	context := &Context{
		Identity: identity,
	}
	state := &state{
		addr:      conf.ServerAddress,
		path:      path,
		identity:  identity,
		connstate: make(chan ConnState, 10),
		verbose:   verbose,
		handler:   handler,
		context:   context,

		reconnectDelay: conf.ReconnectDelay,
	}

	state.identity.Nick = conf.Nick
	log.Infof("Identity %s", state.identity.String())

	// signal to interrupt
	state.interrupt = make(chan os.Signal, 1)
	signal.Notify(state.interrupt, os.Interrupt)

	// read from clients
	srv, err := net.Listen("tcp", conf.BuddyServer)
	if err != nil {
		StartupError("Listen %s", err)
	}

	LineServer := func(con net.Conn, lines chan string) {
		defer con.Close()

		bio := bufio.NewReader(con)
		for {
			line, err := bio.ReadBytes('\n')
			if err != nil {
				if err == io.EOF {
					break
				}
				log.Warnf("ReadBytes: %s", err)
				break
			}
			sline := string(line)
			if sline == ".\n" {
				break
			}
			lines <- sline
		}
	}

	go func() {
		for {
			con, err := srv.Accept()
			if err != nil {
				log.Warnf("Accept: %s", err)
				continue
			}
			go LineServer(con, state.lines)
		}
	}()

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
	verbose   bool
	addr      string
	path      string
	identity  *buddybot.Identity
	interrupt chan os.Signal
	connstate chan ConnState
	lines     chan string

	c *buddybot.Connection

	handler *handler
	context *Context

	reconnectDelay time.Duration
}

func (me *state) loop() error {
	var err error

	for {
		err = me.ioloop()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Warnf("Connection failed; %s: delay %s", err, me.reconnectDelay)
			time.Sleep(me.reconnectDelay)
		}
	}

	return err
}

func (me *state) introduction() error {
	// send a join message
	j := me.context.NewMessage()
	j.Op = buddybot.JoinOp
	j.Message = "JOIN EVENT"
	err := me.c.WriteMessage(websocket.TextMessage, j.Json())
	if err != nil {
		log.Infof("join: write: %s", err)
	}
	log.Debugf("join message sent - %s", j)

	// see who is online
	l := me.context.NewMessage()
	l.Op = buddybot.ClientListOp
	err = me.c.WriteMessage(websocket.TextMessage, l.Json())
	if err != nil {
		log.Infof("clientList: write: %s", err)
	}
	return nil
}

func (me *state) ioloop() error {

	// establish connection
	u := url.URL{Scheme: "ws", Host: me.addr, Path: me.path}

	log.Infof("Connecting to %s", u.String())
	wsconn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Debugf("dial: %s", err)
		return err
	}

	hub := buddybot.NewHub()
	for _, op := range buddybot.OpCodes() {
		hub.OnCallback(op, me.receiveMessage)
	}
	go hub.Start()

	id := int64(1)
	c := buddybot.NewConnection(hub, id, wsconn)
	c.Start()
	hub.Register(c)

	me.c = c
	me.context.Conn = c

	me.introduction()

	log.Infof("Connecton established")

	done := make(chan struct{})

	// just wait on interrupt
	for {
		select {

		case cstate := <-me.connstate:
			if cstate == Disconnected {
				log.Infof("disconnected.")
				return fmt.Errorf("Disconnected")
			}

		case line := <-me.lines:

			line = strings.TrimRight(line, "\n")
			if len(line) == 0 {
				continue
			}

			msg := me.context.NewMessage()
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
			err := me.c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
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

func (me *state) receiveMessage(op buddybot.OpCode, hub *buddybot.Hub, c *buddybot.Connection, m *buddybot.Message) error {

	if m == nil {
		m = buddybot.NewMessage()
		m.Op = op
	}

	if me.verbose {
		log.Tracef("receiveMessage bytes %#v", m)
	}

	if m.Op == buddybot.JoinOp {
		log.Infof("JOIN from=%s cid=%d", m.From, m.Id)
		// me.handler.OnJoin(m)

	} else if m.Op == buddybot.RawMessageOp {
		me.handler.OnMessage(m, me.context)

	} else if m.Op == buddybot.UnregisterOp {
		me.connstate <- Disconnected
		return io.EOF

	} else if m.Op == buddybot.MessageOp {
		me.handler.OnMessage(m, me.context)

	} else if m.Op == buddybot.DirectMessageOp {
		me.handler.OnMessage(m, me.context)

	} else if m.Op == buddybot.NoticeOp {
		me.handler.OnNotice(m, me.context)

	} else if m.Op == buddybot.ClientListOp {
		// me.handler.OnClientList(m)

		cl := buddybot.NewClientList()
		err := cl.FromJsonString(m.Message)
		if err != nil {
			log.Warnf("ClientList: %s", err)
			return nil
		}
		fmt.Printf("Connected Clients:\n")
		for _, c := range cl.List {
			id, err := buddybot.ParseIdentity(c.Identity)
			if err != nil {
				log.Debugf("ParseIdentity %q: %s", c.Identity, err)
				id = &buddybot.Identity{}
			}
			fmt.Printf(" client cid=%-4s %-25s %s\n", strconv.FormatInt(c.Id, 10), id.Nick, id)
		}
		fmt.Printf("\n")

	} else {
		log.Tracef("receiveMessage: %s", m)
	}

	return nil
}

func (me *state) sendMessage(msg []byte) error {
	log.Tracef("sendMessage %s", msg)
	err := me.c.WriteMessage(websocket.TextMessage, msg)
	return err
}
