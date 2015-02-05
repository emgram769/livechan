package main

import (
  "time"
)

type Message struct {
  data []byte
  conn *Connection
}

type Hub struct {
  channels map[string]map[*Connection]time.Time
  broadcast chan Message
  register chan *Connection
  unregister chan *Connection
}

var h = Hub {
  broadcast: make(chan Message),
  register: make(chan *Connection),
  unregister: make(chan *Connection),
  channels: make(map[string]map[*Connection]time.Time),
}

func (h *Hub) run() {
  for {
    select {
    case c := <-h.register:
      if (h.channels[c.channelName] == nil) {
        h.channels[c.channelName] = make(map[*Connection]time.Time)
      }
      h.channels[c.channelName][c] = time.Unix(0,0)
	    c.send <- createJSONs(storage.getChats(c.channelName, "General", 50), c)
    case c := <-h.unregister:
      if _, ok := h.channels[c.channelName][c]; ok {
        delete(h.channels[c.channelName], c)
        close(c.send)
      }
    case m := <-h.broadcast:
      var chat = createChat(m.data, m.conn);
      if (chat.canBroadcast(m.conn)) {
        for c := range h.channels[m.conn.channelName] {
          select {
          case c.send <- chat.createJSON(c):
          default:
            close(c.send)
            delete(h.channels[m.conn.channelName], c)
          }
        }
        storage.insertChat(m.conn.channelName, *chat)
      }
    }
  }
}

