package main

import (
	"image"
	"image/png"
	"os"

	"golang.org/x/image/draw"
)

func main() {
	f, err := os.Open("logo.png")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	src, err := png.Decode(f)
	if err != nil {
		panic(err)
	}

	dst := image.NewRGBA(image.Rect(0, 0, 256, 256))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)

	out, err := os.Create("winres/icon.png")
	if err != nil {
		panic(err)
	}
	defer out.Close()

	png.Encode(out, dst)
}
