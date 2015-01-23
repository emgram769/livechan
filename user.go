package main

import (
  "time"
)

/* Registered users can moderate, own channels, etc. */
type RegisteredUser struct {
  Name string
  Password string // this is hashed of course
  Created time.Time
}

/* Anonymous users kept track of for spamming/banning. */
type User struct {
  IpAddr string
  LastPost time.Time
}

