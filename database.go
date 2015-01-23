package main

import (
  "database/sql"
  _ "github.com/mattn/go-sqlite3"
  "fmt"
  "time"
)

type Database struct {
  db *sql.DB
}

var storage *Database

func (s *Database) insertChannel(channelName string) {
  tx, err := s.db.Begin()
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

func (s *Database) insertConvo(channelId int, convoName string) {
  tx, err := s.db.Begin()
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

func (s *Database) insertChat(channelName string, chat Chat) {
  /* Get the ids. */
  channelId := s.getChatChannelId(channelName)
  if channelId == 0 {
    s.insertChannel(channelName)
    channelId = s.getChatChannelId(channelName)
    if channelId == 0 {
      fmt.Println("Error creating channel.");
      return
    }
  }
  convoId := s.getChatConvoId(channelId, chat.Convo)
  if convoId == 0 {
    s.insertConvo(channelId, chat.Convo)
    convoId = s.getChatConvoId(channelId, chat.Convo)
  }

  tx, err := s.db.Begin()
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

func (s *Database) getChatConvoId(channelId int, convoName string)int {
  stmt, err := s.db.Prepare("select id from Convos where name = ? and channel = ?")
  if err != nil {
    fmt.Println("Error: could not access DB.", err);
    return 0
  }
  defer stmt.Close()
  var id int
  err = stmt.QueryRow(convoName, channelId).Scan(&id)
  if err != nil {
    return 0
  }
  return id
}

func (s *Database) getChatChannelId(channelName string)int {
  stmt, err := s.db.Prepare("select id from Channels where name = ?")
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

func (s *Database) getCount(channelName string) uint64{
  stmt, err := s.db.Prepare(`
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

func (s *Database) getChannels() []string{
  var outputChannels []string
  stmt, err := s.db.Prepare(`
  select channels.name, MAX(chats.date)
  from channels
    left join chats on chats.channel = channels.id
  group by channels.name
  order by chats.date desc 
  limit 20`)
  if err != nil {
    fmt.Println("Couldn't get channels.", err)
    return outputChannels
  }
  defer stmt.Close()
  rows, err := stmt.Query()
  if err != nil {
    fmt.Println("Couldn't get channels.", err)
    return outputChannels
  }
  defer rows.Close()
  for rows.Next() {
    var channelDate int
    var channelName string
    rows.Scan(&channelName, &channelDate)
    outputChannels = append(outputChannels, channelName)
  }
  return outputChannels
}

func (s *Database) getConvos(channelName string) []string{
  var outputConvos []string
  stmt, err := s.db.Prepare(`
  select convos.name, MAX(chats.date)
  from convos
    left join chats on chats.convo = convos.id
  where convos.channel = (
    select id from channels where name = ?
  )
  group by convos.name
  order by chats.date desc 
  limit 20`)
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

func (s *Database) getChats(channelName string, convoName string, numChats uint64) []Chat {
  var outputChats []Chat
  if len(convoName) > 0 {
    stmt, err := s.db.Prepare(`
    select * from 
    (select ip, chats.name, trip, country, message, count, date,
        file_path, file_name, file_preview, file_size,
        file_dimensions
    from chats
    join (select * from channels where channels.name = ?)
      as filtered_channels on chats.channel=filtered_channels.id
    join (select * from convos where convos.name = ?)
      as filtered_convos on chats.convo=filtered_convos.id
    order by count desc limit ?) order by count asc`)
    if err != nil {
      fmt.Println("Couldn't get chats.", err)
      return outputChats
    }
    defer stmt.Close()
    rows, err := stmt.Query(channelName, convoName, numChats)
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

func initDB() *sql.DB{
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
    count INTEGER,
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
      REFERENCES Channels(id) ON DELETE CASCADE,
    UNIQUE(count, channel) ON CONFLICT REPLACE
  )`
  _, err = db.Exec(createChats)
  if err != nil {
    fmt.Println("Unable to create Chats.", err);
  }

  return db
}

