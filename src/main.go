package main

import (
	// "fmt"
	"image"
	"image/jpeg"
	"image/png"
    "image/color"
	"log"
	"os"
    "math"
)

func pixelsToImage(pixels [][]float64) {
    ROWS := len(pixels)
    COLS := len(pixels[0])

    upLeft := image.Point{0, 0}
    lowRight := image.Point{COLS, ROWS}
    img := image.NewRGBA(image.Rectangle{upLeft, lowRight})

    // Colors are defined by Red, Green, Blue, Alpha uint8 values.
    // cyan := color.RGBA{100, 200, 200, 0xff}
    // find max intensity color 
    var maxIntensity uint32 = 0
    var minIntensity uint32 = 99999
    for r := 0; r < ROWS; r++ {
        for c := 0; c < COLS; c++ {
            if uint32(pixels[r][c]) > maxIntensity {
                maxIntensity = uint32(pixels[r][c])
            }
            if uint32(pixels[r][c]) < minIntensity {
                minIntensity = uint32(pixels[r][c])
            }
        }
    }

    log.Printf("[INFO] maxIntensity: %v\n", maxIntensity)
    log.Printf("[INFO] minIntensity: %v\n", minIntensity)


    // Set color for each pixel.
    for r := 0; r < ROWS; r++ {
        for c := 0; c < COLS; c++ {
            x := uint8(pixels[r][c])
            img.Set(c, r, color.RGBA{x, x, x, 0xff})
        }
    }

    // Encode as PNG.
    f, _ := os.Create("image.png")
    png.Encode(f, img)
}

func grayscaleImageToGradient(pixels [][]float64) [][]float64 {
    var gradient [][]float64
    ROWS := len(pixels)
    COLS := len(pixels[0])
    for r := 0; r < ROWS; r++ {
        var row []float64
        for c := 0; c < COLS; c++ {
            var gy float64 = 0
            var gx float64 = 0
            if r != 0 {
                gy += pixels[r - 1][c]
            } else {
                gy += pixels[r][c]
            }

            if r != ROWS - 1 {
                gy -= pixels[r + 1][c]
            } else {
                gy -= pixels[r][c]
            }

            if c != 0 {
                gx += pixels[r][c - 1]
            } else {
                gx += pixels[r][c]
            }

            if c != COLS - 1 {
                gx -= pixels[r][c + 1]
            } else {
                gx -= pixels[r][c]
            }
            row = append(row, math.Sqrt(gx * gx + gy * gy))
        }
        gradient = append(gradient, row)
    }
    return gradient
}

func rgbaToGrayscale(r, g, b, a uint32) float64 {
    // some actual nerd told me to do this and was lying...
    // Y = .2126 * R^gamma + .7152 * G^gamma + .0722 * B^gamma
    // L* = 116 * Y ^ 1/3 - 16
    // R := math.Pow(float64(r)/255, 1.0/3.0)
    // G := math.Pow(float64(g)/255, 1.0/3.0)
    // B := math.Pow(float64(b)/255, 1.0/3.0)
    // y := .2126 * R + .7152 * G + .0722 * B * float64(a/a)
    // l := 116 * math.Pow(y, 1.0/3.0) - 16
    // appease the compiler
    l := 0.299 * float64(r) + 0.587 * float64(g) + 0.114 * float64(b) + float64(a - a)
    return l
}

func imageToGrayscale(imageData image.Image) [][]float64 {
    bounds := imageData.Bounds()
    width, height := bounds.Max.X, bounds.Max.Y

    var pixels [][]float64
    for r := 0; r < height; r++ {
        var row []float64
        for c := 0; c < width; c++ {
            row = append(row, rgbaToGrayscale(imageData.At(c, r).RGBA()))
        }
        pixels = append(pixels, row)
    }
    return pixels
}

func getImageData(imageFile *os.File, filepath string) image.Image {
	var imageData image.Image
    var err error
	switch s := filepath[len(filepath)-3:]; s {
	case "png":
        // pixels are color.RGBA
		imageData, err = png.Decode(imageFile)
		if err != nil {
			log.Fatalf("[ERROR] Image (%s) could not be decoded\n", filepath)
		}
	case "jpg":
        // pixels are color.YCbCr
		imageData, err = jpeg.Decode(imageFile)
		if err != nil {
			log.Fatalf("[ERROR] Image (%s) could not be decoded\n", filepath)
		}
    default:
        log.Fatalf("[ERROR] Image suffix (%s) not supported\n", s)
	}

    return imageData
}

func main() {
	// Reading command-line args
	log.SetFlags(0)
	args := os.Args[1:]
	if len(args) != 1 {
		log.Fatalf("[ERROR] No file path was provided\n")
	}

	// Loading file from filepath
	filepath := args[0]
	imageFile, err := os.Open(filepath)
	if err != nil {
		log.Fatalf("[ERROR] File path (%s) not valid\n", filepath)
	}
	defer imageFile.Close()

    imageData := getImageData(imageFile, filepath)
    grayscaleImage := imageToGrayscale(imageData)
    gradient := grayscaleImageToGradient(grayscaleImage)
    pixelsToImage(gradient)
}
