package main

import (
  "net/http"
  "github.com/dchest/captcha"
  "fmt"
)

func main() {
  db  := initDB()
  storage = &Database{db:db}
  go h.run()
  http.HandleFunc("/channels", channelServer)
  http.HandleFunc("/convos/", convoServer)
  http.HandleFunc("/", htmlServer)
  http.HandleFunc("/ws/", wsServer)
  http.HandleFunc("/static/", staticServer)
  http.HandleFunc("/captcha", captchaServer)
  http.Handle("/captcha/", captcha.Server(captcha.StdWidth, captcha.StdHeight))
  err := http.ListenAndServe(":18080", nil)
  if err != nil {
    fmt.Println("Unable to serve: ", err)
  }
}

