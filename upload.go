package main

import (
  "strings"
  "time"
  "fmt"
  "io/ioutil"
  "encoding/base64"
  "log"
)

func genUploadFilename(filename string) string {
  // FIXME invalid filenames without extension
  // get time
  timeNow := time.Now()
  // get extension
  idx := strings.LastIndex(filename, ".")
  // concat time and file extension
  fileExt := filename[idx+1:]
  return fmt.Sprintf("%d.%s", timeNow.UnixNano(), fileExt)
}


func handleUpload(chat *InChat, fname string) {

  osfname := fmt.Sprintf("upload/%s", fname)
  data, err := base64.StdEncoding.DecodeString(chat.File)
  if err != nil {
    log.Println("error converting base64 upload", err)
    return
  }
  err = ioutil.WriteFile(osfname, data, 0644)
  if err != nil {
    log.Println("failed to save upload");
  }
}
