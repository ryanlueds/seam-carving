package main

import (
	// "fmt"
	"image"
	"image/jpeg" // not really supported
	"image/png"
    "image/color"
	"log"
	"os"
    "math"
    "cmp"
)

func convertFloatToColor(pixels [][]float64) [][]color.RGBA {
    ROWS, COLS := len(pixels), len(pixels[0])
    var colorImage [][]color.RGBA

    for r := 0; r < ROWS; r++ {
        var row []color.RGBA
        for c := 0; c < COLS; c++ {
            x := uint8(pixels[r][c])
            row = append(row, color.RGBA{R:x, G:x, B:x, A:0xff})
        }
        colorImage = append(colorImage, row)
    }

    return colorImage
}
func pixelsToImage(pixels [][]color.RGBA, outFilepath string) {
    ROWS := len(pixels)
    COLS := len(pixels[0])

    upLeft := image.Point{0, 0}
    lowRight := image.Point{COLS, ROWS}
    img := image.NewRGBA(image.Rectangle{upLeft, lowRight})

    // Set color for each pixel.
    for r := 0; r < ROWS; r++ {
        for c := 0; c < COLS; c++ {
            img.Set(c, r, pixels[r][c])
        }
    }

    // Encode as PNG.
    f, _ := os.Create(outFilepath)
    png.Encode(f, img)
}

// calculate gradient of grayscale image using sobel filter
func grayscaleImageToGradient(pixels [][]color.RGBA) [][]color.RGBA {
    var gradient [][]color.RGBA
    ROWS := len(pixels)
    COLS := len(pixels[0])
    for r := 0; r < ROWS; r++ {
        var row []color.RGBA
        for c := 0; c < COLS; c++ {
            if r == 0 || c == 0 || r == ROWS - 1 || c == COLS - 1 {
                x := uint8(float64(pixels[r][c].R) * 128 * math.Sqrt(2) / 256)
                row = append(row, color.RGBA{R:x, G:x, B:x, A:0xff})
                continue
            }
            var gy float64 = 0
            var gx float64 = 0
            // i dont know how to do matmul in go
            gy += float64(pixels[r - 1][c].R * 2)
            gy += float64(pixels[r - 1][c - 1].R)
            gy += float64(pixels[r - 1][c + 1].R)

            gy -= float64(pixels[r + 1][c].R * 2)
            gy -= float64(pixels[r + 1][c - 1].R)
            gy -= float64(pixels[r + 1][c + 1].R)

            gx += float64(pixels[r][c - 1].R * 2)
            gx += float64(pixels[r - 1][c - 1].R)
            gx += float64(pixels[r + 1][c - 1].R)

            gx -= float64(pixels[r][c + 1].R * 2)
            gx -= float64(pixels[r - 1][c + 1].R)
            gx -= float64(pixels[r + 1][c + 1].R)

            gy /= 4 // gy in [-256, 256]
            gx /= 4 // gx in [-256, 256]
            
            gy += 256 // gy in [0, 512]
            gx += 256 // gx in [0, 512]

            gy *= 128 * math.Sqrt(2) / 512 // gy in [0, 128sqrt(2)]
            gx *= 128 * math.Sqrt(2) / 512 // " " 

            // final pixel in [0, 256]
            p := uint8(math.Sqrt(gx * gx + gy + gy))
            final_color := color.RGBA{R:p, G:p, B:p, A:0xff}
            row = append(row, final_color)
        }
        gradient = append(gradient, row)
    }
    return gradient
}

func rgbaToGrayscale(r, g, b uint8) color.RGBA {
    // some actual nerd told me to do this and was lying...
    // log.Printf("%v, %v, %v, %v\n", r, g, b, a)
    // R := math.Pow(float64(r)/0xff, 1.0/1.0)
    // G := math.Pow(float64(g)/0xff, 1.0/1.0)
    // B := math.Pow(float64(b)/0xff, 1.0/1.0)
    // y := .2126 * R + .7152 * G + .0722 * B
    // l := 116 * math.Pow(y, 1.0/3.0) - 16
    l := uint8(0.299 * float64(r) + 0.587 * float64(g) + 0.114 * float64(b))
    gray := color.RGBA{R: l, G: l, B:l, A: 0xff}
    return gray
}

