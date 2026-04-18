package main

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/png"
	"os"

	"golang.org/x/image/draw"
)

func main() {
	if len(os.Args) < 3 {
		println("Usage: png2ico <input.png> <output.ico>")
		os.Exit(1)
	}

	inputPath := os.Args[1]
	outputPath := os.Args[2]

	inFile, err := os.Open(inputPath)
	if err != nil {
		println("Error opening input:", err.Error())
		os.Exit(1)
	}
	defer inFile.Close()

	img, err := png.Decode(inFile)
	if err != nil {
		println("Error decoding PNG:", err.Error())
		os.Exit(1)
	}

	sizes := []int{256, 128, 64, 48, 32, 16}
	var icons [][]byte

	for _, size := range sizes {
		resized := resizeImage(img, size)
		data := encodePNG(resized)
		icons = append(icons, data)
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		println("Error creating output:", err.Error())
		os.Exit(1)
	}
	defer outFile.Close()

	if err := writeICO(outFile, sizes, icons); err != nil {
		println("Error writing ICO:", err.Error())
		os.Exit(1)
	}

	println("Created:", outputPath)
}

func resizeImage(src image.Image, size int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, size, size))
	draw.BiLinear.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)
	return dst
}

func encodePNG(img image.Image) []byte {
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}

func writeICO(w *os.File, sizes []int, icons [][]byte) error {
	count := uint16(len(icons))

	_ = binary.Write(w, binary.LittleEndian, uint16(0))
	_ = binary.Write(w, binary.LittleEndian, uint16(1))
	_ = binary.Write(w, binary.LittleEndian, count)

	dataOffset := uint32(6 + count*16)
	var allData []byte

	for i, size := range sizes {
		imgData := icons[i]
		imgSize := uint32(len(imgData))

		width := uint8(size)
		if size >= 256 {
			width = 0
		}
		height := width

		_, _ = w.Write([]byte{width, height, 0, 0, 1, 0, 32, 0})
		_ = binary.Write(w, binary.LittleEndian, imgSize)
		_ = binary.Write(w, binary.LittleEndian, dataOffset)

		dataOffset += imgSize
		allData = append(allData, imgData...)
	}

	_, err := w.Write(allData)
	return err
}
