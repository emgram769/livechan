package main

import (
  "github.com/gorilla/websocket"
  "net/http"
  "time"
  "log"
)

const (
  writeWait = 10 * time.Second
  pongWait = 60 * time.Second
  pingPeriod = (pongWait * 9) / 10
  maxMessageSize = 512
)

type connection struct {
  ws *websocket.Conn
  send chan []byte
  channelName string
  ipAddr string
}

type message struct {
  data []byte
  conn *connection
}

type hub struct {
  channels map[string]map[*connection]time.Time
  broadcast chan message
  register chan *connection
  unregister chan *connection
}

var h = hub {
  broadcast: make(chan message),
  register: make(chan *connection),
  unregister: make(chan *connection),
  channels: make(map[string]map[*connection]time.Time),
}

func canBroadcast(t time.Time) bool{
  if (time.Now().Sub(t).Seconds() < 4) {
    return false
  }
  return true
}

func (h *hub) run() {
  for {
    select {
    case c := <-h.register:
      if (h.channels[c.channelName] == nil) {
        h.channels[c.channelName] = make(map[*connection]time.Time)
      }
      h.channels[c.channelName][c] = time.Unix(0,0)
    case c := <-h.unregister:
      if _, ok := h.channels[c.channelName][c]; ok {
        delete(h.channels[c.channelName], c)
        close(c.send)
      }
    case m := <-h.broadcast:
      if (canBroadcast(h.channels[m.conn.channelName][m.conn])) {
        h.channels[m.conn.channelName][m.conn] = time.Now()
        for c := range h.channels[m.conn.channelName] {
          select {
          case c.send <- m.data:
          default:
            close(c.send)
            delete(h.channels[m.conn.channelName], c)
          }
        }
      }
    }
  }
}

var upgrader = websocket.Upgrader{
  ReadBufferSize: 1024,
  WriteBufferSize: 1024,
}

func (c *connection) readPump() {
  defer func() {
    h.unregister <- c
    c.ws.Close()
  }()
  c.ws.SetReadLimit(maxMessageSize)
  c.ws.SetReadDeadline(time.Now().Add(pongWait))
  c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
  for {
    _, d, err := c.ws.ReadMessage()
    if err != nil {
      break
    }
    m := message{data:d, conn:c}
    h.broadcast <- m
  }
}

func (c *connection) write(mt int, payload []byte) error {
  c.ws.SetWriteDeadline(time.Now().Add(writeWait))
  return c.ws.WriteMessage(mt, payload)
}

func (c *connection) writePump() {
  ticker := time.NewTicker(pingPeriod)
  defer func() {
    ticker.Stop()
    c.ws.Close()
  }()
  for {
    select {
    case message, ok := <-c.send:
      if !ok {
        c.write(websocket.CloseMessage, []byte{})
        return
      }
      if err := c.write(websocket.TextMessage, message); err != nil {
        return
      }
    case <-ticker.C:
      if err := c.write(websocket.PingMessage, []byte{}); err != nil {
        return
      }
    }
  }
}

func wsServer(w http.ResponseWriter, req *http.Request) {
  if req.Method != "GET" {
    http.Error(w, "Method not allowed", 405)
    return
  }
  ws, err := upgrader.Upgrade(w, req, nil)
  if err != nil {
    log.Println(err)
    return
  }
  c := &connection{send: make(chan []byte, 256), ws: ws}
  c.channelName = req.URL.Path[1:]
  c.ipAddr = req.RemoteAddr
  h.register <- c
  go c.writePump()
  c.readPump()
}

func htmlServer(w http.ResponseWriter, req *http.Request) {
  if req.URL.Path != "/" {
    http.Error(w, "Method not allowed", 405)
    return
  }
  if req.Method != "GET" {
    http.Error(w, "Method not allowed", 405)
    return
  }
  w.Header().Set("Content-Type", "text/html; charset=utf-8")
  http.ServeFile(w, req, "index.html")
}

func staticServer(w http.ResponseWriter, req *http.Request) {
    http.ServeFile(w, req, req.URL.Path[1:])
}

func main() {
  go h.run()
  http.HandleFunc("/", htmlServer)
  http.HandleFunc("/ws/", wsServer)
  http.HandleFunc("/static/", staticServer)
  err := http.ListenAndServe(":8080", nil)
  if err != nil {
    log.Fatal("ListenAndServ: ", err)
  }
}

