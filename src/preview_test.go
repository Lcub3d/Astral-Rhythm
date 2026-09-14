package main

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"
)

func previewOptions(data Data, now time.Time, lang string, light bool, dpi, width int, state PreviewState) CardOptions {
	l := NewLocalizer(lang)
	return CardOptions{Locale: l, DPI: dpi, MaxWidth: width * dpi / 96, MaxHeight: 480 * dpi / 96, Light: light, DayDetails: true, HourDetails: true,
		Preview: state, UpcomingDays: UpcomingDays(data, now, l, light), UpcomingHours: UpcomingHours(data, now, l, light)}
}
func TestPreviewIndependentSession(t *testing.T) {
	var p PreviewState
	if p.Active() {
		t.Fatal("not collapsed initially")
	}
	if p.Toggle(101) || p.Toggle(5) {
		t.Fatal("closed row/legacy control accepted")
	}
	p.Toggle(previewDaysButton)
	p.Toggle(102)
	if !p.Days || p.Hours || !p.Open(102) {
		t.Fatal(p)
	}
	p.Toggle(previewHoursButton)
	p.Toggle(204)
	if !p.Open(102) || !p.Open(204) {
		t.Fatal("selection coupled")
	}
	p.Toggle(previewDaysButton)
	if p.Days || p.DayRow != 0 || !p.Hours || p.HourRow != 4 {
		t.Fatal("collapse lost other drawer")
	}
	p.Toggle(204)
	if p.HourRow != 0 {
		t.Fatal("row won't collapse")
	}
	p.Reset()
	if p != (PreviewState{}) {
		t.Fatal("session leaked")
	}
	b, _ := json.Marshal(DefaultSettings())
	if strings.Contains(strings.ToLower(string(b)), "preview") {
		t.Fatal("preview persisted")
	}
}
func TestPreviewFutureDaysCalendar(t *testing.T) {
	d := DefaultData()
	l := NewLocalizer("zh-CN")
	for _, date := range []string{"2026-09-13", "2026-09-14", "2026-12-31", "2028-02-28", "2026-02-28"} {
		now, _ := time.Parse("2006-01-02", date)
		now = now.Add(23*time.Hour + 59*time.Minute)
		items := UpcomingDays(d, now, l, false)
		if len(items) != 3 {
			t.Fatal("wrong horizon")
		}
		for i, v := range items {
			want := now.AddDate(0, 0, i+1)
			if v.Start.Format("2006-01-02") != want.Format("2006-01-02") || v.Index != int(want.Weekday()) || !v.Start.After(now) {
				t.Fatal(date, v)
			}
			if v.Name != d.DayAt(want).Short() || !strings.Contains(v.When, previewDate(want)) {
				t.Fatal("identity/date changed")
			}
			if v.Color != DayTextColor(v.Index, false) {
				t.Fatal("wrong day palette")
			}
		}
	}
}
func TestPreviewAllMinutesAndMidnight(t *testing.T) {
	d := DefaultData()
	l := NewLocalizer("zh-CN")
	n := 0
	for min := 0; min < 1440; min++ {
		now := time.Date(2026, 12, 31, min/60, min%60, 0, 0, time.UTC)
		rows := UpcomingHours(d, now, l, false)
		if len(rows) != 4 {
			t.Fatal("wrong hour horizon")
		}
		for i, row := range rows {
			if row.Index != (HourIndex(now)+i+1)%12 || !row.Start.After(now) || row.End.Sub(row.Start) != 2*time.Hour || row.Start.Minute() != 0 || row.Start.Hour()%2 != 1 {
				t.Fatalf("minute %d: %+v", min, row)
			}
			if i > 0 && !row.Start.Equal(rows[i-1].End) {
				t.Fatal("gap or duplicate")
			}
			if row.Color != HourTextColor(row.Index, false) || row.Name != hourLabels[row.Index] {
				t.Fatal("branch/name/color mismatch")
			}
			if row.Start.Day() != now.Day() && !strings.Contains(row.When, "明日") {
				t.Fatal("missing day boundary")
			}
			n++
		}
	}
	t.Logf("%d future slots checked across all 1,440 minute positions", n)
}
func TestPreviewBoundarySeconds(t *testing.T) {
	d := DefaultData()
	l := NewLocalizer("zh-CN")
	for _, x := range []struct{ hour, min, sec, want int }{{10, 59, 59, 6}, {11, 0, 0, 7}, {22, 59, 59, 0}, {23, 0, 0, 1}, {23, 59, 59, 1}, {0, 0, 0, 1}, {0, 59, 59, 1}, {1, 0, 0, 2}} {
		now := time.Date(2026, 9, 14, x.hour, x.min, x.sec, 0, time.UTC)
		if got := UpcomingHours(d, now, l, false)[0].Index; got != x.want {
			t.Fatalf("%+v got %d", x, got)
		}
	}
	now := time.Date(2026, 9, 14, 22, 30, 0, 0, time.UTC)
	if !strings.Contains(UpcomingHours(d, now, l, false)[0].When, "次日01:00") {
		t.Fatal("Zi end day ambiguous")
	}
	if !strings.Contains(UpcomingHours(d, now, NewLocalizer("en"), false)[0].When, "(+1d)") {
		t.Fatal("English midnight ambiguity")
	}
}
func TestPreviewCivilTimezoneNoDrift(t *testing.T) {
	d := DefaultData()
	l := NewLocalizer("en")
	for _, off := range []int{-12 * 3600, 0, 8 * 3600, 14 * 3600} {
		// Fixed zone plus the civil contract prevents timezone offsets from
		// accidentally converting today's date to UTC and changing the day star.
		now := time.Date(2026, 3, 8, 0, 30, 0, 0, time.FixedZone("test", off))
		h := UpcomingHours(d, now, l, false)
		days := UpcomingDays(d, now, l, false)
		if h[0].Index != 1 || h[0].Start.Hour() != 1 || days[0].Start.Day() != 9 {
			t.Fatal("UTC conversion drift", off)
		}
	}
}
func TestPreviewSleepLunchAndOriginalPriorities(t *testing.T) {
	d := DefaultData()
	l := NewLocalizer("zh-CN")
	for _, h := range d.Hours {
		tag := previewKeyword("hour", h.Index, h.Advice, l)
		if h.Index <= 2 && !strings.Contains(tag, "睡眠") {
			t.Fatal("night incorrectly encourages work", tag)
		}
		if h.Index == 6 && (tag != "午餐休息" || h.Priority != "") {
			t.Fatal("invented lunch work/priority")
		}
	}
	rows := UpcomingHours(d, time.Date(2026, 9, 14, 9, 30, 0, 0, time.UTC), l, false)
	for i, want := range []string{"", "P2 · P3", "P1", "P3"} {
		if rows[i].Priority != want {
			t.Fatal(rows[i])
		}
	}
}
func TestPreviewEditedAdviceNeverUsesStaleKeyword(t *testing.T) {
	d := DefaultData()
	raw := d.HourByIndex(6)
	raw.Advice = "这是我刚改的安排；不要套用默认建议"
	l := NewLocalizer("en")
	if got := previewKeyword("hour", 6, raw.Advice, l); got != "这是我刚改的安排" {
		t.Fatal("custom content lost", got)
	}
	l.Catalog[raw.Advice] = "My edited arrangement; no default advice"
	if got := previewKeyword("hour", 6, raw.Advice, l); got != "My edited arrangement" {
		t.Fatal("translation ignored", got)
	}
	now := time.Date(2026, 9, 14, 9, 30, 0, 0, time.UTC)
	for i := range d.Hours {
		if d.Hours[i].Index == 6 {
			d.Hours[i] = raw
		}
	}
	p := UpcomingHours(d, now, l, false)[0]
	if p.Keyword != "My edited arrangement" || p.Advice != l.Content(raw.Advice) {
		t.Fatal(p)
	}
	day := d.DayAt(now)
	day.Advice = "客户会议，优先"
	if previewKeyword("day", day.Weekday, day.Advice, l) != day.Advice {
		t.Fatal("stale day summary")
	}
}
func TestPreviewDefaultDoesNotLeakFutureRows(t *testing.T) {
	d := DefaultData()
	now := time.Date(2026, 9, 14, 9, 30, 0, 0, time.UTC)
	opt := previewOptions(d, now, "zh-CN", false, 96, 320, PreviewState{})
	layout := CompactLayout(d.DayAt(now), d.HourAt(now), opt, testMeasure(96))
	if len(layout.Buttons) != 4 {
		t.Fatal("expected two detail and two drawer controls")
	}
	for _, p := range layout.Panels {
		if strings.HasSuffix(p.Kind, "-row") {
			t.Fatal("preview not closed")
		}
	}
	for _, b := range layout.Buttons {
		if b.ID > previewDayRowBase {
			t.Fatal("future row still clickable")
		}
	}
	text := ""
	for _, b := range layout.Blocks {
		text += b.Text
	}
	for _, row := range append(opt.UpcomingDays, opt.UpcomingHours...) {
		if strings.Contains(text, row.Keyword) {
			t.Fatal("future leaked", row.Keyword)
		}
	}
}
func TestPreviewLayoutsAcrossLanguagesThemesAndDPI(t *testing.T) {
	d := DefaultData()
	count := 0
	states := []PreviewState{{}, {Days: true}, {Hours: true}, {Days: true, Hours: true}, {Days: true, Hours: true, DayRow: 2, HourRow: 3}}
	for _, lang := range []string{"zh-CN", "en"} {
		for _, light := range []bool{false, true} {
			for _, dpi := range []int{96, 120, 144, 192, 288} {
				for _, width := range []int{240, 320} {
					m := testMeasure(dpi)
					for day := 0; day < 7; day++ {
						for hour := 0; hour < 12; hour++ {
							for _, state := range states {
								now := time.Date(2026, 9, 6+day, hour*2, 30, 0, 0, time.UTC)
								opt := previewOptions(d, now, lang, light, dpi, width, state)
								layout := CompactLayout(d.DayAt(now), d.HourAt(now), opt, m)
								for _, b := range layout.Blocks {
									if b.Rect.Left < 0 || int(b.Rect.Right) > layout.Width || b.Rect.Top < 0 || int(b.Rect.Bottom) > layout.ContentHeight {
										t.Fatalf("out of bounds: %+v", b)
									}
									if got := m(b.Text, b.Size, b.Bold); got > b.Rect.W()+1 {
										t.Fatalf("%s dpi%d width%d %q: %d > %d", lang, dpi, width, b.Text, got, b.Rect.W())
									}
									if lang == "en" && hasHan(b.Text) {
										t.Fatal("untranslated", b.Text)
									}
								}
								var dayRect, hourRect Rect
								mainPanels := 0
								for _, p := range layout.Panels {
									if p.Kind == "day" {
										dayRect = p.Rect
										mainPanels++
									}
									if p.Kind == "hour" {
										hourRect = p.Rect
										mainPanels++
									}
								}
								if mainPanels != 2 || dayRect.Bottom >= hourRect.Top {
									t.Fatal("main cards mixed")
								}
								for _, b := range layout.Buttons {
									parent := dayRect
									if b.ID == 5 || b.ID == previewHoursButton || b.ID > previewHourRowBase {
										parent = hourRect
									}
									if !parent.Contains(Point{b.Rect.Left, b.Rect.Top}) || b.Rect.Bottom > parent.Bottom {
										t.Fatal("drawer escaped parent")
									}
								}
								if !strings.Contains(sectionString(layout, "day"), compact(opt.Locale.Content(d.DayAt(now).Advice))) || !strings.Contains(sectionString(layout, "hour"), compact(opt.Locale.Content(d.HourAt(now).Advice))) {
									t.Fatal("current advice replaced")
								}
								count++
							}
						}
					}
				}
			}
		}
	}
	t.Logf("%d preview layouts checked (portable estimated text widths)", count)
}
func TestPreviewRowsExpandOriginalAdviceAndScroll(t *testing.T) {
	d := DefaultData()
	now := time.Date(2026, 9, 14, 9, 30, 0, 0, time.UTC)
	opt := previewOptions(d, now, "en", false, 96, 240, PreviewState{Days: true, Hours: true, DayRow: 1, HourRow: 1})
	opt.UpcomingHours[0].Advice = strings.Repeat("An edited long source cell is preserved in full. ", 40)
	original := reflect.ValueOf(d).Interface()
	layout := CompactLayout(d.DayAt(now), d.HourAt(now), opt, testMeasure(96))
	if layout.ContentHeight <= layout.Height || !strings.Contains(sectionString(layout, "hour"), compact(opt.UpcomingHours[0].Advice)) {
		t.Fatal("long future advice lost")
	}
	if !reflect.DeepEqual(d, original) {
		t.Fatal("view mutated data")
	}
}
func TestPreviewEllipsisKeepsUnicodeAndBounds(t *testing.T) {
	m := testMeasure(96)
	for _, text := range []string{"午餐休息", "Routine & connect", strings.Repeat("中文English", 10000)} {
		for _, width := range []int{0, 3, 10, 30, 100} {
			got := fitPreviewText(text, width, 11, false, m)
			if m(got, 11, false) > width {
				t.Fatal("ellipsis overflow")
			}
		}
	}
}
