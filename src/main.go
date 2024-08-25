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

func pixelsToImage(pixels [][]float64, outFilepath string) {
    ROWS := len(pixels)
    COLS := len(pixels[0])

    upLeft := image.Point{0, 0}
    lowRight := image.Point{COLS, ROWS}
    img := image.NewRGBA(image.Rectangle{upLeft, lowRight})

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
            x := uint8((pixels[r][c] - float64(minIntensity)) / float64(maxIntensity - minIntensity) * 0xff)
            // x := uint8(pixels[r][c])
            // random uint16 to uint8 conversion out of nowhere
            // x := uint8(pixels[r][c] / 0xffff * 0xff)
            img.Set(c, r, color.RGBA{x, x, x, 0xff})
        }
    }

    // Encode as PNG.
    f, _ := os.Create(outFilepath)
    png.Encode(f, img)
}

// calculate gradient of grayscale image using sobel filter
func grayscaleImageToGradient(pixels [][]float64) [][]float64 {
    var gradient [][]float64
    ROWS := len(pixels)
    COLS := len(pixels[0])
    for r := 0; r < ROWS; r++ {
        var row []float64
        for c := 0; c < COLS; c++ {
            if r == 0 || c == 0 || r == ROWS - 1 || c == COLS - 1 {
                row = append(row, 0)
                continue
            }
            var gy float64 = 0
            var gx float64 = 0
            // i dont know how to do matmul in go
            gy += pixels[r - 1][c] * 2
            gy += pixels[r - 1][c - 1]
            gy += pixels[r - 1][c + 1]

            gy -= pixels[r + 1][c] * 2
            gy -= pixels[r + 1][c - 1]
            gy -= pixels[r + 1][c + 1]

            gx += pixels[r][c - 1] * 2
            gx += pixels[r - 1][c - 1]
            gx += pixels[r + 1][c - 1]


            gx -= pixels[r][c + 1] * 2
            gx -= pixels[r - 1][c + 1]
            gx -= pixels[r + 1][c + 1]

            row = append(row, math.Sqrt(gx * gx + gy * gy))
        }
        gradient = append(gradient, row)
    }
    return gradient
}

func rgbaToGrayscale(r, g, b uint8) float64 {
    // some actual nerd told me to do this and was lying...
    // log.Printf("%v, %v, %v, %v\n", r, g, b, a)
    // R := math.Pow(float64(r)/0xff, 1.0/1.0)
    // G := math.Pow(float64(g)/0xff, 1.0/1.0)
    // B := math.Pow(float64(b)/0xff, 1.0/1.0)
    // y := .2126 * R + .7152 * G + .0722 * B
    // l := 116 * math.Pow(y, 1.0/3.0) - 16
    l := 0.299 * float64(r) + 0.587 * float64(g) + 0.114 * float64(b)
    return l
}

func imageToGrayscale(imageData image.Image) [][]float64 {
    bounds := imageData.Bounds()
    width, height := bounds.Max.X, bounds.Max.Y

    var pixels [][]float64
    for r := 0; r < height; r++ {
        var row []float64
        for c := 0; c < width; c++ {
            R, G, B, _ := imageData.At(c, r).RGBA()
            R /= 0xff
            G /= 0xff
            B /= 0xff
            row = append(row, rgbaToGrayscale(uint8(R), uint8(G), uint8(B)))
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
    pixelsToImage(grayscaleImage, "grayscale.png")
    gradient := grayscaleImageToGradient(grayscaleImage)
    pixelsToImage(gradient, "gradient.png")
}
