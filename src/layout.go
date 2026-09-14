package main

import (
	"strings"
	"unicode"
)

// All positions are physical pixels. DPI is applied once, at layout time.
// This file is shared by the native application and all cross-platform tests.
type VisualPanel struct {
	Rect  Rect
	Kind  string
	Index int
}
type VisualIcon struct {
	Rect Rect
	Path string
}
type CardOptions struct {
	Preview                        PreviewState
	UpcomingDays, UpcomingHours    []PreviewItem
	Locale                         Localizer
	DPI, MaxWidth, MaxHeight       int
	Light, DayDetails, HourDetails bool
}
type CardLayout struct {
	Width, Height, ContentHeight int
	Body                         Rect
	Blocks                       []TextBlock
	Panels                       []VisualPanel
	Icons                        []VisualIcon
	Buttons                      []Button
}
type MeasureText func(text string, size int, bold bool) int

func WrapText(s string, width, size int, bold bool, measure MeasureText) []string {
	s = strings.ReplaceAll(strings.ReplaceAll(s, "\r\n", "\n"), "\r", "\n")
	var lines []string
	buf := ""
	for _, ch := range s {
		if ch == '\n' {
			lines = append(lines, buf)
			buf = ""
			continue
		}
		candidate := buf + string(ch)
		if buf != "" && measure(candidate, size, bold) > width {
			// Wrap Latin text at word boundaries; keep source words intact when possible.
			if ch == ' ' || ch == '\t' {
				lines = append(lines, strings.TrimRightFunc(buf, unicode.IsSpace))
				buf = ""
				continue
			}
			if n := strings.LastIndexAny(buf, " \t"); n > 0 {
				head, tail := strings.TrimRightFunc(buf[:n], unicode.IsSpace), strings.TrimLeftFunc(buf[n+1:], unicode.IsSpace)
				lines = append(lines, head)
				buf = tail + string(ch)
				continue
			}
			// Avoid leaving a lone separator at the beginning of a CJK line.
			if strings.ContainsRune("，。、；：！？）》】", ch) {
				r := []rune(buf)
				if len(r) > 1 {
					lines = append(lines, string(r[:len(r)-1]))
					buf = string(r[len(r)-1:]) + string(ch)
					continue
				}
			}
			lines = append(lines, buf)
			buf = ""
		}
		buf += string(ch)
	}
	if buf != "" {
		lines = append(lines, buf)
	}
	if len(lines) == 0 {
		lines = []string{""}
	}
	return lines
}

