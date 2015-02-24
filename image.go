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
  
  var ar, thumb_w, thumb_h, scale float64
  
  scale = 350
  
  // aspect ratio
  ar = float64(w) / float64(h)
  
  thumb_w = ar * scale
  thumb_h = ar * scale
  
  err = wand.ScaleImage(uint(thumb_w), uint(thumb_h))
  if err != nil {
    log.Println("could not scale image to make thumbnail", err)
    return err
  }
  err = wand.WriteImage(outFname)
  return err
}