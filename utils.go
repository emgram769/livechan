package main

import (
  "strings"
)

func Ipv4Same(ipaddr1 string, ipaddr2 string) bool {
  delim := ":"
  return strings.Split(ipaddr1, delim)[0] == strings.Split(ipaddr2, delim)[0]
}

func ExtractIpv4(addr string) string {
  return strings.Split(addr, ":")[0]
}
