package main

import (
	"fmt"
	"image"
	_ "image/png"
	"log"
	"sync"

	"github.com/google/hilbert"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"golang.org/x/image/colornames"
)

const (
	W = 1024
	H = 1024
)

type IPVIZ struct {
	hilb       *hilbert.Hilbert
	ipChan     chan IPRecord
	picLock    sync.Mutex
	ipImage    *image.RGBA
	totalConns int
}

func NewIPVIZ() (*IPVIZ, error) {
	hilb, err := hilbert.NewHilbert(1024)
	if err != nil {
		return nil, err
	}

	ipImage := image.NewRGBA(image.Rect(0, 0, W, H))

	var picLock sync.Mutex
	ipChan := make(chan IPRecord, 100)
	viz := IPVIZ{
		hilb:    hilb,
		ipChan:  ipChan,
		ipImage: ipImage,
		picLock: picLock,
	}
	go listen("0.0.0.0", 9999, ipChan)
	go func() {
		for iprec := range ipChan {
			x, y, err := hilb.Map(int(iprec.ip / 256 / 16))
			if err != nil {
				log.Fatalf("Map failed: %v %d", err, (iprec.ip / 256 / 16))
			}
			picLock.Lock()
			if iprec.orig {
				ipImage.Set(x, y, colornames.Green)
			} else {
				ipImage.Set(x, y, colornames.Red)
			}
			viz.totalConns++
			picLock.Unlock()
		}
	}()
	return &viz, nil
}

func (v *IPVIZ) update(screen *ebiten.Image) error {
	if ebiten.IsDrawingSkipped() {
		return nil
	}
	mx, my := ebiten.CursorPosition()
	ip, err := v.hilb.MapInverse(mx, my)
	if err != nil {
		ip = 0
	}

	v.picLock.Lock()
	screen.ReplacePixels(v.ipImage.Pix)
	ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS: %0.2f. Conns: %d. IP=%s", ebiten.CurrentTPS(), v.totalConns, int2ip(uint32(ip*256*16))))
	v.picLock.Unlock()
	return nil
}

func main() {
	ipviz, err := NewIPVIZ()
	if err != nil {
		log.Fatal(err)
	}
	if err := ebiten.Run(ipviz.update, W, H, 1, "Noise (Ebiten Demo)"); err != nil {
		log.Fatal(err)
	}
}
