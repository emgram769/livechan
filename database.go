package main

var db = make(map[string][]Chat)

func getChats(channelName string) []Chat{
  return db[channelName]
}

func insertChat(channelName string, chat Chat) {
  db[channelName] = append(db[channelName], chat)
  var channelLen = len(db[channelName])
  if (channelLen > 50) {
    db[channelName] = db[channelName][channelLen - 50:channelLen]
  }
}

