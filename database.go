package main

var db = make(map[string][]Chat)

func getChats(channelName string) []Chat{
  return db[channelName]
}

func insertChat(channelName string, chat Chat) {
  db[channelName] = append(db[channelName], chat)
}

