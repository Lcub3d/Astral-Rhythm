package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const appVersion = "1.0.0"
const appID = "Astral"
const appNameZH = "七曜工作法"
const appNameEN = "Astral Rhythm"

// English strings are local, reviewable translations of the bundled source cells.
// Matching by exact source text prevents changed workbooks from displaying stale advice.
//
//go:embed locales/en.json
var englishJSON []byte

type Localizer struct {
	Language string
	Catalog  map[string]string
}

func ResolveLanguage(preference, system string) string {
	if preference == "en" {
		return "en"
	}
	if preference == "zh-CN" {
		return "zh-CN"
	}
	if strings.HasPrefix(strings.ToLower(system), "zh") {
		return "zh-CN"
	}
	return "en"
}
func NewLocalizer(language string) Localizer {
	var m map[string]string
	if e := json.Unmarshal(englishJSON, &m); e != nil {
		panic(e)
	}
	return Localizer{Language: language, Catalog: m}
}
func (l Localizer) English() bool { return l.Language == "en" }
func (l Localizer) Text(zh, en string) string {
	if l.English() {
		return en
	}
	return zh
}
func (l Localizer) Name() string  { return l.Text(appNameZH, appNameEN) }
func (l Localizer) Title() string { return l.Name() + " V" + appVersion }
func (l Localizer) Content(source string) string {
	if l.English() {
		if v, ok := l.Catalog[source]; ok && v != "" {
			return v
		}
	}
	return source
}
func (l *Localizer) LoadOverrides(base string) error {
	p := filepath.Join(base, "locales", "en.json")
	st, e := os.Stat(p)
	if os.IsNotExist(e) {
		return nil
	}
	if e != nil {
		return e
	}
	if st.Size() > 2<<20 {
		return fmt.Errorf("English catalog exceeds 2 MB")
	}
	b, e := os.ReadFile(p)
	if e != nil {
		return e
	}
	var m map[string]string
	if e = json.Unmarshal([]byte(strings.TrimPrefix(string(b), "\ufeff")), &m); e != nil {
		return fmt.Errorf("locales/en.json: %w", e)
	}
	// Validate everything first: a malformed catalog never partially updates the UI.
	for k, v := range m {
		if len(k) > 32768 || len(v) > 32768 || strings.ContainsRune(k, 0) || strings.ContainsRune(v, 0) {
			return fmt.Errorf("invalid English catalog entry")
		}
	}
	for k, v := range m {
		if k != "" && v != "" {
			l.Catalog[k] = v
		}
	}
	return nil
}
func (l Localizer) Day(d Day) Day {
	d.Name = l.Content(d.Name)
	d.WeekLabel = l.Content(d.WeekLabel)
	d.China = l.Content(d.China)
	d.Advice = l.Content(d.Advice)
	d.Role = l.Content(d.Role)
	d.Sutra = l.Content(d.Sutra)
	d.West = l.Content(d.West)
	return d
}
func (l Localizer) DayShort(d Day) string { return l.Content(d.Short()) }
func (l Localizer) Hour(h Hour) Hour {
	h.Name = l.Content(h.Name)
	h.Stems = l.Content(h.Stems)
	h.Meaning = l.Content(h.Meaning)
	h.Advice = l.Content(h.Advice)
	h.Priority = l.Content(h.Priority)
	return h
}
func (l Localizer) CurrentText(data Data, now time.Time) string {
	if !l.English() {
		text := CurrentText(data, now)
		_, body, _ := strings.Cut(text, "\r\n")
		return fmt.Sprintf("%s  %s\r\n%s", l.Title(), now.Format("2006-01-02 15:04"), body)
	}
	raw := data.HourAt(now)
	day, h := l.Day(data.DayAt(now)), l.Hour(raw)
	var b strings.Builder
	fmt.Fprintf(&b, "%s %s  %s\r\n\r\n[Today · Day Star]\r\n%s  %s  [%s]\r\nBest for: %s\r\nSymbolism: %s\r\nSutra: %s\r\nWestern: %s\r\n", l.Name(), appVersion, now.Format("2006-01-02 15:04"), day.WeekLabel, day.Name, day.Role, day.Advice, day.China, day.Sutra, day.West)
	fmt.Fprintf(&b, "\r\n[Now · Shichen]\r\n%s  %s\r\nBranch: %s\r\nStems: %s\r\nMeaning: %s\r\nBest for: %s\r\n", h.Name, h.Period, l.Content(BranchLabel(raw)), h.Stems, h.Meaning, h.Advice)
	if h.Priority != "" {
		fmt.Fprintf(&b, "Priority: %s\r\n", h.Priority)
	}
	return b.String()
}
func (l Localizer) AllText(data Data) string {
	if !l.English() {
		return AllText(data)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%s | All guidance\r\n\r\nSeven Day Stars\r\n\r\n", l.Name())
	for n := 1; n <= 7; n++ {
		for _, raw := range data.Days {
			if raw.Weekday != n%7 {
				continue
			}
			d := l.Day(raw)
			fmt.Fprintf(&b, "%s  %s  [%s]\r\nBest for: %s\r\nSymbolism: %s\r\nSutra: %s\r\nWestern: %s\r\n\r\n", d.WeekLabel, d.Name, d.Role, d.Advice, d.China, d.Sutra, d.West)
		}
	}
	b.WriteString("Twelve Shichen (two-hour periods)\r\n\r\n")
	for i := 0; i < 12; i++ {
		raw := data.HourByIndex(i)
		h := l.Hour(raw)
		fmt.Fprintf(&b, "%s  %s\r\nBranch: %s\r\nStems: %s\r\nMeaning: %s\r\nBest for: %s\r\n", h.Name, h.Period, l.Content(BranchLabel(raw)), h.Stems, h.Meaning, h.Advice)
		if h.Priority != "" {
			fmt.Fprintf(&b, "Priority: %s\r\n", h.Priority)
		}
		b.WriteString("\r\n")
	}
	return b.String()
}

func isCurrentWindowTitle(title string) bool {
	return title == NewLocalizer("zh-CN").Title() || title == NewLocalizer("en").Title()
}
