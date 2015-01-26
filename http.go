package main

import (
  "github.com/dchest/captcha"
  "github.com/gorilla/websocket"
  "net/http"
  "fmt"
  "strings"
)

var upgrader = websocket.Upgrader{
  ReadBufferSize: 1024,
  WriteBufferSize: 1024,
}

func wsServer(w http.ResponseWriter, req *http.Request) {
  channelName := req.URL.Path[4:] // Slice off "/ws/"
  if req.Method != "GET" {
    http.Error(w, "Method not allowed", 405)
    return
  }
  if (storage.getChatChannelId(channelName) == 0) {
    http.Error(w, "Method not allowed", 405)
    return
  }
  ws, err := upgrader.Upgrade(w, req, nil)
  if err != nil {
    fmt.Println(err)
    return
  }
  c := &Connection{
    send: make(chan []byte, 256),
    ws: ws,
    channelName: channelName,
    ipAddr: req.RemoteAddr,
    }
  h.register <- c

  /* Start a reader/writer pair for the new connection. */
  go c.writer()
  /* Nature of go treats this handler as a goroutine.
     Small optimization to not spawn a new one. */
  c.reader()
}

func channelServer(w http.ResponseWriter, req *http.Request) {
  if req.Method != "GET" {
    http.Error(w, "Method not allowed", 405)
    return
  }
  w.Header().Set("Content-Type", "text/html; charset=utf-8")
  fmt.Fprintf(w, "%+v", storage.getChannels());
}

func convoServer(w http.ResponseWriter, req *http.Request) {
  if req.Method != "GET" {
    http.Error(w, "Method not allowed", 405)
    return
  }
  w.Header().Set("Content-Type", "text/html; charset=utf-8")
  fmt.Fprintf(w, "%+v %s", storage.getConvos(req.URL.Path[8:]), req.URL.Path[8:]);
}

func htmlServer(w http.ResponseWriter, req *http.Request) {
  channelName := req.URL.Path[1:] // Omit the leading "/"

  /* Disallow / in the name. */
  if strings.Contains(channelName, "/") {
    http.Error(w, "Method not allowed", 405)
    return
  }

  if req.Method != "GET" {
    http.Error(w, "Method not allowed", 405)
    return
  }

  if channelName == "" {
    channelName = "General"
  }

  if (storage.getChatChannelId(channelName) == 0) {
    http.Error(w, "No registration pages, yet!", 405)
    return
  }

  w.Header().Set("Content-Type", "text/html; charset=utf-8")
  http.ServeFile(w, req, "index.html")
}

func captchaServer(w http.ResponseWriter, req *http.Request) {
  if req.Method == "GET" {
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    fmt.Fprintf(w, "{captcha: %s}", captcha.New());
    return
  } else if req.Method == "POST" {
    if captcha.VerifyString(req.FormValue("captchaId"), req.FormValue("captchaSolution")) {
    } else {
    }
  }
}

func staticServer(w http.ResponseWriter, req *http.Request) {
    http.ServeFile(w, req, req.URL.Path[1:])
}

