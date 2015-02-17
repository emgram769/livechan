package main

import (
  "database/sql"
  _ "github.com/mattn/go-sqlite3"
  "time"
  "log"
)

type Database struct {
  db *sql.DB
}

var storage *Database

func (s *Database) deleteChatForIP(ipaddr string) {
  log.Println("delete all chat for ", ipaddr)
  tx, err := s.db.Begin()
  if err != nil {
    log.Println("Error: could not access DB.", err)
    return
  }
  stmt, err := s.db.Prepare("SELECT file_path FROM Chats WHERE ip = ?")
  rows, err := stmt.Query(&ipaddr)
  defer rows.Close()
  for rows.Next() {
    var chat Chat;
    rows.Scan(&chat.FilePath)
    chat.DeleteFile();
  }
  defer stmt.Close()
  stmt, err = tx.Prepare("DELETE FROM Chats WHERE ip = ?")
  defer stmt.Close()
  if err != nil {
    log.Println("Error: could not access DB.", err)
    return
  }
  _, err = stmt.Query(ipaddr)
  tx.Commit()
}


func (s *Database) insertChannel(channelName string) {
  tx, err := s.db.Begin()
  if err != nil {
    log.Println("Error: could not access DB.", err)
    return
  }
  stmt, err := tx.Prepare("INSERT INTO Channels(name) VALUES(?)")
  defer stmt.Close()
  if err != nil {
    log.Println("Error: could not access DB.", err);
    return
  }
  _, err = stmt.Exec(channelName)
  tx.Commit()
}

