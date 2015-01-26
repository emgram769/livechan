package main

import (
  "time"
)

const (
  restr_none = 0       // No restriction
  restr_rate = 1       // Rate limited (4 seconds)
  restr_captcha = 2    // Captcha sessions (24 hours)
  restr_nofiles = 4    // No uploads
  restr_registered = 8 // Users must be registered
  gen_country = 1
  gen_id = 2
)

type Channel struct {
  Name string
  Restrictions uint64
  Generated uint64
  Options string // JSON
}

type Owner struct {
  User string
  Channel string
  Permissions uint64
}

type Ban struct {
  Offense string
  Date time.Time
  Expiration time.Time
  IpAddr string
}

