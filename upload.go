package main

import (
  "strings"
  "time"
  "fmt"
  "io/ioutil"
  "encoding/base64"
  "path/filepath"
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
  fname := fmt.Sprintf("%d.%s", timeNow.UnixNano(), fileExt)
  return filepath.Clean(fname)
}

// handle file upload
func handleUpload(chat *InChat, fname string) {

  osfname := filepath.Join("upload", fname)
  thumbnail := filepath.Join("thumbs", fname)
  data, err := base64.StdEncoding.DecodeString(chat.File)
  if err != nil {
    log.Println("error converting base64 upload", err)
    return
  }
  err = generateThumbnail(fname, thumbnail, data)
  if err != nil {
    log.Println("failed to generate thumbnail", err)
    return
  }
  err = ioutil.WriteFile(osfname, data, 0644)
  if err != nil {
    log.Println("failed to save upload", err);
  }
}
