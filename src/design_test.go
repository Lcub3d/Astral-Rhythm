package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"image/png"
	"math"
	"os"
	"strings"
	"testing"
	"time"
)

func TestArtworkCompletenessAndRasterization(t *testing.T) {
	count := 0
	for _, light := range []bool{true, false} {
		for _, kind := range []string{"day", "hour"} {
			n := 7
			if kind == "hour" {
				n = 12
			}
			for i := 0; i < n; i++ {
				for _, bg := range []bool{false, true} {
					name := assetName(kind, i, light, bg)
					b, e := artwork.ReadFile(name)
					if e != nil {
						t.Fatal(e)
					}
					im, e := png.Decode(bytes.NewReader(b))
					if e != nil {
						t.Fatal(name, e)
					}
					if im.Bounds().Dx() < 128 || im.Bounds().Dy() < 128 {
						t.Fatal("undersized artwork", name)
					}
					count++
				}
			}
		}
	}
	if count != 76 {
		t.Fatal(count)
	}
	b, e := artwork.ReadFile("assets/icons/app.png")
	if e != nil {
		t.Fatal(e)
	}
	if _, e = png.Decode(bytes.NewReader(b)); e != nil {
		t.Fatal(e)
	}
	t.Log("76 day/hour PNG assets plus the app icon decode correctly; both light/dark variants included")
}
func TestAllIdentitiesHaveDistinctArtwork(t *testing.T) {
	for _, light := range []bool{true, false} {
		for _, bg := range []bool{false, true} {
			seen := map[[32]byte]string{}
			for _, kind := range []string{"day", "hour"} {
				n := 7
				if kind == "hour" {
					n = 12
				}
				for i := 0; i < n; i++ {
					name := assetName(kind, i, light, bg)
					b, e := artwork.ReadFile(name)
					if e != nil {
						t.Fatal(e)
					}
					hash := sha256.Sum256(b)
					if old, ok := seen[hash]; ok {
						t.Fatalf("duplicate: %s / %s", old, name)
					}
					seen[hash] = name
				}
			}
			if len(seen) != 19 {
				t.Fatal("missing identities")
			}
		}
	}
}
func luminance(c uint32) float64 {
	channel := func(shift uint) float64 {
		v := float64((c>>shift)&255) / 255
		if v <= .04045 {
			return v / 12.92
		}
		return math.Pow((v+.055)/1.055, 2.4)
	}
	return .2126*channel(16) + .7152*channel(8) + .0722*channel(0)
}
func TestAccentTextContrast(t *testing.T) {
	all := append(append([]VisualIdentity{}, visualDesign.Days...), visualDesign.Hours...)
	for _, v := range all {
		for _, light := range []bool{true, false} {
			f, b := luminance(v.Accent(light)), luminance(v.Surface(light))
			if f < b {
				f, b = b, f
			}
			ratio := (f + .05) / (b + .05)
			if ratio < 4.5 {
				t.Errorf("%s light=%v: %.2f:1", v.Name, light, ratio)
			}
		}
	}
}
func TestAll84PairsKeepExplanationsSeparate(t *testing.T) {
	d := DefaultData()
	for _, day := range d.Days {
		for _, hour := range d.Hours {
			hs, ds := HourExplanations(hour), DayExplanations(day)
			if len(hs) != 2 || hs[0].Value != hour.Stems || hs[1].Value != hour.Meaning {
				t.Fatal("hour field mix-up")
			}
			if len(ds) != 3 || ds[0].Value != day.China || ds[1].Value != day.Sutra || ds[2].Value != day.West {
				t.Fatal("weekday field mix-up")
			}
		}
	}
}
func TestHourProgressForEntireDay(t *testing.T) {
	for m := 0; m < 1440; m++ {
		now := time.Date(2026, 9, 9, m/60, m%60, 0, 0, time.UTC)
		p := HourProgress(now)
		want := float64((m+60)%120) / 120
		if p < 0 || p >= 1 || math.Abs(p-want) > 1e-9 {
			t.Fatalf("%v: got %v, want %v", now, p, want)
		}
	}
}
func TestIndependentExplanationSettingsAndLegacyMigration(t *testing.T) {
	s := DefaultSettings()
	if e := json.Unmarshal([]byte(`{"gap":163,"font_size":14,"theme":"dark","traditional":true}`), &s); e != nil {
		t.Fatal(e)
	}
	s.Normalize()
	if s.Gap != 163 || s.FontSize != 14 || s.Theme != "dark" || !s.HourDetails || s.DayDetails {
		t.Fatalf("bad v1 migration: %+v", s)
	}
	for _, h := range []bool{false, true} {
		for _, d := range []bool{false, true} {
			s.HourDetails = h
			s.DayDetails = d
			b, _ := json.Marshal(s)
			var got Settings
			if e := json.Unmarshal(b, &got); e != nil {
				t.Fatal(e)
			}
			got.Normalize()
			if got.HourDetails != h || got.DayDetails != d {
				t.Fatal("explanations are coupled")
			}
		}
	}
}
func TestCopiedSectionsDoNotInterleave(t *testing.T) {
	d := DefaultData()
	now := time.Date(2026, 9, 9, 13, 40, 0, 0, time.UTC)
	text := CurrentText(d, now)
	parts := strings.Split(text, "【此刻 · 时辰】")
	if len(parts) != 2 || !strings.Contains(parts[0], "【今日 · 曜日】") {
		t.Fatal("missing day-first boundary")
	}
	if strings.Contains(parts[0], "地支藏干") || strings.Contains(parts[1], d.DayAt(now).China) || strings.Contains(parts[1], d.DayAt(now).Sutra) {
		t.Fatal("day/hour explanations mixed")
	}
}
func TestNoPersistentFooterBoilerplate(t *testing.T) {
	b, e := os.ReadFile("ui_windows.go")
	if e != nil {
		t.Fatal(e)
	}
	for _, needle := range []string{"传统象征不代表固定生理规律", "悬停查看 · 单击固定", "按本地时间显示"} {
		if strings.Contains(string(b), needle) {
			t.Fatal("footer returned:", needle)
		}
	}
	for _, note := range DefaultData().Notes {
		if strings.Contains(AllText(DefaultData()), note) {
			t.Fatal("unrequested explanatory footer in all-arrangements export")
		}
	}
}
