package main

import (
	"strings"
	"time"
	"fmt"
	"io/ioutil"
	"encoding/base64"
)

func genUploadFilename(filename string) string {
  // FIXME invalid filenames without extension
  // get time
	fmt.Println(filename)
  timeNow := time.Now()
  // get extension
  idx := strings.LastIndex(filename, ".")
  // concat time and file extension
  fileExt := filename[idx+1:]
  return fmt.Sprintf("%d.%s", timeNow.UnixNano(), fileExt)
}


func handleUpload(filedata string, filename string) string {

	fname := genUploadFilename(filename)
	osfname := fmt.Sprintf("upload/%s", fname)
	data, err := base64.StdEncoding.DecodeString(filedata)
	if err != nil {
		fmt.Println("error converting base64 upload", err)
	}
	err = ioutil.WriteFile(osfname, data, 0644)
	if err != nil {
		fmt.Println("failed to save upload");
		return ""
	}

	return fname
}
