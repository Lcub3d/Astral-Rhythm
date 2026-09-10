//go:build windows

package main

import (
 "encoding/json"
 "image"
 "image/png"
 "os"
 "path/filepath"
 "runtime"
 "testing"
 "time"
)

// Export the application's own GDI text and pixel compositor, not the desktop.
func TestNativeDocumentationRenders(t *testing.T) {
 dir := os.Getenv("ASTRAL_DOCS_DIR")
 if dir == "" { t.Skip("optional native documentation export") }
 runtime.LockOSThread()
 defer runtime.UnlockOSThread()
 if e := os.MkdirAll(dir, 0755); e != nil { t.Fatal(e) }
 now := time.Date(2026, 9, 8, 13, 40, 0, 0, time.UTC)
 var records []map[string]any
 for _, lang := range []string{"zh-CN", "en"} {
  for _, light := range []bool{false, true} {
   mode := "dark"
   if light { mode = "light" }
   name := lang + "-" + mode
   a := &App{L: NewLocalizer(lang), Data: DefaultData(), Settings: DefaultSettings(), DPI: 192, Fonts: map[string]uintptr{}, Colors: darkPalette, TaskbarLight: light}
   if light { a.Colors = lightPalette }
   surface, e := newNativeSurface(1, 1)
   if e != nil { t.Fatal(e) }
   layout := CompactLayout(a.Data.DayAt(now), a.Data.HourAt(now), CardOptions{Locale: a.L, DPI: a.DPI, MaxHeight: 2600, Light: light, HourDetails: true}, func(s string, size int, bold bool) int { return a.textWidth(surface.DC, s, size, bold) })
   a.Blocks, a.Panels, a.DrawIcons, a.Buttons, a.Body = layout.Blocks, layout.Panels, layout.Icons, layout.Buttons, layout.Body
   a.CardWidth, a.CardHeight, a.ContentHeight = layout.Width, layout.Height, layout.ContentHeight
   if layout.Height != layout.ContentHeight { t.Fatal("documentation card would be truncated") }
   for _, block := range layout.Blocks {
    if width := a.textWidth(surface.DC, block.Text, block.Size, block.Bold); width > block.Rect.W()+2 {
     t.Fatalf("native text clips in %s: %q (%d > %d)", name, block.Text, width, block.Rect.W())
    }
   }
   card := NewPixelBuffer(layout.Width, layout.Height)
   a.paintCard(card)
   writeDocumentationPNG(t, filepath.Join(dir, "card-"+name+".png"), card)
   widgetWidth := a.S(78)
   if lang == "en" { widgetWidth = a.S(112) }
   widget := NewPixelBuffer(widgetWidth, a.S(48))
   a.paintWidgetAt(widget, now)
   if widget.Pix[3] > 1 { t.Fatal("transparent taskbar has a visible background") }
   writeDocumentationPNG(t, filepath.Join(dir, "taskbar-"+name+".png"), widget)
   records = append(records, map[string]any{"name": name, "width": card.W, "height": card.H, "dpi": a.DPI, "renderer": "Win32 GDI + application pixel compositor", "fixture": now.Format("2006-01-02 15:04"), "desktop_capture": false})
   surface.Close()
   a.clearFonts()
  }
 }
 metadata, e := json.MarshalIndent(records, "", "  ")
 if e != nil { t.Fatal(e) }
 if e = os.WriteFile(filepath.Join(dir, "render-metadata.json"), metadata, 0644); e != nil { t.Fatal(e) }
 t.Log("8 native PNGs exported using real Windows font metrics; offscreen UI renders, not desktop screenshots")
}

func writeDocumentationPNG(t *testing.T, path string, b *PixelBuffer) {
 t.Helper()
 im := image.NewRGBA(image.Rect(0, 0, b.W, b.H))
 for i := 0; i < len(b.Pix); i += 4 {
  im.Pix[i], im.Pix[i+1], im.Pix[i+2], im.Pix[i+3] = b.Pix[i+2], b.Pix[i+1], b.Pix[i], b.Pix[i+3]
 }
 f, e := os.Create(path)
 if e != nil { t.Fatal(e) }
 if e = png.Encode(f, im); e != nil { f.Close(); t.Fatal(e) }
 if e = f.Close(); e != nil { t.Fatal(e) }
}
