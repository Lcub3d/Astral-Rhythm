//go:build windows

package main

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"math"
	"runtime"
	"strings"
	"syscall"
	"unsafe"
)

var (
	alphaBlend     = syscall.NewLazyDLL("msimg32.dll").NewProc("AlphaBlend")
	getStockObject = gdi32.NewProc("GetStockObject")
	getTextExtent  = gdi32.NewProc("GetTextExtentPoint32W")
	moveToEx       = gdi32.NewProc("MoveToEx")
	lineTo         = gdi32.NewProc("LineTo")
)

type nativeArt struct {
	DC, Bitmap, Old uintptr
	W, H            int
	Pixels          []byte
}

var artSources = map[string]image.Image{}
var artCache = map[string]nativeArt{}

func clearArtCache() {
	for _, v := range artCache {
		selectObject.Call(v.DC, v.Old)
		deleteObject.Call(v.Bitmap)
		deleteDC.Call(v.DC)
	}
	artCache = map[string]nativeArt{}
}

// Resample in premultiplied RGBA, so the artwork stays clean at 125/150/200% DPI.
func artBitmap(name string, width, height int) (nativeArt, error) {
	key := fmt.Sprintf("%s/%d/%d", name, width, height)
	if v, ok := artCache[key]; ok {
		return v, nil
	}
	if width < 1 || height < 1 || width > 4096 || height > 4096 {
		return nativeArt{}, fmt.Errorf("invalid artwork size")
	}
	im := artSources[name]
	if im == nil {
		b, e := artwork.ReadFile(name)
		if e != nil {
			return nativeArt{}, e
		}
		im, e = png.Decode(bytes.NewReader(b))
		if e != nil {
			return nativeArt{}, e
		}
		if !strings.Contains(name, "/backgrounds/") {
			artSources[name] = im
		}
	}
	bi := BitmapInfo{Size: 40, Width: int32(width), Height: -int32(height), Planes: 1, BitCount: 32}
	var p unsafe.Pointer
	bm, _, _ := createDIBSection.Call(0, uintptr(unsafe.Pointer(&bi)), 0, uintptr(unsafe.Pointer(&p)), 0, 0)
	if bm == 0 || p == nil {
		return nativeArt{}, fmt.Errorf("cannot allocate artwork bitmap")
	}
	// Pointer belongs to the Win32 DIB allocation, not the Go heap.
	dst := unsafe.Slice((*byte)(p), width*height*4)
	b := im.Bounds()
	sw, sh := b.Dx(), b.Dy()
	sample := func(x, y int) (float64, float64, float64, float64) {
		x = clamp(x, 0, sw-1)
		y = clamp(y, 0, sh-1)
		r, g, bl, a := im.At(b.Min.X+x, b.Min.Y+y).RGBA()
		return float64(r) / 257, float64(g) / 257, float64(bl) / 257, float64(a) / 257
	}
	for y := 0; y < height; y++ {
		sy := (float64(y)+.5)*float64(sh)/float64(height) - .5
		y0 := int(math.Floor(sy))
		fy := sy - float64(y0)
		for x := 0; x < width; x++ {
			sx := (float64(x)+.5)*float64(sw)/float64(width) - .5
			x0 := int(math.Floor(sx))
			fx := sx - float64(x0)
			r0, g0, b0, a0 := sample(x0, y0)
			r1, g1, b1, a1 := sample(x0+1, y0)
			r2, g2, b2, a2 := sample(x0, y0+1)
			r3, g3, b3, a3 := sample(x0+1, y0+1)
			blend := func(v0, v1, v2, v3 float64) byte {
				return byte(math.Round((v0*(1-fx)+v1*fx)*(1-fy) + (v2*(1-fx)+v3*fx)*fy))
			}
			off := (y*width + x) * 4
			dst[off] = blend(b0, b1, b2, b3)
			dst[off+1] = blend(g0, g1, g2, g3)
			dst[off+2] = blend(r0, r1, r2, r3)
			dst[off+3] = blend(a0, a1, a2, a3)
		}
	}
	dc, _, _ := createCompatibleDC.Call(0)
	if dc == 0 {
		deleteObject.Call(bm)
		return nativeArt{}, fmt.Errorf("cannot allocate artwork DC")
	}
	old, _, _ := selectObject.Call(dc, bm)
	v := nativeArt{dc, bm, old, width, height, dst}
	artCache[key] = v
	return v, nil
}
func drawArt(dc uintptr, name string, r Rect) {
	if r.W() < 1 || r.H() < 1 {
		return
	}
	v, e := artBitmap(name, r.W(), r.H())
	if e != nil {
		logError(e)
		return
	}
	alphaBlend.Call(dc, u(int(r.Left)), u(int(r.Top)), u(r.W()), u(r.H()), v.DC, 0, 0, u(v.W), u(v.H), 0x01ff0000)
}
func artPanel(dc uintptr, name string, r Rect, radius int) {
	saved, _, _ := saveDC.Call(dc)
	region, _, _ := createRoundRectRgn.Call(u(int(r.Left)), u(int(r.Top)), u(int(r.Right)+1), u(int(r.Bottom)+1), u(radius*2), u(radius*2))
	if region != 0 { // Combine with any existing scroll clip instead of replacing it.
		gdi32.NewProc("ExtSelectClipRgn").Call(dc, region, 1)
		deleteObject.Call(region)
	}
	drawArt(dc, name, r)
	restoreDC.Call(dc, saved)
}
func strokeLine(dc uintptr, x1, y1, x2, y2 int, c uint32, width int) {
	p, _, _ := createPen.Call(0, u(max(1, width)), color(c))
	old, _, _ := selectObject.Call(dc, p)
	moveToEx.Call(dc, u(x1), u(y1), 0)
	lineTo.Call(dc, u(x2), u(y2))
	selectObject.Call(dc, old)
	deleteObject.Call(p)
}

func frameRound(dc uintptr, r Rect, c uint32, rad int) {
	pen, _, _ := createPen.Call(0, 1, color(c))
	brush, _, _ := getStockObject.Call(5)
	op, _, _ := selectObject.Call(dc, pen)
	ob, _, _ := selectObject.Call(dc, brush)
	roundRect.Call(dc, u(int(r.Left)), u(int(r.Top)), u(int(r.Right)), u(int(r.Bottom)), u(rad*2), u(rad*2))
	selectObject.Call(dc, op)
	selectObject.Call(dc, ob)
	deleteObject.Call(pen)
}
func (a *App) isLight() bool { return a.Colors.Panel == lightPalette.Panel }
func (a *App) textWidth(dc uintptr, s string, size int, bold bool) int {
	var z struct{ X, Y int32 }
	v, _ := syscall.UTF16FromString(s)
	if len(v) <= 1 {
		return 0
	}
	old, _, _ := selectObject.Call(dc, a.font(size, bold))
	getTextExtent.Call(dc, uintptr(unsafe.Pointer(&v[0])), u(len(v)-1), uintptr(unsafe.Pointer(&z)))
	selectObject.Call(dc, old)
	runtime.KeepAlive(v)
	return int(z.X)
}
func blendColor(a, b uint32, t float64) uint32 {
	var r uint32
	for _, s := range []uint{0, 8, 16} {
		v := math.Round(float64((a>>s)&255)*(1-t) + float64((b>>s)&255)*t)
		r |= uint32(v) << s
	}
	return r
}
