package main

import (
  "github.com/gorilla/websocket"
  "net/http"
  "time"
  "log"
  "fmt"
  "strings"
)

const (
  writeWait = 10 * time.Second
  pongWait = 60 * time.Second
  pingPeriod = (pongWait * 9) / 10
  maxMessageSize = 512
)

type Connection struct {
  ws *websocket.Conn
  send chan []byte
  channelName string
  ipAddr string
}

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
      c.send <- createJSONs(getChats(c.channelName, "General", 50))
    case c := <-h.unregister:
      if _, ok := h.channels[c.channelName][c]; ok {
        delete(h.channels[c.channelName], c)
        close(c.send)
      }
    case m := <-h.broadcast:
      var chat = createChat(m.data, m.conn);
      if (canBroadcast(chat, m.conn)) {
        for c := range h.channels[m.conn.channelName] {
          select {
          case c.send <- createJSON(chat):
          default:
            close(c.send)
            delete(h.channels[m.conn.channelName], c)
          }
        }
        insertChat(m.conn.channelName, *chat)
      }
    }
  }
}

var upgrader = websocket.Upgrader{
  ReadBufferSize: 1024,
  WriteBufferSize: 1024,
}

func (c *Connection) readPump() {
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
    m := Message{data:d, conn:c}
    h.broadcast <- m
  }
}

func (c *Connection) write(mt int, payload []byte) error {
  c.ws.SetWriteDeadline(time.Now().Add(writeWait))
  return c.ws.WriteMessage(mt, payload)
}

func (c *Connection) writePump() {
  ticker := time.NewTicker(pingPeriod)
  defer func() {
    ticker.Stop()
    c.ws.Close()
  }()
  for {
    select {
    case Message, ok := <-c.send:
      if !ok {
        c.write(websocket.CloseMessage, []byte{})
        return
      }
      if err := c.write(websocket.TextMessage, Message); err != nil {
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
  c := &Connection{send: make(chan []byte, 256), ws: ws}
  c.channelName = req.URL.Path[4:]
  c.ipAddr = req.RemoteAddr
  h.register <- c
  go c.writePump()
  c.readPump()
}

func channelServer(w http.ResponseWriter, req *http.Request) {
  if req.Method != "GET" {
    http.Error(w, "Method not allowed", 405)
    return
  }
  w.Header().Set("Content-Type", "text/html; charset=utf-8")
  fmt.Fprintf(w, "%+v", getChannels());
}

func convoServer(w http.ResponseWriter, req *http.Request) {
  if req.Method != "GET" {
    http.Error(w, "Method not allowed", 405)
    return
  }
  w.Header().Set("Content-Type", "text/html; charset=utf-8")
  fmt.Fprintf(w, "%+v %s", getConvos(req.URL.Path[8:]), req.URL.Path[8:]);
}

func htmlServer(w http.ResponseWriter, req *http.Request) {
  if strings.Contains(req.URL.Path[1:], "/") {
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
  initDB()
  go h.run()
  http.HandleFunc("/channels", channelServer)
  http.HandleFunc("/convos/", convoServer)
  http.HandleFunc("/", htmlServer)
  http.HandleFunc("/ws/", wsServer)
  http.HandleFunc("/static/", staticServer)
  err := http.ListenAndServe(":8080", nil)
  if err != nil {
    log.Fatal("ListenAndServ: ", err)
  }
}

