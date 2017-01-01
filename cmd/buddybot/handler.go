package main

import (
	"container/list"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/sigmonsays/buddybot"
)

type chatHandler struct {
	mx        sync.Mutex
	staticDir string
	history   *list.List
}

func (h *chatHandler) serveHome(w http.ResponseWriter, r *http.Request) {
	log.Printf("request %s", r.URL)
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	home_html := filepath.Join(h.staticDir, "home.html")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	homeTempl := template.Must(template.ParseFiles(home_html))
	homeTempl.Execute(w, r.Host)
}

func (h *chatHandler) addHistory(m *buddybot.Message) {
	h.mx.Lock()
	defer h.mx.Unlock()
	h.history.PushBack(m)
	if h.history.Len() > 5 {
		if e := h.history.Front(); e != nil {
			h.history.Remove(e)
		}
	}
}

func (h *chatHandler) getHistory() []*buddybot.Message {
	h.mx.Lock()
	defer h.mx.Unlock()
	ret := make([]*buddybot.Message, 0)
	for e := h.history.Front(); e != nil; e = e.Next() {
		ret = append(ret, e.Value.(*buddybot.Message))
	}
	return ret
}

func (h *chatHandler) handleMessage(op buddybot.OpCode, hub *buddybot.Hub, c *buddybot.Connection, m *buddybot.Message) error {
	log.Printf("handleMessage op:%s\n", op)

	if op == buddybot.MessageOp {
		hub.SendBroadcast(m)

	} else if op == buddybot.RegisterOp {
		// play back history
		for _, hm := range h.getHistory() {
			hm.Op = HistoryOp
			hub.SendMessage(c, hm)
		}
	} else if op == buddybot.UnregisterOp {
		hub.Send(buddybot.NoticeOp, fmt.Sprintf("%s has left", c.Name))
		//} else if m.Op == HistoryOp {
		//		hub.sendBroadcast(m)

	} else if op == buddybot.NickOp {

		if c.Name == "" {
			hub.Send(buddybot.NoticeOp, fmt.Sprintf("%s has joined", m.From))
		} else {
			hub.Send(buddybot.NoticeOp, fmt.Sprintf("%s has changed their name to %s", c.Name, m.From))
		}
		c.Name = m.From

	} else if op == buddybot.NoticeOp {
		hub.SendBroadcast(m)

	} else if op == buddybot.ClientListOp {
		hub.SendClientList(m.Id)

	} else {
		log.Printf("Unhandled op %+v\n", m)
	}
	return nil
}
