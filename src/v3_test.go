package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode"
)

func testMeasure(dpi int) MeasureText {
	return func(s string, size int, bold bool) int {
		w := 0.
		for _, r := range s {
			if r < 128 {
				if r == ' ' {
					w += float64(size) * .30
				} else {
					w += float64(size) * .57
				}
			} else {
				w += float64(size)
			}
		}
		return int(w*float64(dpi)/96 + .5)
	}
}
func compact(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, s)
}
func sectionString(l CardLayout, kind string) string {
	var r Rect
	for _, p := range l.Panels {
		if p.Kind == kind {
			r = p.Rect
			break
		}
	}
	var b strings.Builder
	for _, v := range l.Blocks {
		if v.Rect.Top >= r.Top && v.Rect.Bottom <= r.Bottom {
			b.WriteString(v.Text)
		}
	}
	return compact(b.String())
}
func TestV3AllSevenDayElements(t *testing.T) {
	want := []string{"火", "水", "火", "水", "木", "金", "土"}
	for i, v := range want {
		if DayElement(i) != v {
			t.Fatalf("weekday %d", i)
		}
	}
	if DayElement(-1) != "" || DayElement(7) != "" {
		t.Fatal("index bounds")
	}
}
func TestV3AllTwelveBranchElements(t *testing.T) {
	want := []string{"水", "土", "木", "木", "土", "火", "火", "土", "金", "金", "土", "水"}
	for i, v := range want {
		if HourElement(i) != v {
			t.Fatalf("branch %d", i)
		}
	}
	if HourElement(-1) != "" || HourElement(12) != "" {
		t.Fatal("index bounds")
	}
}
func TestV3ColorsStayInElementFamilies(t *testing.T) {
	for _, kind := range []string{"day", "hour"} {
		n := 7
		if kind == "hour" {
			n = 12
		}
		for i := 0; i < n; i++ {
			for _, light := range []bool{false, true} {
				el, col := DayElement(i), DayTextColor(i, light)
				if kind == "hour" {
					el, col = HourElement(i), HourTextColor(i, light)
				}
				r, g, b := int(col>>16&255), int(col>>8&255), int(col&255)
				ok := false
				switch el {
				case "木":
					ok = g > r && g > b
				case "火":
					ok = r > g && r > b
				case "土":
					ok = r > g && g > b
				case "金":
					ok = max(r, max(g, b))-min(r, min(g, b)) <= 36
				case "水":
					ok = b > r
				}
				if !ok {
					t.Errorf("%s %d light=%v: %s / %06x", kind, i, light, el, col)
				}
			}
		}
	}
}
func TestV3WeiHasEarthTextNotWoodGreen(t *testing.T) {
	if HourElement(7) != "土" {
		t.Fatal("未 must use Earth palette")
	}
	c := HourTextColor(7, false)
	if c>>16&255 < c>>8&255 {
		t.Fatalf("unexpected green text: %06x", c)
	}
}
func TestV3LayoutDayFirstEveryPair(t *testing.T) {
	d := DefaultData()
	count := 0
	for _, dpi := range []int{96, 120, 144, 192, 288} {
		for _, light := range []bool{false, true} {
			for _, day := range d.Days {
				for _, hour := range d.Hours {
					l := CompactLayout(day, hour, CardOptions{DPI: dpi, Light: light, HourDetails: true, MaxHeight: 900 * dpi / 96}, testMeasure(dpi))
					var dp, hp Rect
					for _, p := range l.Panels {
						if p.Kind == "day" {
							dp = p.Rect
						}
						if p.Kind == "hour" {
							hp = p.Rect
						}
					}
					if dp.Top >= hp.Top || dp.Bottom >= hp.Top {
						t.Fatalf("panels overlapped/order wrong %d/%d", day.Weekday, hour.Index)
					}
					if l.Blocks[0].Text != "曜星" {
						t.Fatal("unexpected card header")
					}
					for _, b := range l.Blocks {
						if b.Rect.Left < 0 || int(b.Rect.Right) > l.Width || b.Rect.Top < 0 || int(b.Rect.Bottom) > l.ContentHeight {
							t.Fatalf("block outside layout: %+v", b)
						}
						if got := testMeasure(dpi)(b.Text, b.Size, b.Bold); got > b.Rect.W()+1 {
							t.Fatalf("text overflow dpi=%d %q: %d > %d", dpi, b.Text, got, b.Rect.W())
						}
					}
					ds, hs := sectionString(l, "day"), sectionString(l, "hour")
					if !strings.Contains(ds, compact(day.Advice)) || !strings.Contains(ds, compact(day.China)) {
						t.Fatalf("day text missing %s", day.Name)
					}
					if !strings.Contains(hs, compact(hour.Advice)) || !strings.Contains(hs, compact(hour.Meaning)) {
						t.Fatalf("hour text missing %s", hour.Name)
					}
					if strings.Contains(hs, day.China) || strings.Contains(ds, hour.Meaning) {
						t.Fatal("weekday/hour explanations mixed")
					}
					count++
				}
			}
		}
	}
	t.Logf("%d layout combinations: 84 day/hour pairs × 2 themes × 5 DPI scales", count)
}
func TestV3LongDetailsScrollWithoutDeletingContent(t *testing.T) {
	d := DefaultData()
	day := d.DayAt(time.Date(2026, 9, 8, 13, 0, 0, 0, time.UTC))
	h := d.HourByIndex(7)
	day.Advice = strings.Repeat("这是一段加长的用户工作安排，", 20)
	l := CompactLayout(day, h, CardOptions{DPI: 96, MaxHeight: 280, DayDetails: true, HourDetails: true}, testMeasure(96))
	if l.Height != 280 || l.ContentHeight <= l.Height {
		t.Fatal("long card not scrollable")
	}
	for _, s := range []string{day.Advice, day.Sutra, day.West} {
		if !strings.Contains(sectionString(l, "day"), compact(s)) {
			t.Fatal("weekday text lost")
		}
	}
	if !strings.Contains(sectionString(l, "hour"), compact(h.Advice)) {
		t.Fatal("hour text lost")
	}
}
func TestV3CompactCardHasNoFooterToolbar(t *testing.T) {
	d := DefaultData()
	l := CompactLayout(d.Days[0], d.Hours[0], CardOptions{DPI: 96, HourDetails: true}, testMeasure(96))
	if len(l.Buttons) != 2 {
		t.Fatal("only per-section controls expected")
	}
	for _, b := range l.Buttons {
		if b.ID != 5 && b.ID != 6 {
			t.Fatal("footer action returned")
		}
	}
	for _, b := range l.Blocks {
		for _, x := range []string{"免责声明", "操作提示", "七曜时辰", "复制当前", "全部安排"} {
			if strings.Contains(b.Text, x) {
				t.Fatal("persistent boilerplate", x)
			}
		}
	}
}
func TestV3IndependentSectionDetails(t *testing.T) {
	d := DefaultData()
	day, h := d.Days[0], d.Hours[7]
	for _, dd := range []bool{false, true} {
		for _, hd := range []bool{false, true} {
			l := CompactLayout(day, h, CardOptions{DPI: 96, DayDetails: dd, HourDetails: hd}, testMeasure(96))
			ds, hs := sectionString(l, "day"), sectionString(l, "hour")
			if strings.Contains(ds, compact(day.Sutra)) != dd || strings.Contains(hs, compact(h.Meaning)) != hd {
				t.Fatal("details controls coupled")
			}
		}
	}
}
func TestV3EachSkinUsesItsOwnIndex(t *testing.T) {
	d := DefaultData()
	for _, day := range d.Days {
		for _, h := range d.Hours {
			l := CompactLayout(day, h, CardOptions{DPI: 96}, testMeasure(96))
			for _, p := range l.Panels {
				if p.Kind == "day" && p.Index != day.Weekday {
					t.Fatal("weekday skin doesn't track day")
				}
				if p.Kind == "hour" && p.Index != h.Index {
					t.Fatal("hour skin doesn't track hour")
				}
			}
		}
	}
}
func TestV3NativeTransparencyPath(t *testing.T) {
	b, e := os.ReadFile("ui_windows.go")
	if e != nil {
		t.Fatal(e)
	}
	s := strings.Split(string(b), "func (a *App) makeLayout")[0]
	for _, bad := range []string{"RoundRect(", "canvasPanel(", "artPanel(", "rectFill("} {
		if strings.Contains(s, bad) {
			t.Fatal("widget has a plate", bad)
		}
	}
	if !strings.Contains(s, "HourTextColor(hour.Index") || !strings.Contains(s, "DayTextColor(day.Weekday") {
		t.Fatal("font color not element-aware")
	}
	b, e = os.ReadFile("surface_windows.go")
	if e != nil {
		t.Fatal(e)
	}
	if !bytes.Contains(b, []byte("updateLayeredWindow.Call")) {
		t.Fatal("no native alpha presenter")
	}
}
func TestV3PremultipliedAlphaMath(t *testing.T) {
	b := NewPixelBuffer(3, 3)
	b.solid(1, 1, 0xFE8040, 128)
	i := (1*3 + 1) * 4
	if b.Pix[i+3] != 128 || b.Pix[i+2] != 127 || b.Pix[i+1] != 64 || b.Pix[i] != 32 {
		t.Fatalf("premult error %v", b.Pix[i:i+4])
	}
	b.solid(1, 1, 0x50C0F0, 128)
	if b.Pix[i+3] != 192 {
		t.Fatalf("source-over alpha %d", b.Pix[i+3])
	}
	for j := 0; j < len(b.Pix); j += 4 {
		for k := 0; k < 3; k++ {
			if b.Pix[j+k] > b.Pix[j+3] {
				t.Fatal("not premultiplied")
			}
		}
	}
}
func TestV3RoundCornersAndClip(t *testing.T) {
	b := NewPixelBuffer(40, 40)
	b.RoundRect(R(0, 0, 40, 40), 12, 0x324D5B, b.Bounds())
	if b.Pix[3] != 0 {
		t.Fatal("corner is not transparent")
	}
	if b.Pix[(20*40+20)*4+3] != 255 {
		t.Fatal("center isn't opaque")
	}
	src := NewPixelBuffer(8, 8)
	src.RoundRect(src.Bounds(), 0, 0xffffff, src.Bounds())
	clean := NewPixelBuffer(20, 20)
	clean.Blit(src, 0, 0, R(4, 4, 4, 4), nil, 0)
	for y := 0; y < 20; y++ {
		for x := 0; x < 20; x++ {
			a := clean.Pix[(y*20+x)*4+3]
			if (x < 4 || x >= 8 || y < 4 || y >= 8) && a != 0 {
				t.Fatal("clip leak")
			}
		}
	}
}
func TestV3AssetHeaderIcons(t *testing.T) {
	for _, kind := range []string{"day", "hour"} {
		for _, mode := range []string{"dark", "light"} {
			p := "assets/icons/header-" + kind + "-" + mode + ".png"
			if _, e := artwork.ReadFile(p); e != nil {
				t.Fatal(e)
			}
		}
	}
	if visualDesign.Version != 3 {
		t.Fatalf("artwork manifest still v%d", visualDesign.Version)
	}
}
func TestV3DiagnosticLayoutExport(t *testing.T) {
	dir := os.Getenv("ASTRAL_AUDIT_DIR")
	if dir == "" {
		t.Skip("optional visual-review export")
	}
	if e := os.MkdirAll(dir, 0755); e != nil {
		t.Fatal(e)
	}
	d := DefaultData()
	var records []any
	for _, light := range []bool{false, true} {
		for _, wd := range []int{2, 3, 4} {
			day := Day{}
			for _, v := range d.Days {
				if v.Weekday == wd {
					day = v
				}
			}
			h := d.HourByIndex(7)
			l := CompactLayout(day, h, CardOptions{DPI: 144, Light: light, HourDetails: true}, testMeasure(144))
			records = append(records, map[string]any{"name": fmt.Sprintf("day%d-light%v", wd, light), "layout": l, "light": light})
		}
	}
	b, _ := json.MarshalIndent(records, "", "  ")
	if e := os.WriteFile(filepath.Join(dir, "layouts.json"), b, 0644); e != nil {
		t.Fatal(e)
	}
}

func TestV3SettingsMigrationPreservesUserPreferences(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("APPDATA", dir)
	path := settingsPath()
	if e := os.MkdirAll(filepath.Dir(path), 0700); e != nil {
		t.Fatal(e)
	}
	old := []byte(`{"gap":183,"font_size":14,"theme":"light","hour_details":false,"day_details":true,"autostart":true}`)
	if e := os.WriteFile(path, old, 0600); e != nil {
		t.Fatal(e)
	}
	s := ReadSettings()
	if s.Gap != 183 || s.FontSize != 14 || s.Theme != "light" || !s.Autostart {
		t.Fatalf("user preference overwritten: %+v", s)
	}
	if !s.HourDetails || s.DayDetails || s.UIVersion != 3 {
		t.Fatal("first-open V3 compact state incorrect")
	}
	s.HourDetails = false
	s.DayDetails = true
	if e := SaveSettings(s); e != nil {
		t.Fatal(e)
	}
	got := ReadSettings()
	if got.HourDetails || !got.DayDetails {
		t.Fatal("V3 user's panel choices overwritten")
	}
}
