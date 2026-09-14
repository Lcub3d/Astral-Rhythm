package main

import (
	"fmt"
	"strings"
	"time"
)

const (
	previewDaysButton  = 20
	previewHoursButton = 21
	previewDayRowBase  = 100
	previewHourRowBase = 200
)

// PreviewState is a viewing session, never a persisted preference. Closing the
// popup returns to the present; the independently saved detail switches stay put.
type PreviewState struct {
	Days, Hours     bool
	DayRow, HourRow int // one-based; zero means no expanded advice
}

func (p *PreviewState) Reset()      { *p = PreviewState{} }
func (p PreviewState) Active() bool { return p.Days || p.Hours }
func (p PreviewState) Open(id int) bool {
	switch {
	case id == previewDaysButton:
		return p.Days
	case id == previewHoursButton:
		return p.Hours
	case id > previewDayRowBase && id <= previewDayRowBase+3:
		return p.Days && p.DayRow == id-previewDayRowBase
	case id > previewHourRowBase && id <= previewHourRowBase+4:
		return p.Hours && p.HourRow == id-previewHourRowBase
	}
	return false
}
func (p *PreviewState) Toggle(id int) bool {
	switch {
	case id == previewDaysButton:
		p.Days = !p.Days
		p.DayRow = 0
	case id == previewHoursButton:
		p.Hours = !p.Hours
		p.HourRow = 0
	case id > previewDayRowBase && id <= previewDayRowBase+3 && p.Days:
		n := id - previewDayRowBase
		if p.DayRow == n {
			n = 0
		}
		p.DayRow = n
	case id > previewHourRowBase && id <= previewHourRowBase+4 && p.Hours:
		n := id - previewHourRowBase
		if p.HourRow == n {
			n = 0
		}
		p.HourRow = n
	default:
		return false
	}
	return true
}

type PreviewItem struct {
	Kind                                  string
	Index                                 int
	Start, End                            time.Time // civil calendar coordinates, not elapsed-time forecasts
	Name, When, Keyword, Advice, Priority string
	Color                                 uint32
}

// These are short editorial summaries of the unchanged bundled advice, not new
// auspiciousness or performance claims. An edited cell NEVER inherits a stale
// keyword: show a clause from the user's current advice instead.
var dayPreviewWords = [7][2]string{
	{"休息复盘", "Rest & reflect"}, {"规划启动", "Plan & start"},
	{"集中攻坚", "Hard tasks"}, {"学习研究", "Study & write"},
	{"战略协作", "Strategy"}, {"展示交流", "Present & connect"}, {"整理维护", "Organize"},
}
var hourPreviewWords = [12][2]string{
	{"安稳睡眠", "Sleep"}, {"深度睡眠", "Deep sleep"}, {"继续睡眠", "Keep sleeping"},
	{"晨起活动", "Morning walk"}, {"早餐规划", "Breakfast & plan"}, {"专注攻坚", "Deep work"},
	{"午餐休息", "Lunch & rest"}, {"日常协作", "Routine & connect"}, {"重点推进", "Second peak"},
	{"运动收工", "Move & wind down"}, {"陪伴放松", "Family & relax"}, {"收尾待眠", "Wind down"},
}
var previewBaseline = DefaultData()

func previewKeyword(kind string, index int, advice string, l Localizer) string {
	var original string
	var pair [2]string
	if kind == "day" && index >= 0 && index < len(dayPreviewWords) {
		for _, d := range previewBaseline.Days {
			if d.Weekday == index {
				original = d.Advice
				break
			}
		}
		pair = dayPreviewWords[index]
	} else if kind == "hour" && index >= 0 && index < len(hourPreviewWords) {
		original = previewBaseline.HourByIndex(index).Advice
		pair = hourPreviewWords[index]
	}
	if original != "" && advice == original {
		return l.Text(pair[0], pair[1])
	}
	current := strings.TrimSpace(l.Content(advice))
	if i := strings.IndexAny(current, "\n\r；;。:"); i > 0 {
		current = strings.TrimSpace(current[:i])
	}
	return current
}

// Match wallClockNow: preserve the displayed date/hour but discard timezone
// offsets. Civil two-hour slots must not skip, duplicate or drift at DST changes.
func previewCivil(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.UTC)
}
func previewDate(t time.Time) string { return fmt.Sprintf("%d/%d", t.Month(), t.Day()) }
func previewPriority(p string) string {
	p = strings.ToUpper(strings.TrimSpace(p))
	return strings.NewReplacer("、", " · ", ",", " · ").Replace(p)
}

func UpcomingDays(data Data, now time.Time, l Localizer, light bool) []PreviewItem {
	now = previewCivil(now)
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	out := make([]PreviewItem, 0, 3)
	for n := 1; n <= 3; n++ {
		start := midnight.AddDate(0, 0, n)
		d := data.DayAt(start)
		when := l.Content(d.WeekLabel) + " · " + previewDate(start)
		if n == 1 {
			when = l.Text("明日", "Tomorrow") + " · " + when
		}
		if n == 2 && !l.English() {
			when = "后天 · " + when
		}
		out = append(out, PreviewItem{Kind: "day", Index: d.Weekday, Start: start, End: start.AddDate(0, 0, 1), Name: l.DayShort(d), When: when,
			Keyword: previewKeyword("day", d.Weekday, d.Advice, l), Advice: l.Content(d.Advice), Color: DayTextColor(d.Weekday, light)})
	}
	return out
}
func UpcomingHours(data Data, now time.Time, l Localizer, light bool) []PreviewItem {
	now = previewCivil(now)
	start := NextBoundary(now)
	out := make([]PreviewItem, 0, 4)
	for n := 0; n < 4; n++ {
		end := start.Add(2 * time.Hour)
		h := data.HourAt(start)
		when := start.Format("15:04") + "–" + end.Format("15:04")
		if end.Day() != start.Day() {
			when = start.Format("15:04") + l.Text("–次日", "–") + end.Format("15:04")
			if l.English() {
				when += " (+1d)"
			}
		}
		if start.YearDay() != now.YearDay() || start.Year() != now.Year() {
			when = l.Text("明日 ", "Tomorrow ") + when
		}
		out = append(out, PreviewItem{Kind: "hour", Index: h.Index, Start: start, End: end, Name: l.Content(h.Name), When: when,
			Keyword: previewKeyword("hour", h.Index, h.Advice, l), Advice: l.Content(h.Advice), Priority: previewPriority(h.Priority), Color: HourTextColor(h.Index, light)})
		start = end
	}
	return out
}
