package main

import (
	"image"
	"math"
)

// PixelBuffer uses premultiplied BGRA, the exact format consumed by
// UpdateLayeredWindow. The compositor is also exercised by portable tests.
type PixelBuffer struct {
	Pix  []byte
	W, H int
}

func NewPixelBuffer(w, h int) *PixelBuffer {
	if w < 1 || h < 1 || w > 8192 || h > 8192 {
		return nil
	}
	return &PixelBuffer{make([]byte, w*h*4), w, h}
}
func (b *PixelBuffer) Bounds() Rect { return R(0, 0, b.W, b.H) }
func intersectRect(a, b Rect) Rect {
	return Rect{int32(max(int(a.Left), int(b.Left))), int32(max(int(a.Top), int(b.Top))), int32(min(int(a.Right), int(b.Right))), int32(min(int(a.Bottom), int(b.Bottom)))}
}
func coverage(v float64) uint32 {
	if v <= 0 {
		return 0
	}
	if v >= 1 {
		return 255
	}
	return uint32(v*255 + .5)
}
func (b *PixelBuffer) over(x, y int, blue, green, red, alpha uint32) {
	if alpha == 0 || x < 0 || y < 0 || x >= b.W || y >= b.H {
		return
	}
	i := (y*b.W + x) * 4
	inv := 255 - alpha
	b.Pix[i] = byte(min(255, int(blue+(uint32(b.Pix[i])*inv+127)/255)))
	b.Pix[i+1] = byte(min(255, int(green+(uint32(b.Pix[i+1])*inv+127)/255)))
	b.Pix[i+2] = byte(min(255, int(red+(uint32(b.Pix[i+2])*inv+127)/255)))
	b.Pix[i+3] = byte(min(255, int(alpha+(uint32(b.Pix[i+3])*inv+127)/255)))
}
func (b *PixelBuffer) solid(x, y int, hex, alpha uint32) {
	b.over(x, y, ((hex&255)*alpha+127)/255, (((hex>>8)&255)*alpha+127)/255, (((hex>>16)&255)*alpha+127)/255, alpha)
}
func roundDistance(x, y float64, r Rect, rad float64) float64 {
	cx := (float64(r.Left) + float64(r.Right)) / 2
	cy := (float64(r.Top) + float64(r.Bottom)) / 2
	rad = math.Min(rad, math.Min(float64(r.W())/2, float64(r.H())/2))
	x = math.Abs(x-cx) - float64(r.W())/2 + rad
	y = math.Abs(y-cy) - float64(r.H())/2 + rad
	return math.Hypot(math.Max(x, 0), math.Max(y, 0)) + math.Min(math.Max(x, y), 0) - rad
}
func (b *PixelBuffer) RoundRect(r Rect, rad float64, col uint32, clip Rect) {
	area := intersectRect(intersectRect(r, b.Bounds()), clip)
	for y := int(area.Top); y < int(area.Bottom); y++ {
		for x := int(area.Left); x < int(area.Right); x++ {
			a := coverage(.5 - roundDistance(float64(x)+.5, float64(y)+.5, r, rad))
			b.solid(x, y, col, a)
		}
	}
}
func (b *PixelBuffer) RoundFrame(r Rect, rad, thickness float64, col uint32, clip Rect) {
	area := intersectRect(intersectRect(r, b.Bounds()), clip)
	for y := int(area.Top); y < int(area.Bottom); y++ {
		for x := int(area.Left); x < int(area.Right); x++ {
			d := roundDistance(float64(x)+.5, float64(y)+.5, r, rad)
			a := coverage(.5 - d)
			in := coverage(.5 - d - thickness)
			if a > in {
				b.solid(x, y, col, a-in)
			}
		}
	}
}
func (b *PixelBuffer) Blit(src *PixelBuffer, x0, y0 int, clip Rect, round *Rect, rad float64) {
	if src == nil {
		return
	}
	area := intersectRect(intersectRect(R(x0, y0, src.W, src.H), b.Bounds()), clip)
	for y := int(area.Top); y < int(area.Bottom); y++ {
		for x := int(area.Left); x < int(area.Right); x++ {
			i := ((y-y0)*src.W + x - x0) * 4
			a := uint32(src.Pix[i+3])
			if a == 0 {
				continue
			}
			m := uint32(255)
			if round != nil {
				m = coverage(.5 - roundDistance(float64(x)+.5, float64(y)+.5, *round, rad))
				if m == 0 {
					continue
				}
			}
			b.over(x, y, (uint32(src.Pix[i])*m+127)/255, (uint32(src.Pix[i+1])*m+127)/255, (uint32(src.Pix[i+2])*m+127)/255, (a*m+127)/255)
		}
	}
}
func (b *PixelBuffer) Mask(alpha []byte, w, h, x0, y0 int, col uint32, clip Rect) {
	if len(alpha) < w*h {
		return
	}
	area := intersectRect(intersectRect(R(x0, y0, w, h), b.Bounds()), clip)
	for y := int(area.Top); y < int(area.Bottom); y++ {
		for x := int(area.Left); x < int(area.Right); x++ {
			b.solid(x, y, col, uint32(alpha[(y-y0)*w+x-x0]))
		}
	}
}
func (b *PixelBuffer) Line(x1, y1, x2, y2, thickness float64, col uint32, clip Rect) {
	r := R(int(math.Floor(math.Min(x1, x2)-thickness)), int(math.Floor(math.Min(y1, y2)-thickness)), int(math.Ceil(math.Abs(x2-x1)+2*thickness))+1, int(math.Ceil(math.Abs(y2-y1)+2*thickness))+1)
	a := intersectRect(intersectRect(r, b.Bounds()), clip)
	vx, vy := x2-x1, y2-y1
	den := vx*vx + vy*vy
	for y := int(a.Top); y < int(a.Bottom); y++ {
		for x := int(a.Left); x < int(a.Right); x++ {
			px, py := float64(x)+.5-x1, float64(y)+.5-y1
			t := 0.
			if den > 0 {
				t = math.Max(0, math.Min(1, (px*vx+py*vy)/den))
			}
			d := math.Hypot(px-t*vx, py-t*vy)
			b.solid(x, y, col, coverage(thickness/2+.5-d))
		}
	}
}
func (b *PixelBuffer) Image() *image.RGBA {
	out := image.NewRGBA(image.Rect(0, 0, b.W, b.H))
	for i := 0; i < len(b.Pix); i += 4 {
		out.Pix[i], out.Pix[i+1], out.Pix[i+2], out.Pix[i+3] = b.Pix[i+2], b.Pix[i+1], b.Pix[i], b.Pix[i+3]
	}
	return out
}
