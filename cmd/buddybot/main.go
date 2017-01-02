package main

import (
	"container/list"
	"flag"
	"net/http"
	"os"
	"path/filepath"

	"github.com/sigmonsays/buddybot"

	gologging "github.com/sigmonsays/go-logging"
)

const (
	HistoryOp = iota + 100
)

var addr = flag.String("addr", ":8080", "http service address")

func main() {

	gopath := os.Getenv("GOPATH")
	var staticDir string
	loglevel := "warn"
	flag.StringVar(&staticDir, "static", "", "location of static data")
	flag.StringVar(&loglevel, "log", loglevel, "change log level")

	if staticDir == "" && gopath != "" {
		staticDir = filepath.Join(gopath, "src/github.com/sigmonsays/buddybot/static")
	}

	flag.Parse()

	gologging.SetLogLevel(loglevel)

	hub := buddybot.NewHub()

	srv, err := buddybot.NewHandler(hub)
	if err != nil {
		log.Errorf("NewHandler: ", err)
	}

	handler := &chatHandler{
		staticDir: staticDir,
		history:   list.New(),
	}

	opcodes := []buddybot.OpCode{
		buddybot.RegisterOp,
		buddybot.UnregisterOp,
		buddybot.NoticeOp,
		buddybot.NickOp,
		buddybot.MessageOp,
	}
	for _, op := range opcodes {
		hub.OnCallback(op, handler.handleMessage)
	}
	go hub.Start()

	mx := http.NewServeMux()

	log.Infof("serving static data from %s", staticDir)

	mx.HandleFunc("/", handler.serveHome)
	mx.HandleFunc("/ws", srv.ServeWebSocket)
	mx.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))

	alias := "/chat"
	mx.HandleFunc(alias, handler.serveHome)
	mx.HandleFunc(alias+"/ws", srv.ServeWebSocket)
	mx.Handle(alias+"/static/", http.StripPrefix(alias+"/static/", http.FileServer(http.Dir(staticDir))))

	hs := &http.Server{
		Addr:    *addr,
		Handler: mx,
	}

	err = hs.ListenAndServe()
	if err != nil {
		log.Errorf("ListenAndServe: ", err)
	}
}
