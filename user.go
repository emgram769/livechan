package main

import (
  "time"
)

/* Registered users can moderate, own channels, etc. */
type User struct {
  Name string
  Password string
  Created time.Time
  Identifiers string // JSON
  Session string
}

