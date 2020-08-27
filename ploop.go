package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"os"
	"path/filepath"
    "strings"
    "strconv"
	"time"
	"gocv.io/x/gocv"
)

func isImage(path string) bool {
	extension := strings.ToLower(filepath.Ext(path))
	switch extension {
	case
		".jpg",
		".jpeg",
		".png",
		".tif",
		".tiff":
		return true
	}
	return false
}

func listImages(directory string) []string {
	files := make([]string, 0)
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure acessing path %q: %v\n", path, err)
			return err
		}
		if isImage(path) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		log.Fatalf("walk error %q: %v\n", directory, err)
	}

	return files
}

func annotateImage(im *gocv.Mat, text string) {
	org       := image.Pt(50, 50)
    green     := color.RGBA{0, 255, 0, 0}
    scale     := 1.5
    thickness := 2

	gocv.PutText(im, text, org, gocv.FontHersheyComplex, scale, green, thickness)
}

func main() {
	if len(os.Args) != 5 {
		log.Fatal("Usage: ploop [image directory] [output file] [start time] [interval]")
	}
    fmt.Println("Welcome to Ploop timelapse creator.")

	inputDirectory := os.Args[1]
	outputFile     := os.Args[2]
    startTime, err  := time.Parse(time.RFC3339, os.Args[3])
    if err != nil {
        fmt.Println("Could not parse date")
    }
    workingTime    := startTime
    interval, _    := strconv.ParseInt(os.Args[4], 10, 32)
    duration       := time.Duration(interval)
    files          := listImages(inputDirectory)

	window := gocv.NewWindow("current image")
	defer window.Close()

	im := gocv.IMRead(files[0], gocv.IMReadAnyColor)
	imWidth := im.Cols()
	imHeight := im.Rows()
	im.Close()

	writer, err := gocv.VideoWriterFile(outputFile, "mp4v", 30, imWidth, imHeight, true)
	if err != nil {
		fmt.Printf("error opening video writer device: %v", outputFile)
	}
	defer writer.Close()

	fmt.Printf("writing %d images to %q", len(files), outputFile)
	for _, path := range files {
        im := gocv.IMRead(path, gocv.IMReadAnyColor)
        if im.Cols() == 0 || im.Rows() == 0 {
            continue
        }

        annotateImage(&im, workingTime.String())
		window.IMShow(im)
		if window.WaitKey(1) >= 0 {
			break
        }
        workingTime = workingTime.Add(time.Second * duration)
        writer.Write(im)
		im.Close()
	}
}