func CompactLayout(day Day, hour Hour, opt CardOptions, measure MeasureText) CardLayout {
	dpi := opt.DPI
	if dpi < 1 {
		dpi = 96
	}
	s := func(n int) int { return (n*dpi + 48) / 96 }
	l := opt.Locale
	rawDay, rawHour := day, hour
	day, hour = l.Day(day), l.Hour(hour)
	w := s(320)
	if opt.MaxWidth > 0 {
		w = min(w, opt.MaxWidth)
	}
	w = max(w, s(240))
	out := CardLayout{Width: w}
	c := darkPalette
	if opt.Light {
		c = lightPalette
	}
	x0, inner := s(10), s(14)
	x := x0 + inner
	cw := w - 2*x0
	tw := cw - 2*inner
	labelW := s(42)
	if l.English() {
		labelW = s(64)
	}
	y := s(10)
	line := func(text string, x, y, width, height, size int, bold bool, col uint32) {
		out.Blocks = append(out.Blocks, TextBlock{text, R(x, y, width, height), size, bold, col})
	}
	row := func(label, value string, yy int, labelColor uint32) int {
		if value == "" {
			return yy
		}
		line(label, x, yy, labelW, s(20), 12, false, labelColor)
		for _, str := range WrapText(value, tw-labelW, 12, false, measure) {
			line(str, x+labelW, yy, tw-labelW, s(20), 12, false, c.Text)
			yy += s(20)
		}
		return yy + s(2)
	}
	dc, hc := DayTextColor(day.Weekday, opt.Light), HourTextColor(hour.Index, opt.Light)
	mode := "dark"
	if opt.Light {
		mode = "light"
	}
	dy := y
	out.Panels = append(out.Panels, VisualPanel{R(x0, y, cw, 0), "day", day.Weekday})
	dpi0 := len(out.Panels) - 1
	out.Icons = append(out.Icons, VisualIcon{R(x, y+s(13), s(24), s(24)), "assets/icons/header-day-" + mode + ".png"})
	line(l.Text("曜星", "Day Star"), x+s(35), y+s(16), tw-s(70), s(23), 15, false, c.Text)
	out.Buttons = append(out.Buttons, Button{true, 6, l.Text("曜星更多", "Day-star details"), R(w-x-s(20), y+s(12), s(22), s(28))})
	y += s(54)
	out.Icons = append(out.Icons, VisualIcon{R(x, y, s(23), s(23)), assetName("day", day.Weekday, opt.Light, false)})
	name := l.DayShort(rawDay)
	nameW := max(s(47), measure(name, 16, true)+s(3))
	line(name, x+s(35), y+s(1), nameW, s(25), 16, true, c.Text)
	badge := l.Text("五行·"+DayElement(day.Weekday), l.Content(DayElement(day.Weekday)))
	badgeW := max(s(34), measure(badge, 11, false)+s(14))
	bx := x + s(35) + nameW + s(6)
	out.Panels = append(out.Panels, VisualPanel{R(bx, y+s(2), badgeW, s(22)), "day-badge", day.Weekday})
	line(badge, bx+s(7), y+s(5), badgeW-s(10), s(18), 11, false, dc)
	weekRole := day.WeekLabel + " · " + day.Role
	remaining := x + tw - (bx + badgeW + s(11))
	if !l.English() && measure(weekRole, 11, false) <= remaining {
		line(weekRole, bx+badgeW+s(11), y+s(5), remaining, s(18), 11, false, c.Muted)
	}
	y += s(35)
	if l.English() {
		y = row("Day", weekRole, y, c.Muted)
	}
	y = row(l.Text("象征", "Symbol"), day.China, y, blendRGB(dc, c.Muted, .45))
	y = row(l.Text("宜", "Best for"), day.Advice, y, blendRGB(dc, c.Muted, .25))
	if opt.DayDetails {
		y += s(4)
		y = row(l.Text("星名", "Name"), day.Name+" · "+day.WeekLabel+" · "+day.Role, y, dc)
		y = row(l.Text("宿曜", "Sutra"), day.Sutra, y, dc)
		y = row(l.Text("西方", "Western"), day.West, y, dc)
	}
	y = appendPreviewDrawer(&out, "day", opt.UpcomingDays, opt, x, y, tw, measure)
	y += s(10)
	out.Panels[dpi0].Rect = R(x0, dy, cw, y-dy)
	y += s(10)
	hy := y
	out.Panels = append(out.Panels, VisualPanel{R(x0, y, cw, 0), "hour", hour.Index})
	hpi := len(out.Panels) - 1
	out.Icons = append(out.Icons, VisualIcon{R(x, y+s(13), s(24), s(24)), "assets/icons/header-hour-" + mode + ".png"})
	line(l.Text("时辰", "Shichen"), x+s(35), y+s(16), tw-s(60), s(23), 15, false, c.Text)
	out.Buttons = append(out.Buttons, Button{true, 5, l.Text("时辰详情", "Shichen details"), R(w-x-s(20), y+s(12), s(22), s(28))})
	y += s(54)
	out.Icons = append(out.Icons, VisualIcon{R(x, y, s(23), s(23)), assetName("hour", hour.Index, opt.Light, false)})
	hnw := max(s(46), measure(hour.Name, 16, true)+s(3))
	line(hour.Name, x+s(35), y+s(1), hnw, s(25), 16, true, c.Text)
	px := x + s(35) + hnw + s(10)
	if measure(hour.Period, 13, false) > x+tw-px {
		y += s(27)
		px = x + s(35)
	}
	line(hour.Period, px, y+s(4), x+tw-px, s(22), 13, false, c.Text)
	y += s(35)
	labelColor := uint32(0x9ACBBC)
	if opt.Light {
		labelColor = 0x467366
	}
	y = row(l.Text("地支", "Branch"), l.Content(BranchLabel(rawHour)), y, labelColor)
	if opt.HourDetails {
		y = row(l.Text("藏干", "Stems"), strings.ReplaceAll(hour.Stems, "、", " · "), y, labelColor)
		y = row(l.Text("含义", "Meaning"), hour.Meaning, y, labelColor)
	}
	y = row(l.Text("宜", "Best for"), hour.Advice, y, labelColor)
	if hour.Priority != "" {
		y = row(l.Text("等级", "Priority"), strings.ReplaceAll(strings.ToUpper(hour.Priority), "、", " · "), y, hc)
	}
	y = appendPreviewDrawer(&out, "hour", opt.UpcomingHours, opt, x, y, tw, measure)
	y += s(10)
	out.Panels[hpi].Rect = R(x0, hy, cw, y-hy)
	y += s(10)
	out.ContentHeight = y
	out.Height = y
	if opt.MaxHeight > 0 {
		out.Height = min(y, opt.MaxHeight)
	}
	out.Height = max(out.Height, s(160))
	out.Body = R(0, 0, w, out.Height)
	return out
}
