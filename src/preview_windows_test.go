//go:build windows

package main

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestNativePreviewLayoutsAndRenders(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	surface, err := newNativeSurface(1, 1)
	if err != nil {
		t.Fatal(err)
	}
	defer surface.Close()
	data := DefaultData()
	checked := 0
	for _, lang := range []string{"zh-CN", "en"} {
		for _, light := range []bool{false, true} {
			a := &App{L: NewLocalizer(lang), Data: data, Settings: DefaultSettings(), DPI: 192, Fonts: map[string]uintptr{}, Colors: darkPalette, TaskbarLight: light}
			if light {
				a.Colors = lightPalette
			}
			m := func(s string, sz int, bold bool) int { return a.textWidth(surface.DC, s, sz, bold) }
			for _, width := range []int{240, 320} {
				for day := 0; day < 7; day++ {
					for hour := 0; hour < 12; hour++ {
						now := time.Date(2026, 9, 6+day, hour*2, 30, 0, 0, time.UTC)
						opt := previewOptions(data, now, lang, light, a.DPI, width, PreviewState{Days: true, Hours: true, DayRow: 1, HourRow: 1})
						opt.MaxHeight = 3200
						layout := CompactLayout(data.DayAt(now), data.HourAt(now), opt, m)
						for _, b := range layout.Blocks {
							if w := m(b.Text, b.Size, b.Bold); w > b.Rect.W()+2 {
								t.Fatalf("native preview clips: %s width%d %q: %d > %d", lang, width, b.Text, w, b.Rect.W())
							}
						}
						checked++
					}
				}
			}
			dir := os.Getenv("ASTRAL_DOCS_DIR")
			if dir != "" {
				if err = os.MkdirAll(dir, 0755); err != nil {
					t.Fatal(err)
				}
				mode := "dark"
				if light {
					mode = "light"
				}
				now := time.Date(2026, 9, 14, 9, 30, 0, 0, time.UTC)
				for _, state := range []struct {
					name string
					p    PreviewState
				}{{"preview", PreviewState{Days: true, Hours: true}}, {"preview-hours", PreviewState{Hours: true}}, {"preview-advice", PreviewState{Hours: true, HourRow: 1}}} {
					opt := previewOptions(data, now, lang, light, a.DPI, 320, state.p)
					opt.DayDetails = false
					opt.MaxHeight = 3200
					a.Preview = state.p
					layout := CompactLayout(data.DayAt(now), data.HourAt(now), opt, m)
					a.Blocks, a.Panels, a.DrawIcons, a.Buttons, a.Body = layout.Blocks, layout.Panels, layout.Icons, layout.Buttons, layout.Body
					a.CardWidth, a.CardHeight, a.ContentHeight = layout.Width, layout.Height, layout.ContentHeight
					if layout.Height != layout.ContentHeight {
						t.Fatal("export clipped")
					}
					b := NewPixelBuffer(layout.Width, layout.Height)
					a.paintCard(b)
					writeDocumentationPNG(t, filepath.Join(dir, state.name+"-"+lang+"-"+mode+".png"), b)
				}
			}
			a.clearFonts()
		}
	}
	t.Logf("%d preview layouts checked against actual Windows GDI font metrics; native offscreen exports, not desktop interaction tests", checked)
}

func TestPreviewNativeRefreshPreservesDrawers(t *testing.T) {
	a := &App{Data: DefaultData(), L: NewLocalizer("zh-CN"), Colors: darkPalette}
	now := time.Date(2026, 9, 14, 9, 30, 0, 0, time.UTC)
	opt := CardOptions{}
	a.applyPreviewOptions(&opt, now)
	a.Preview = PreviewState{Days: true, Hours: true, DayRow: 1, HourRow: 2}
	a.applyPreviewOptions(&opt, now.Add(time.Minute))
	if !opt.Preview.Open(101) || !opt.Preview.Open(202) {
		t.Fatal("minute refresh reset a reading session")
	}
	a.applyPreviewOptions(&opt, now.Add(2*time.Hour))
	if !opt.Preview.Days || !opt.Preview.Hours || opt.Preview.DayRow != 0 || opt.Preview.HourRow != 0 {
		t.Fatal("rollover selection drift")
	}
	if opt.UpcomingHours[0].Name != "未时" {
		t.Fatal("not using fresh time")
	}
	a.Preview.Reset()
	a.applyPreviewOptions(&opt, now)
	if opt.Preview.Active() {
		t.Fatal("reopened popup is not focused on now")
	}
}
