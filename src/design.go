package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// PNGs are native runtime artwork. Matching SVG masters ship beside them for editing.
//
//go:embed assets/icons/*.png assets/backgrounds/*.png assets/design.json
var artwork embed.FS

type VisualIdentity struct {
	Name      string `json:"name"`
	Title     string `json:"title"`
	Light     string `json:"light"`
	Dark      string `json:"dark"`
	SoftLight string `json:"soft_light"`
	SoftDark  string `json:"soft_dark"`
}
type DesignSystem struct {
	Name    string           `json:"name"`
	Version int              `json:"version"`
	Days    []VisualIdentity `json:"days"`
	Hours   []VisualIdentity `json:"hours"`
}

func ReadDesign() DesignSystem {
	b, e := artwork.ReadFile("assets/design.json")
	if e != nil {
		panic(e)
	}
	var d DesignSystem
	if e = json.Unmarshal(b, &d); e != nil {
		panic(e)
	}
	if len(d.Days) != 7 || len(d.Hours) != 12 {
		panic("incomplete artwork manifest")
	}
	return d
}

var visualDesign = ReadDesign()

func hexColor(s string) uint32 {
	v, e := strconv.ParseUint(strings.TrimPrefix(s, "#"), 16, 32)
	if e != nil {
		return 0x698763
	}
	return uint32(v)
}
func (v VisualIdentity) Accent(light bool) uint32 {
	if light {
		return hexColor(v.Light)
	}
	return hexColor(v.Dark)
}
func (v VisualIdentity) Surface(light bool) uint32 {
	if light {
		return hexColor(v.SoftLight)
	}
	return hexColor(v.SoftDark)
}
func assetName(kind string, index int, light bool, background bool) string {
	mode := "dark"
	if light {
		mode = "light"
	}
	folder := "icons"
	if background {
		folder = "backgrounds"
	}
	return fmt.Sprintf("assets/%s/%s-%d-%s.png", folder, kind, index, mode)
}
func HourProgress(t time.Time) float64 {
	elapsed := float64(((t.Hour()+1)%24)%2*3600 + t.Minute()*60 + t.Second())
	return elapsed / 7200
}

// These sections prevent weekday symbolism from entering the hour explanation.
type Explanation struct{ Label, Value string }

func HourExplanations(h Hour) []Explanation {
	return []Explanation{{"地支藏干", h.Stems}, {"时辰含义", h.Meaning}}
}
func DayExplanations(d Day) []Explanation {
	return []Explanation{{"中国传统 · 五行象征", d.China}, {"《宿曜经》倾向", d.Sutra}, {"西方象征", d.West}}
}
