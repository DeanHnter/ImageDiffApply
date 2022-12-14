package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"
)

func HandleError(msg string, err error) {
	if err != nil {
		if len(msg) > 0 {
			os.Exit(1)
		}
		os.Exit(1)
	}
}

func main() {
	// Create two arrays of the same size representing two images
	imageA, err := LoadPNG("githubsmall.png")
	HandleError("", err)
	imageB, err := LoadPNG("githubfull.png")
	HandleError("", err)

	//first parameter image should be smaller or equal in size to right
	output := InterpolateResizeImage(imageA, imageB) //make image a the size of image b
	diffmask := DifferenceImageRGBA(&imageB, &output)
	HandleError("", err)
	appliedimg := ApplyDifferenceImageRGBA(&output, diffmask)
	//save result
	saveImage("diffmask.png", diffmask)
	saveImage("diffapplied.png", appliedimg)
}

func DiffImage(img1, img2 *image.RGBA) *image.RGBA {
	bounds := img1.Bounds()
	out := image.NewRGBA(bounds)
	draw.Draw(out, bounds, img1, image.Point{}, draw.Src)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c1 := img1.At(x, y)
			c2 := img2.At(x, y)
			r1, g1, b1, a1 := c1.RGBA()
			r2, g2, b2, a2 := c2.RGBA()
			out.Set(x, y, color.RGBA{
				uint8(r1 - r2),
				uint8(g1 - g2),
				uint8(b1 - b2),
				uint8(a1 - a2),
			})
		}
	}
	return out
}

func DifferenceImageRGBA(img1, img2 *image.RGBA) *image.RGBA {
	// get image width and height
	bounds := img1.Bounds()
	width := bounds.Max.X
	height := bounds.Max.Y

	// create new image with same size as source images
	differenceImg := image.NewRGBA(image.Rect(0, 0, width, height))

	// loop through all pixels in both images
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			// get pixel from both images
			p1 := img1.At(x, y)
			p2 := img2.At(x, y)

			// calculate difference in each color channel
			r1, g1, b1, _ := p1.RGBA()
			r2, g2, b2, _ := p2.RGBA()
			rDiff := uint8(r1 - r2)
			gDiff := uint8(g1 - g2)
			bDiff := uint8(b1 - b2)
			//aDiff := uint8(a1 - a2)
			//A: aDiff
			// set pixel in new image
			differenceImg.Set(x, y, color.RGBA{R: rDiff, G: gDiff, B: bDiff, A: 255})
		}
	}

	return differenceImg
}

func ApplyDifferenceImageRGBA(img *image.RGBA, diffImage *image.RGBA) *image.RGBA {
	// get image width and height
	bounds := img.Bounds()
	width := bounds.Max.X
	height := bounds.Max.Y

	// create new image with same size as source images
	appliedImg := image.NewRGBA(image.Rect(0, 0, width, height))

	// loop through all pixels in both images
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			// get pixel from both images
			p := img.At(x, y)
			diffP := diffImage.At(x, y)

			// calculate difference in each color channel
			r, g, b, _ := p.RGBA()
			diffR, diffG, diffB, _ := diffP.RGBA()
			rDiff := uint8(r + diffR)
			gDiff := uint8(g + diffG)
			bDiff := uint8(b + diffB)
			//aDiff := uint8(a1 + diffA)
			//A: aDiff
			// set pixel in new image
			appliedImg.Set(x, y, color.RGBA{R: rDiff, G: gDiff, B: bDiff, A: 255})
		}
	}

	return appliedImg
}

func LoadPNG(file string) (image.RGBA, error) {
	var imgout *image.RGBA
	reader, err := os.Open(file)
	if err != nil {
		return *imgout, err
	}
	defer reader.Close()

	img, err := png.Decode(reader)
	if err != nil {
		return *imgout, err
	}
	bounds := img.Bounds()
	imgout = image.NewRGBA(bounds)
	for x := 0; x < bounds.Size().X; x++ {
		for y := 0; y < bounds.Size().Y; y++ {
			imgout.Set(x, y, img.At(x, y))
		}
	}
	return *imgout, nil
}

func InterpolateResizeImage(img1, img2 image.RGBA) image.RGBA {
	// get width and height of both images
	w1, h1 := img1.Bounds().Dx(), img1.Bounds().Dy()
	w2, h2 := img2.Bounds().Dx(), img2.Bounds().Dy()
	// compare width and height
	maxwidth, maxheight := 0, 0
	if w1 <= w2 {
		maxwidth = w2
	} else {
		//maxwidth = w2 disabled - was used to interop a merged size image
		panic("Input image is larger in width than the target!")
	}
	if h1 <= h2 {
		maxheight = h2
	} else {
		//maxheight = h2 disabled - was used to interop a merged size image
		panic("Input image is larger in height than the target!")
	}
	// create new image
	newImg := image.NewRGBA(image.Rect(0, 0, maxwidth, maxheight))

	// Iterate over the pixels in the new image and interpolate resize it
	for x := 0; x < maxwidth; x++ {
		for y := 0; y < maxheight; y++ {
			// Calculate the position in the original image
			origX := math.Round(float64(x) * float64(img1.Bounds().Dx()) / float64(maxwidth))
			origY := math.Round(float64(y) * float64(img1.Bounds().Dy()) / float64(maxheight))

			// Get the color of the original pixel
			c := img1.At(int(origX), int(origY))
			if c, ok := c.(color.RGBA); ok {
				// Set the color of the new pixel
				draw.Draw(newImg, image.Rect(x, y, x+1, y+1), &image.Uniform{c}, image.ZP, draw.Src)
			}
		}
	}

	return *newImg
}

func saveImage(fileName string, img *image.RGBA) {
	// Create the file
	f, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// Write the image to the file
	err = png.Encode(f, img)
	if err != nil {
		panic(err)
	}
}
