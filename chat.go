package main

import (
  "encoding/json"
  "fmt"
  "time"
  "strings"
)

var count uint64 = 0

/* To be stored in the DB. */
type Chat struct {
  Message string
  Name string
  Date time.Time
  Count uint64
  IpAddr string
}

/* To be visible to users. */
type OutChat struct {
  Message string
  Name string
  Date time.Time
  Count uint64
  Trip string
}

func createChat(data []byte, conn *Connection) *Chat{
  c := new(Chat)
  err:=json.Unmarshal(data, c)
  if err != nil {
    fmt.Println("error: ", err)
  }

  c.Name = strings.TrimSpace(c.Name)
  if len(c.Name) == 0 {
    c.Name = "Anonymous"
  }

  c.Message = strings.TrimSpace(c.Message)

  c.Date = time.Now()
  c.IpAddr = conn.ipAddr
  return c
}

func createJSON(chat *Chat) []byte{
  outChat := OutChat{
    Name: chat.Name,
    Message: chat.Message,
    Date: chat.Date,
    Count: chat.Count,
  }
  j, err := json.Marshal(outChat)
  if err != nil {
    fmt.Println("error: ", err)
  }
  return j
}

func createJSONs(chats []Chat) []byte{
  var outChats []OutChat
  for _, chat := range chats {
    outChat := OutChat{
      Name: chat.Name,
      Message: chat.Message,
      Date: chat.Date,
      Count: chat.Count,
    }
    outChats = append(outChats, outChat)
  }
  j, err := json.Marshal(outChats)
  if err != nil {
    fmt.Println("error: ", err)
  }
  return j
}

func canBroadcast(chat *Chat, conn *Connection) bool{
  if len(chat.Message) == 0 {
    return false
  }
  var t = h.channels[conn.channelName][conn]
  if time.Now().Sub(t).Seconds() < 4 {
    return false
  }
  h.channels[conn.channelName][conn] = time.Now()
  count = count + 1
  chat.Count = count
  return true
}

