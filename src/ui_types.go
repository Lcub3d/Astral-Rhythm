package main

type Palette struct{ BG, Panel, Text, Muted, Accent, Border, Soft uint32 }

var darkPalette = Palette{0x142027, 0x111A20, 0xDCE3E6, 0x9DACB3, 0xA3CDBD, 0x2C3A43, 0x23333B}
var lightPalette = Palette{0xF1F4F5, 0xFCFDFD, 0x283941, 0x697981, 0x3E806B, 0xE1E8E8, 0xEDF3F1}

type TextBlock struct {
	Text  string
	Rect  Rect
	Size  int
	Bold  bool
	Color uint32
}
type Button struct {
	InBody bool
	ID     int
	Text   string
	Rect   Rect
}
