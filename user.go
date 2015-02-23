package main

import (
  "strconv"
  "time"
  "fmt"
)

/* Registered users can moderate, own channels, etc. */
type User struct {
  Name string
  Password string
  Created time.Time
  //Identifiers string // JSON
  Attributes map[string]string
  Session string
}

// generate channel property name
// conveniece
func chanPropName(chanName string, propName string) string {
  return fmt.Sprintf("%s.%s", chanName, propName)
}

// get mod priviledge level
func (user *User) GetModLevel(chanName string) int {
  key := chanPropName(chanName, "modlevel")
  attr, ok := user.Attributes[key]
  if ok {
    modlevel, err := strconv.Atoi(attr)
    if err == nil {
      return modlevel
    }
  }
  return 0
}


// is a channel janitor
func (user *User) IsChanJan(chanName string) bool {
  return user.GetModLevel(chanName) > 0
}

// is a channel Moderator
func (user *User) IsChanMod(chanName string) bool {
  return user.GetModLevel(chanName) > 1 
}

// ban a user
func (user *User) BanUser(chanName string) bool {
  // unimplemented
  return false
}