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

type hub struct {
  connections map[*connection]bool
  broadcast chan []byte
  register chan *connection
  unregister chan *connection
}

var h = hub {
  broadcast: make(chan []byte),
  register: make(chan *connection),
  unregister: make(chan *connection),
  connections: make(map[*connection]bool),
}

func (h *hub) run() {
  for {
    select {
    case c := <-h.register:
      h.connections[c] = true
    case c := <-h.unregister:
      if _, ok := h.connections[c]; ok {
        delete(h.connections, c)
        close(c.send)
      }
    case m := <-h.broadcast:
      for c := range h.connections {
        select {
        case c.send <- m:
        default:
          close(c.send)
          delete(h.connections, c)
        }
      }
    }
  }
}

var upgrader = websocket.Upgrader{
  ReadBufferSize: 1024,
  WriteBufferSize: 1024,
}

type connection struct {
  ws *websocket.Conn
  send chan []byte
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
    _, message, err := c.ws.ReadMessage()
    if err != nil {
      break
    }
    h.broadcast <- message
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
  h.register <- c
  go c.writePump()
  c.readPump()
}

func htmlServer(w http.ResponseWriter, req *http.Request) {
  if req.URL.Path != "/" {
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
  http.HandleFunc("/ws", wsServer)
  http.HandleFunc("/static/", staticServer)
  err := http.ListenAndServe(":8080", nil)
  if err != nil {
    log.Fatal("ListenAndServ: ", err)
  }
}

