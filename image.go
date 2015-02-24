package main

import (
  "log"
  "github.com/gographics/imagick/imagick"
)

// generate thumbnail for image
func generateThumbnail(inFname, outFname string, data []byte) error {
  log.Println("generating thumbnail for", inFname)

  var err error

  wand := imagick.NewMagickWand()
  defer wand.Destroy()
  
  err = wand.ReadImageBlob(data)
  if err != nil {
    return err
  }
  w := wand.GetImageWidth()
  h := wand.GetImageHeight()
  
  var thumb_w, thumb_h, scale float64
  
  scale = 180
  modifer := scale / float64(w)
  
  thumb_w = modifer * float64(w)
  thumb_h = modifer * float64(h)
  
  err = wand.ScaleImage(uint(thumb_w), uint(thumb_h))
  if err != nil {
    log.Println("could not scale image to make thumbnail", err)
    return err
  }
  err = wand.WriteImage(outFname)
  return err
}