func (s *Database) insertConvo(channelId int, convoName string) {
  tx, err := s.db.Begin()
  if err != nil {
    log.Println("Error: could not access DB.", err)
    return
  }
  stmt, err := tx.Prepare("INSERT INTO Convos(channel, name) VALUES(?, ?)")
  defer stmt.Close()
  if err != nil {
    log.Println("Error: could not access DB.", err);
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
      log.Println("Error creating channel.", channelName);
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
    log.Println("Error: could not access DB.", err);
    return
  }
  stmt, err := tx.Prepare(`
  INSERT INTO Chats(
    ip, name, trip, country, message, count, chat_date, 
    file_path, file_name, file_preview, file_size, 
    file_dimensions, convo, channel
  )
  VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
  defer stmt.Close()
  if err != nil {
    log.Println("Error: could not access DB.", err);
    return
  }
  _, err = stmt.Exec(chat.IpAddr, chat.Name, chat.Trip, chat.Country, chat.Message, chat.Count, chat.Date.UnixNano(), chat.FilePath, chat.FileName, chat.FilePreview, chat.FileSize, chat.FileDimensions, convoId, channelId)
  tx.Commit()
  if err != nil {
    log.Println("Error: could not insert into DB.", err);
    return
  }
}

func (s *Database) getChatConvoId(channelId int, convoName string)int {
  stmt, err := s.db.Prepare("SELECT id FROM Convos WHERE name = ? AND channel = ?")
  if err != nil {
    log.Println("Error: could not access DB.", err);
    return 0
  }
  defer stmt.Close()
  var id int
  err = stmt.QueryRow(convoName, channelId).Scan(&id)
  if err != nil {
    log.Println("error getting convo id", err)
    return 0
  }
  return id
}

func (s *Database) getChatChannelId(channelName string)int {
  stmt, err := s.db.Prepare("SELECT id FROM Channels WHERE name = ?")
  if err != nil {
    log.Println("Error: could not access DB.", err);
    return 0
  }
  defer stmt.Close()
  var id int
  err = stmt.QueryRow(channelName).Scan(&id)
  if err != nil {
    log.Println("error getting channel id", err)
    return 0
  }
  return id
}

func (s *Database) getCount(channelName string) uint64{
  stmt, err := s.db.Prepare(`
  SELECT MAX(count)
  FROM chats
  WHERE channel = (
    SELECT id FROM channels WHERE name = ?
  )
  `)
  if err != nil {
    log.Println("Couldn't get count.", err)
    return 0
  }
  var count uint64
  stmt.QueryRow(channelName).Scan(&count)
  return count
}

func (s *Database) getChannels() []string{
  var outputChannels []string
  stmt, err := s.db.Prepare(`
  SELECT channels.name, MAX(chats.date)
  FROM channels
    left join chats ON chats.channel = channels.id
  GROUP BY channels.name
  ORDER BY chats.date LIMIT DESC 20`)
  if err != nil {
    log.Println("Couldn't get channels.", err)
    return outputChannels
  }
  defer stmt.Close()
  rows, err := stmt.Query()
  if err != nil {
    log.Println("Couldn't get channels.", err)
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
  SELECT convos.name, MAX(chats.date)
  FROM convos
    left join chats ON chats.convo = convos.id
  WHERE convos.channel = (
    SELECT id FROM channels WHERE name = ?
  )
  GROUP BY convos.name
  ORDER BY chats.date DESC LIMIT 20`)
  if err != nil {
    log.Println("Couldn't get convos.", err)
    return outputConvos
  }
  defer stmt.Close()
  rows, err := stmt.Query(channelName)
  if err != nil {
    log.Println("Couldn't get convos.", err)
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
    SELECT * FROM
    (SELECT ip, chats.name, trip, country, message, count, chat_date,
        file_path, file_name, file_preview, file_size,
        file_dimensions
    FROM chats
    JOIN (SELECT * FROM channels WHERE channels.name = ?)
      AS filtered_channels ON chats.channel=filtered_channels.id
    JOIN (SELECT * FROM convos WHERE convos.name = ?)
      AS filtered_convos ON chats.convo=filtered_convos.id
    ORDER BY COUNT DESC LIMIT ?) ORDER BY COUNT ASC`)
    if err != nil {
      log.Println("Couldn't get chats.", err)
      return outputChats
    }
    defer stmt.Close()
    rows, err := stmt.Query(channelName, convoName, numChats)
    if err != nil {
      log.Println("Couldn't get chats.", err)
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

func (s *Database) getPermissions(channelName string, userName string) uint64 {
  stmt, err := s.db.Prepare(`
  SELECT permissions FROM Owners
  WHERE user = (SELECT id FROM users WHERE name = ?)
  AND channel = (SELECT id FROM channels WHERE name = ?)
  `)
  if err != nil {
    log.Println("Error: could not access DB.", err);
    return 0
  }
  defer stmt.Close()
  var permissions uint64
  err = stmt.QueryRow(userName, channelName).Scan(&permissions)
  if err != nil {
    return 0
  }
  return 0
}

func (s *Database) isBanned(channelName string, ipAddr string) bool {
  stmt, err := s.db.Prepare(`
  SELECT COUNT(*) FROM Bans
  WHERE ip = ? AND channel = (SELECT id FROM channels WHERE name = ?)
  `)
  if err != nil {
    log.Println("Error: could not access DB.", err);
    return false
  }
  defer stmt.Close()
  var isbanned int
  err = stmt.QueryRow(ipAddr, channelName).Scan(&isbanned)
  if err != nil {
    log.Println("failed to query database for ban", err)
    return false
  }
  if (isbanned > 0) {
    return true
  } else {
    return false;
  }
}

func (s *Database) getBan(channelName string, ipAddr string) Ban {
  var ban Ban
  stmt, err := s.db.Prepare(`
  SELECT offense, date, expiration, ip FROM Bans
  WHERE ip = ? AND channel = (SELECT id FROM channels WHERE name = ?)
  `)
  if err != nil {
    log.Println("Error: could not access DB.", err)
    return ban
  }
  defer stmt.Close()
  var unixTime int64
  var unixTimeExpiration int64
  err = stmt.QueryRow(ipAddr, channelName).Scan(&ban.Offense, &unixTime, &unixTimeExpiration, &ban.IpAddr)
  ban.Date = time.Unix(0, unixTime)
  ban.Expiration = time.Unix(0, unixTimeExpiration)
  if err != nil {
    log.Println("error getting ban", err)
    return ban
  }
  return ban
}

func initDB() *sql.DB{
  log.Println("initialize database")
  db, err := sql.Open("sqlite3", "./livechan.db");
  if err != nil {
    log.Println("Unable to open db.", err);
  }

  /* Create the tables. */
  createChannels := `CREATE TABLE IF NOT EXISTS Channels(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(255),
    api_key VARCHAR(255),
    options TEXT,
    restricted INTEGER,
    generated INTEGER
  )`
  _, err = db.Exec(createChannels)
  if err != nil {
    log.Println("Unable to create Channels.", err);
  }

  createConvos := `CREATE TABLE IF NOT EXISTS Convos(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(255),
    channel INTEGER,
    creator VARCHAR(255),
    date INTEGER,
    FOREIGN KEY(channel)
      REFERENCES Channels(id) ON DELETE CASCADE
  )`
  _, err = db.Exec(createConvos)
  if err != nil {
    log.Println("Unable to create Convos.", err);
  }

  createChats := `CREATE TABLE IF NOT EXISTS Chats(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ip VARCHAR(255),
    name VARCHAR(255),
    trip VARCHAR(255),
    country VARCHAR(255),
    message TEXT,
    count INTEGER,
    chat_date INTEGER,
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
    log.Println("Unable to create Chats.", err);
  }

  createUsers := `CREATE TABLE IF NOT EXISTS Users(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(255),
    password VARCHAR(255),
    salt VARCHAR(255),
    session VARCHAR(255),
    date INTEGER,
    identifiers TEXT
  )`
  _, err = db.Exec(createUsers)
  if err != nil {
    log.Println("Unable to create Users.", err);
  }

  createBans := `CREATE TABLE IF NOT EXISTS Bans(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ip VARCHAR(255),
    offense TEXT,
    date INTEGER,
    expiration INTEGER,
    banner INTEGER,
    FOREIGN KEY(banner)
      REFERENCES Users(id) ON DELETE CASCADE
  )`
  _, err = db.Exec(createBans)
  if err != nil {
    log.Println("Unable to create Bans.", err);
  }

  createOwners := `CREATE TABLE IF NOT EXISTS Owners(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user INTEGER,
    channel INTEGER,
    permissions INTEGER,
    FOREIGN KEY(user)
      REFERENCES Users(id) ON DELETE CASCADE,
    FOREIGN KEY(channel)
      REFERENCES Channels(id) ON DELETE CASCADE
  )`
  _, err = db.Exec(createOwners)
  if err != nil {
    log.Println("Unable to create Owners.", err);
  }

  return db
}

