//go:build windows

package main

import (
	"fmt"
	"runtime"
	"unsafe"
)

var updateLayeredWindow = user32.NewProc("UpdateLayeredWindow")

type nativeSurface struct {
	*PixelBuffer
	DC, Bitmap, Old uintptr
}

func newNativeSurface(width, height int) (*nativeSurface, error) {
	if width < 1 || height < 1 || width > 8192 || height > 8192 {
		return nil, fmt.Errorf("invalid surface %dx%d", width, height)
	}
	bi := BitmapInfo{Size: 40, Width: int32(width), Height: -int32(height), Planes: 1, BitCount: 32}
	var ptr unsafe.Pointer
	bm, _, err := createDIBSection.Call(0, uintptr(unsafe.Pointer(&bi)), 0, uintptr(unsafe.Pointer(&ptr)), 0, 0)
	if bm == 0 || ptr == nil {
		return nil, fmt.Errorf("CreateDIBSection: %v", err)
	}
	dc, _, err := createCompatibleDC.Call(0)
	if dc == 0 {
		deleteObject.Call(bm)
		return nil, fmt.Errorf("CreateCompatibleDC: %v", err)
	}
	old, _, _ := selectObject.Call(dc, bm)
	pixels := unsafe.Slice((*byte)(ptr), width*height*4)
	clear(pixels)
	return &nativeSurface{&PixelBuffer{pixels, width, height}, dc, bm, old}, nil
}
func (s *nativeSurface) Close() {
	if s == nil || s.DC == 0 {
		return
	}
	selectObject.Call(s.DC, s.Old)
	deleteObject.Call(s.Bitmap)
	deleteDC.Call(s.DC)
	s.DC = 0
	s.Bitmap = 0
	s.Pix = nil
}
func (s *nativeSurface) Present(hwnd uintptr) error {
	p := Point{}
	size := Point{int32(s.W), int32(s.H)}
	blend := [4]byte{0, 0, 255, 1} // AC_SRC_OVER, flags=0, alpha=255, AC_SRC_ALPHA.
	// No SetLayeredWindowAttributes/colour-key call: it would disable this path.
	ok, _, err := updateLayeredWindow.Call(hwnd, 0, 0, uintptr(unsafe.Pointer(&size)), s.DC, uintptr(unsafe.Pointer(&p)), 0, uintptr(unsafe.Pointer(&blend)), 2)
	runtime.KeepAlive(s)
	if ok == 0 {
		return fmt.Errorf("UpdateLayeredWindow: %v", err)
	}
	return nil
}

type glyphMask struct {
	Alpha []byte
	W, H  int
}

var glyphMasks = map[string]glyphMask{}

func clearGlyphMasks() { glyphMasks = map[string]glyphMask{} }
func (a *App) glyph(text string, width, height, size int, bold bool, flags int) (glyphMask, error) {
	key := fmt.Sprintf("%d:%d:%d:%d:%t:%d:%s", a.DPI, width, height, size, bold, flags, text)
	if m, ok := glyphMasks[key]; ok {
		return m, nil
	}
	s, e := newNativeSurface(width, height)
	if e != nil {
		return glyphMask{}, e
	}
	defer s.Close()
	// White grayscale glyphs on zero RGB become a coverage mask. This avoids
	// ClearType's colored fringe on a transparent Windows taskbar.
	r := R(0, 0, width, height)
	a.text(s.DC, text, r, size, bold, 0xFFFFFF, flags|DT_SINGLELINE|DT_VCENTER)
	gdiFlush.Call()
	m := glyphMask{make([]byte, width*height), width, height}
	for i := 0; i < width*height; i++ {
		m.Alpha[i] = byte(max(int(s.Pix[i*4]), max(int(s.Pix[i*4+1]), int(s.Pix[i*4+2]))))
	}
	if len(glyphMasks) >= 512 {
		clearGlyphMasks()
	}
	glyphMasks[key] = m
	return m, nil
}
func (a *App) canvasText(dst *PixelBuffer, text string, r Rect, size int, bold bool, col uint32, flags int, clip Rect) {
	if r.W() < 1 || r.H() < 1 {
		return
	}
	m, e := a.glyph(text, r.W(), r.H(), size, bold, flags)
	if e != nil {
		logError(e)
		return
	}
	dst.Mask(m.Alpha, m.W, m.H, int(r.Left), int(r.Top), col, clip)
}
func canvasArt(dst *PixelBuffer, name string, r Rect, clip Rect) {
	if r.W() < 1 || r.H() < 1 {
		return
	}
	v, e := artBitmap(name, r.W(), r.H())
	if e != nil {
		logError(e)
		return
	}
	dst.Blit(&PixelBuffer{v.Pixels, v.W, v.H}, int(r.Left), int(r.Top), clip, nil, 0)
}
func canvasPanel(dst *PixelBuffer, name string, r Rect, rad int, base, border uint32, clip Rect) {
	dst.RoundRect(r, float64(rad), base, clip)
	// Artwork keeps its aspect ratio and stays at the top. Longer explanations
	// extend the quiet reading field instead of stretching the moon or mountains.
	v, e := artBitmap(name, r.W(), max(1, r.W()*500/960))
	if e == nil {
		cl := intersectRect(clip, r)
		dst.Blit(&PixelBuffer{v.Pixels, v.W, v.H}, int(r.Left), int(r.Top), cl, &r, float64(rad))
	} else {
		logError(e)
	}
	dst.RoundFrame(r, float64(rad), 1, border, clip)
}
