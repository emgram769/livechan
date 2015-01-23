package main

import (
  "net/http"
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
  err := http.ListenAndServe(":8080", nil)
  if err != nil {
    fmt.Println("Unable to serve: ", err)
  }
}

