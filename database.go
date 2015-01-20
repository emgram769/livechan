package main

import (
  "database/sql"
  _ "github.com/mattn/go-sqlite3"
  "fmt"
  "time"
)

var livechanDB *sql.DB

func insertChannel(channelName string) {
  tx, err := livechanDB.Begin()
  if err != nil {
    fmt.Println("Error: could not access DB.", err);
    return
  }
  stmt, err := tx.Prepare("insert into Channels(name) values(?)")
  defer stmt.Close()
  if err != nil {
    fmt.Println("Error: could not access DB.", err);
    return
  }
  _, err = stmt.Exec(channelName)
  tx.Commit()
}

func insertConvo(channelId int, convoName string) {
  tx, err := livechanDB.Begin()
  if err != nil {
    fmt.Println("Error: could not access DB.", err);
    return
  }
  stmt, err := tx.Prepare("insert into Convos(channel, name) values(?, ?)")
  defer stmt.Close()
  if err != nil {
    fmt.Println("Error: could not access DB.", err);
    return
  }
  _, err = stmt.Exec(channelId, convoName)
  tx.Commit()
}

func insertChat(channelName string, chat Chat) {
  /* Get the ids. */
  channelId := getChatChannelId(channelName)
  if channelId == 0 {
    insertChannel(channelName)
    channelId = getChatChannelId(channelName)
  }
  convoId := getChatConvoId(chat.Convo)
  if convoId == 0 {
    insertConvo(channelId, chat.Convo)
    convoId = getChatConvoId(chat.Convo)
  }

  tx, err := livechanDB.Begin()
  if err != nil {
    fmt.Println("Error: could not access DB.", err);
    return
  }
  stmt, err := tx.Prepare(`
  insert into Chats(
    ip, name, trip, country, message, count, date, 
    file_path, file_name, file_preview, file_size, 
    file_dimensions, convo, channel
  )
  values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
  defer stmt.Close()
  if err != nil {
    fmt.Println("Error: could not access DB.", err);
    return
  }
  _, err = stmt.Exec(chat.IpAddr, chat.Name, chat.Trip, chat.Country, chat.Message, chat.Count, chat.Date.UnixNano(), chat.FilePath, chat.FileName, chat.FilePreview, chat.FileSize, chat.FileDimensions, convoId, channelId)
  tx.Commit()
  if err != nil {
    fmt.Println("Error: could not insert into DB.", err);
    return
  }
}

func getChatConvoId(convoName string)int {
  stmt, err := livechanDB.Prepare("select id from Convos where name = ?")
  if err != nil {
    fmt.Println("Error: could not access DB.", err);
    return 0
  }
  defer stmt.Close()
  var id int
  err = stmt.QueryRow(convoName).Scan(&id)
  if err != nil {
    return 0
  }
  return id
}

func getChatChannelId(channelName string)int {
  stmt, err := livechanDB.Prepare("select id from Channels where name = ?")
  if err != nil {
    fmt.Println("Error: could not access DB.", err);
    return 0
  }
  defer stmt.Close()
  var id int
  err = stmt.QueryRow(channelName).Scan(&id)
  if err != nil {
    return 0
  }
  return id
}

func getCount(channelName string) uint64{
  stmt, err := livechanDB.Prepare(`
  select MAX(count)
  from chats
  where channel = (
    select id from channels where name = ?
  )
  `)
  if err != nil {
    fmt.Println("Couldn't get count.", err)
    return 0
  }
  var count uint64
  stmt.QueryRow(channelName).Scan(&count)
  return count
}

func getConvos(channelName string) []string{
  var outputConvos []string
  stmt, err := livechanDB.Prepare(`
  select convos.name, MAX(chats.date)
  from convos
    left join chats on chats.convo = convos.id
  where convos.channel = (
    select id from channels where name = ?
  )
  group by convos.name
  order by chats.date desc 
  limit 2`)
  if err != nil {
    fmt.Println("Couldn't get convos.", err)
    return outputConvos
  }
  defer stmt.Close()
  rows, err := stmt.Query(channelName)
  if err != nil {
    fmt.Println("Couldn't get convos.", err)
    return outputConvos
  }
  defer rows.Close()
  for rows.Next() {
    var convoDate int
    var convoName string
    rows.Scan(&convoName, &convoDate)
    outputConvos = append(outputConvos, convoName)
  }
  return outputConvos
}

func getChats(channelName string, convoName string) []Chat {
  var outputChats []Chat
  if len(convoName) > 0 {
    stmt, err := livechanDB.Prepare(`
    select ip, name, trip, country, message, count, date,
      file_path, file_name, file_preview, file_size,
      file_dimensions
    from chats
    where convo = (
      select id from Convos where name = ?
    ) and channel = (
      select id from Channels where name = ?
    )
    order by count limit 10`)
    if err != nil {
      fmt.Println("Couldn't get chats.", err)
      return outputChats
    }
    defer stmt.Close()
    rows, err := stmt.Query(convoName, channelName)
    if err != nil {
      fmt.Println("Couldn't get chats.", err)
      return outputChats
    }
    defer rows.Close()
    for rows.Next() {
      var chat Chat
      var unixTime int64
      rows.Scan(&chat.IpAddr, &chat.Name, &chat.Trip, &chat.Country,
        &chat.Message, &chat.Count, &unixTime, &chat.FilePath,
        &chat.FileName, &chat.FilePreview, &chat.FileSize, &chat.FileDimensions)
      chat.Date = time.Unix(0, unixTime)
      chat.Convo = convoName
      outputChats = append(outputChats, chat)
    }
  }
  return outputChats
}

func initDB() {
  db, err := sql.Open("sqlite3", "./livechan.db");
  if err != nil {
    fmt.Println("Unable to open db.", err);
  }

  /* Create the tables. */
  createChannels := `CREATE TABLE IF NOT EXISTS Channels(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(255)
  )`
  _, err = db.Exec(createChannels)
  if err != nil {
    fmt.Println("Unable to create Channels.", err);
  }

  createConvos := `CREATE TABLE IF NOT EXISTS Convos(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(255),
    channel INTEGER,
    FOREIGN KEY(channel)
      REFERENCES Channels(id) ON DELETE CASCADE
  )`
  _, err = db.Exec(createConvos)
  if err != nil {
    fmt.Println("Unable to create Convos.", err);
  }

  createChats := `CREATE TABLE IF NOT EXISTS Chats(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ip VARCHAR(255),
    name TEXT,
    trip VARCHAR(255),
    country VARCHAR(255),
    message TEXT,
    count INTEGER UNIQUE ON CONFLICT REPLACE,
    date INTEGER,
    file_path TEXT,
    file_name TEXT,
    file_preview TEXT,
    file_size INTEGER,
    file_dimensions TEXT,
    convo INTEGER,
    channel INTEGER,
    FOREIGN KEY(convo)
      REFERENCES Convos(id) ON DELETE CASCADE,
    FOREIGN KEY(channel)
      REFERENCES Channels(id) ON DELETE CASCADE
  )`
  _, err = db.Exec(createChats)
  if err != nil {
    fmt.Println("Unable to create Chats.", err);
  }

  livechanDB = db

  /*var chat Chat
  chat.Count = 3
  chat.Convo = "General"
  chat.Date = time.Unix(100,0)
  chat.Message = "hiy234a"
  chat.IpAddr = "324234"
  insertChat("test", chat)
  chat.Count = 4
  chat.Convo = "General2"
  chat.Date = time.Unix(500,0)
  chat.IpAddr = "93242"
  chat.Message = "[pffffp["
  insertChat("test", chat)
  chat.Count = 5
  chat.Message = "florida"
  chat.Convo = "General3"
  chat.Date = time.Unix(300,0)
  insertChat("test", chat)
  chat.Count = 6
  chat.Convo = "General4"
  chat.Date = time.Unix(400,0)
  insertChat("test", chat)
  fmt.Printf("%+v\n", getConvos("test"))*/
}

