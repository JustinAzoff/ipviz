package main

import (
	"image"
	_ "image/png"
	"log"
	"os"
	"sync"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/google/hilbert"
	"golang.org/x/image/colornames"
)

const (
	W = 4096
	H = 4096
)

func loadPicture(path string) (pixel.Picture, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return pixel.PictureDataFromImage(img), nil
}

func run() {

	hilb, err := hilbert.NewHilbert(4096)
	if err != nil {
		log.Fatal(err)
	}

	ipChan := make(chan uint32, 100)
	go listen("0.0.0.0", 9999, ipChan)

	cfg := pixelgl.WindowConfig{
		Title:  "Pixel Rocks!",
		Bounds: pixel.R(0, 0, 1024, 768),
		VSync:  true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	win.Clear(colornames.White)

	/*
		im, err := loadPicture("scan.png")
		if err != nil {
			log.Fatal(err)
		}*/
	pic := pixel.MakePictureData(pixel.R(0, 0, W, H))
	//pic := pixel.PictureDataFromPicture(im)
	sprite := pixel.NewSprite(pic, pic.Bounds())

	var totalPoints int
	var picLock sync.Mutex
	go func() {
		for ip := range ipChan {
			totalPoints++
			x, y, err := hilb.Map(int(ip / 256))
			if err != nil {
				log.Fatal(err)
			}
			lx := float64(x)
			ly := float64(y)
			picLock.Lock()
			for xx := lx; xx < lx+20 && xx < W; xx++ {
				for yy := ly; yy < ly+20 && yy < H; yy++ {
					pi := pic.Index(pixel.V(xx, yy))
					pic.Pix[pi] = colornames.Black
				}
			}
			picLock.Unlock()
		}
	}()

	var ticker = time.Tick(1000 * time.Millisecond)
	for !win.Closed() {
		if win.JustPressed(pixelgl.KeyQ) {
			return
		}

		//Scale
		win_bounds := win.Bounds().Max
		win_width := win_bounds.X
		win_height := win_bounds.Y
		bounds := sprite.Frame().Max
		width := bounds.X
		height := bounds.Y

		xratio := float64(win_width) / width
		yratio := float64(win_height) / height

		var scale float64
		if xratio > 1 && yratio > 1 {
			scale = 1
		} else if xratio < yratio {
			scale = xratio
		} else {
			scale = yratio
		}
		//End scale

		select {
		case <-ticker:
			log.Printf("Rendering %d points", totalPoints)
			picLock.Lock()
			sprite = pixel.NewSprite(pic, pic.Bounds())
			picLock.Unlock()
		default:
		}

		sprite.Draw(win, pixel.IM.Scaled(pixel.ZV, scale).Moved(win.Bounds().Center()))
		win.Update()
	}
}

func main() {
	pixelgl.Run(run)
}
