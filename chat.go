package main

import (
  "encoding/json"
  "fmt"
  "time"
  "strings"
  "os"
  "log"
)

type InChat struct {
  Convo string
  Name string
  Message string
  File string
  FileName string
}

/* To be stored in the DB. */
type Chat struct {
  IpAddr string
  Name string
  Trip string
  Country string
  Message string
  Count uint64
  Date time.Time
  FilePath string
  FileName string
  FilePreview string
  FileSize string
  FileDimensions string
  Convo string
  UserID string
}

/* To be visible to users. */
type OutChat struct {
  UserCount uint64
  Name string
  Trip string
  Country string
  Message string
  Count uint64
  Date time.Time
  FilePath string
  FileName string
  FilePreview string
  FileSize string
  FileDimensions string
  Convo string
  Capcode string //for stuff like (you) and (mod)
}

func createChat(data []byte, conn *Connection) *Chat {
  c := new(Chat)
  inchat := new(InChat)
  err:=json.Unmarshal(data, inchat)
  if err != nil {
    log.Println(conn.ipAddr, "error creating chat: ", err)
  }
  if len(inchat.File) > 0 && len(inchat.FileName) > 0 {
    // TODO FilePreview, FileDimensions
    c.FilePath = genUploadFilename(inchat.FileName)
    c.FileName = inchat.FileName
    log.Println(conn.ipAddr, "uploaded file", c.FilePath)
    handleUpload(inchat, c.FilePath);
  }
  c.Name = strings.TrimSpace(inchat.Name)
  if len(c.Name) == 0 {
    c.Name = "Anonymous"
  }

  c.Convo = strings.TrimSpace(inchat.Convo)
  if len(c.Convo) == 0 {
    c.Convo = "General"
  }
  
  c.Message = strings.TrimSpace(inchat.Message)
  c.Date = time.Now().UTC()
  c.IpAddr = ExtractIpv4(conn.ipAddr);
  return c
}

func (chat *Chat) DeleteFile() {
  os.Remove(fmt.Sprintf("upload/%s",chat.FilePath));
}

func (chat *Chat) genCapcode(conn *Connection) string {
  cap := ""
  if ExtractIpv4(conn.ipAddr) == chat.IpAddr {
    cap = "(You)"
  }
  return cap
}

func (chat *OutChat) createJSON() []byte {
  j, err := json.Marshal(chat)
  if err != nil {
    log.Println("error: ", err)
  }
  return j
}

func (chat *Chat) createJSON(conn *Connection) []byte{
  outChat := OutChat{
    Name: chat.Name,
    Message: chat.Message,
    Date: chat.Date,
    Count: chat.Count,
    Convo: chat.Convo,
    FilePath: chat.FilePath,
    Capcode: chat.genCapcode(conn),
  }
  return outChat.createJSON()
}

func createJSONs(chats []Chat, conn * Connection) []byte{
  var outChats []OutChat
  for _, chat := range chats {
    outChat := OutChat{
      Name: chat.Name,
      Message: chat.Message,
      Date: chat.Date,
      Count: chat.Count,
      Convo: chat.Convo,
      FilePath: chat.FilePath,
      Capcode: chat.genCapcode(conn),
    }
    outChats = append(outChats, outChat)
  }
  j, err := json.Marshal(outChats)
  if err != nil {
    log.Println("error marshalling json: ", err)
  }
  return j
}

func (chat *Chat) canBroadcast(conn *Connection) bool{
  if len(chat.Message) == 0 {
    return false
  }
  var t = h.channels[conn.channelName][conn]
  // limit minimum broadcast time to 4 seconds
  if time.Now().Sub(t).Seconds() < 4 {
    return false
  }
  h.channels[conn.channelName][conn] = time.Now()
  chat.Count = storage.getCount(conn.channelName) + 1
  return true
}