func imageToGrayscale(imageData [][]color.RGBA) [][]color.RGBA {
    ROWS, COLS := len(imageData), len(imageData[0])

    var pixels [][]color.RGBA
    for r := 0; r < ROWS; r++ {
        var row []color.RGBA
        for c := 0; c < COLS; c++ {
            R, G, B, _ := imageData[r][c].RGBA()
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
        // pixels are color.YCbCr... im not dealing with this
		imageData, err = jpeg.Decode(imageFile)
		if err != nil {
			log.Fatalf("[ERROR] Image (%s) could not be decoded\n", filepath)
		}
    default:
        log.Fatalf("[ERROR] Image suffix (%s) not supported\n", s)
	}

    return imageData
}

func gradToMinEnergy(pixels [][]color.RGBA) [][]float64 {
    ROWS, COLS := len(pixels), len(pixels[0])

    // energy of vertical seam
    var energy [][]float64

    // first row initialized
    var energy_row []float64
    for c := 0; c < COLS; c++ {
        energy_row = append(energy_row, float64(pixels[0][c].R))
    }
    energy = append(energy, energy_row)

    for r := 1; r < ROWS; r++ {
        var energy_row []float64
        for c := 0; c < COLS; c++ {
            // minEnergy from previous row neighbors
            var min_energy float64 = energy[r - 1][c]

            for dc := c - 1; dc <= c + 1; dc++ {
                if dc >= 0 && dc < COLS && energy[r - 1][dc] < min_energy {
                    min_energy = energy[r - 1][dc]
                }
            }

            energy_row = append(energy_row, min_energy + float64(pixels[r][c].R))
        }
        energy = append(energy, energy_row)
    }

    return energy
}

// return slice of idxs of vertical seam (with lowest energy) from pixels
func computeSeam(energy [][]float64) []int {
    ROWS, COLS := len(energy), len(energy[0])

    idxs := make([]int, ROWS)
    idxs[ROWS - 1] = argMin(energy[ROWS - 1])
    for r := ROWS - 2; r >= 0; r-- {
        prev_c := idxs[r + 1]
        min_energy := energy[r][prev_c]
        min_idx := prev_c

        for dc := -1; dc <= 1; dc++ {
            c := prev_c + dc
            if c >= 0 && c < COLS && energy[r][c] < min_energy {
                min_energy = energy[r][c]
                min_idx = c
            }
        }

        idxs[r] = min_idx
    }

    return idxs
}

func remove[T any](slice []T, idx int) []T {
    return append(slice[:idx], slice[idx+1:]...)
}

func removeSeam[T any](seam []int, pixels [][]T) {
    for r, c := range seam {
        pixels[r] = remove(pixels[r], c)
    }
}

// returns idx of max element in slice
func argMin[T cmp.Ordered](arr []T) int {
    minIdx := 0
    minEle := arr[0]
    for i, ele := range arr {
        if ele >= minEle {
            minEle = ele
            minIdx = i
        }
    }
    return minIdx
}

func imageToSlices(img image.Image) [][]color.RGBA {
    bounds := img.Bounds()
    width, height := bounds.Max.X, bounds.Max.Y

    var pixels [][]color.RGBA
    for r := 0; r < height; r++ {
        var row []color.RGBA
        for c := 0; c < width; c++ {
            r, g, b, _ := img.At(c, r).RGBA()
            R := uint8(r)
            G := uint8(g)
            B := uint8(b)
            row = append(row, color.RGBA{R:R, G:G, B:B, A:0xff})
        }
        pixels = append(pixels, row)
    }
    return pixels

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
    width, height := imageData.Bounds().Max.X, imageData.Bounds().Max.Y
    log.Printf("[INFO] Image size: %vx%v\n", width, height)

    imageSlice := imageToSlices(imageData)
    grayscaleImage := imageToGrayscale(imageSlice)
    gradient := grayscaleImageToGradient(grayscaleImage)
    energy := gradToMinEnergy(gradient)
    seam := computeSeam(energy)

    for i := 0; i < (4 * width / 5); i++ {
        log.Printf("[LOADING]... %v", i)
        removeSeam(seam, imageSlice)
        removeSeam(seam, gradient)
        energy = gradToMinEnergy(gradient)
        seam = computeSeam(energy)
    }
    pixelsToImage(imageSlice, "res.png")
}
