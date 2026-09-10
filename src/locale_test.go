package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
	"unicode"
)

func hasHan(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}
func TestAstralBrandNames(t *testing.T) {
	if appID != "Astral" || appNameZH != "七曜工作法" || appNameEN != "Astral Rhythm" || appVersion != "1.0.0" {
		t.Fatal("incorrect product identity")
	}
	for _, lang := range []string{"zh-CN", "en"} {
		l := NewLocalizer(lang)
		if !isCurrentWindowTitle(l.Title()) {
			t.Fatal("current window is not recognized")
		}
		if !strings.Contains(strings.SplitN(l.CurrentText(DefaultData(), time.Now()), "\r\n", 2)[0], appVersion) {
			t.Fatal("export header does not use the public version")
		}
		if !strings.HasSuffix(l.Title(), "V"+appVersion) {
			t.Fatal("incorrect title version")
		}
	}
	if isCurrentWindowTitle("七曜时辰 V3") || isCurrentWindowTitle("Unrelated app") {
		t.Fatal("legacy/other window confused with current build")
	}
}
func TestLanguageResolution(t *testing.T) {
	for _, c := range []struct{ preference, system, want string }{
		{"system", "zh-CN", "zh-CN"}, {"system", "zh-TW", "zh-CN"}, {"system", "ZH-HK", "zh-CN"},
		{"system", "en-US", "en"}, {"system", "de-DE", "en"}, {"system", "", "en"},
		{"en", "zh-CN", "en"}, {"zh-CN", "en-US", "zh-CN"},
	} {
		if got := ResolveLanguage(c.preference, c.system); got != c.want {
			t.Fatalf("%+v => %s", c, got)
		}
	}
}
func TestLanguageSettingsRoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("APPDATA", dir)
	for _, lang := range []string{"en", "zh-CN", "system"} {
		s := DefaultSettings()
		s.Language = lang
		s.Gap = 231
		if e := SaveSettings(s); e != nil {
			t.Fatal(e)
		}
		got := ReadSettings()
		if got.Language != lang || got.Gap != 231 {
			t.Fatalf("lost settings: %+v", got)
		}
	}
	s := DefaultSettings()
	s.Language = "invalid"
	s.Normalize()
	if s.Language != "system" {
		t.Fatal("invalid preference should follow system")
	}
	if filepath.Base(filepath.Dir(settingsPath())) != "Astral" {
		t.Fatal("wrong app settings folder")
	}
}
func TestEnglishCatalogCoversAllBundledAdvice(t *testing.T) {
	d := DefaultData()
	l := NewLocalizer("en")
	n := 0
	check := func(s string) {
		if s == "" {
			return
		}
		if v := l.Content(s); v == s || hasHan(v) {
			t.Errorf("missing translation: %q -> %q", s, v)
		}
		n++
	}
	for _, d := range d.Days {
		for _, s := range []string{d.WeekLabel, d.Name, d.Short(), d.China, d.Advice, d.Role, d.Sutra, d.West, DayElement(d.Weekday)} {
			check(s)
		}
	}
	for _, h := range d.Hours {
		for _, s := range []string{h.Name, h.Stems, h.Meaning, h.Advice, h.Priority, BranchLabel(h)} {
			check(s)
		}
	}
	t.Logf("%d nonempty source fields have English translations", n)
}
func TestTranslationPreservesSourceAndIdentity(t *testing.T) {
	d := DefaultData()
	before, _ := json.Marshal(d)
	l := NewLocalizer("en")
	for _, raw := range d.Days {
		v := l.Day(raw)
		if v.Weekday != raw.Weekday || v.Advice == raw.Advice {
			t.Fatal("day translation failed")
		}
	}
	for _, raw := range d.Hours {
		v := l.Hour(raw)
		if v.Index != raw.Index || v.Period != raw.Period || v.Advice == raw.Advice {
			t.Fatal("hour translation failed")
		}
	}
	after, _ := json.Marshal(d)
	if string(before) != string(after) {
		t.Fatal("source data mutated")
	}
	if l.Content("用户修改的新工作建议") != "用户修改的新工作建议" {
		t.Fatal("unknown advice must not use an unrelated translation")
	}
	zh := NewLocalizer("zh-CN")
	if !reflect.DeepEqual(zh.Day(d.Days[0]), d.Days[0]) || !reflect.DeepEqual(zh.Hour(d.Hours[0]), d.Hours[0]) {
		t.Fatal("Chinese source altered")
	}
}
func TestEnglishOverrideExactMatchAndAtomicValidation(t *testing.T) {
	base := t.TempDir()
	p := filepath.Join(base, "locales")
	if e := os.Mkdir(p, 0700); e != nil {
		t.Fatal(e)
	}
	write := func(s string) {
		if e := os.WriteFile(filepath.Join(p, "en.json"), []byte(s), 0600); e != nil {
			t.Fatal(e)
		}
	}
	l := NewLocalizer("en")
	write("\ufeff" + `{"用户修改的新工作建议":"A custom work plan", "木":"Timber"}`)
	if e := l.LoadOverrides(base); e != nil {
		t.Fatal(e)
	}
	if l.Content("木") != "Timber" || l.Content("用户修改的新工作建议") != "A custom work plan" {
		t.Fatal("overrides missing")
	}
	write(`{"木":"Overwritten", "invalid":"bad\u0000text"}`)
	if e := l.LoadOverrides(base); e == nil {
		t.Fatal("NUL should be rejected")
	}
	if l.Content("木") != "Timber" {
		t.Fatal("invalid override partially applied")
	}
	write(`{"木":`)
	if e := l.LoadOverrides(base); e == nil {
		t.Fatal("broken JSON should be rejected")
	}
	if l.Content("木") != "Timber" {
		t.Fatal("parse failure changed dictionary")
	}
}
func TestEnglishExportsAndChineseBrand(t *testing.T) {
	d := DefaultData()
	now := time.Date(2026, 9, 8, 13, 40, 0, 0, time.UTC)
	en := NewLocalizer("en")
	for _, s := range []string{en.CurrentText(d, now), en.AllText(d)} {
		if !strings.Contains(s, "Astral") || hasHan(s) {
			t.Fatalf("English export not fully localized: %s", s)
		}
	}
	zh := NewLocalizer("zh-CN")
	if !strings.HasPrefix(zh.CurrentText(d, now), "七曜工作法") || !strings.HasPrefix(zh.AllText(d), "七曜工作法") {
		t.Fatal("Chinese export not renamed")
	}
}
func TestLatinWordWrapping(t *testing.T) {
	m := func(s string, _ int, _ bool) int { return len([]rune(s)) }
	got := WrapText("Plan review and write reports", 12, 12, false, m)
	if compact(strings.Join(got, "")) != compact("Plan review and write reports") {
		t.Fatal("lost characters", got)
	}
	for _, s := range got {
		if len([]rune(s)) > 12 {
			t.Fatal("line too wide", got)
		}
	}
	joined := strings.Join(got, "|")
	if strings.Contains(joined, "rev|iew") || strings.Contains(joined, "repo|rts") {
		t.Fatal("split a short word", got)
	}
	got = WrapText("AAAAAAAAAAAAAAAAAAAA", 5, 12, false, m)
	if strings.Join(got, "") != "AAAAAAAAAAAAAAAAAAAA" {
		t.Fatal("lost long-word text")
	}
	for _, s := range got {
		if len(s) > 5 {
			t.Fatal("long token overflow")
		}
	}
}
func TestBilingualLayoutsAllPairs(t *testing.T) {
	d := DefaultData()
	count := 0
	for _, lang := range []string{"zh-CN", "en"} {
		localizer := NewLocalizer(lang)
		for _, dpi := range []int{96, 120, 144, 192, 288} {
			measure := testMeasure(dpi)
			for _, width := range []int{240, 320} {
				for _, light := range []bool{false, true} {
					for _, details := range []bool{false, true} {
						for _, day := range d.Days {
							for _, hour := range d.Hours {
								l := CompactLayout(day, hour, CardOptions{Locale: localizer, DPI: dpi, MaxWidth: width * dpi / 96, MaxHeight: 480 * dpi / 96, Light: light, DayDetails: details, HourDetails: true}, measure)
								if l.Blocks[0].Text != localizer.Text("曜星", "Day Star") {
									t.Fatal("wrong section title")
								}
								for _, b := range l.Blocks {
									if b.Rect.Left < 0 || int(b.Rect.Right) > l.Width || b.Rect.Top < 0 || int(b.Rect.Bottom) > l.ContentHeight {
										t.Fatalf("%s dpi=%d width=%d: block outside layout %+v", lang, dpi, width, b)
									}
									if got := measure(b.Text, b.Size, b.Bold); got > b.Rect.W()+1 {
										t.Fatalf("%s dpi=%d width=%d: %q width %d > %d", lang, dpi, width, b.Text, got, b.Rect.W())
									}
									if lang == "en" && hasHan(b.Text) {
										t.Fatalf("untranslated English layout field: %q", b.Text)
									}
								}
								ds, hs := sectionString(l, "day"), sectionString(l, "hour")
								if !strings.Contains(ds, compact(localizer.Content(day.Advice))) || !strings.Contains(hs, compact(localizer.Content(hour.Advice))) {
									t.Fatal("lost source advice")
								}
								if strings.Contains(ds, compact(localizer.Content(hour.Meaning))) || strings.Contains(hs, compact(localizer.Content(day.China))) {
									t.Fatal("mixed sections")
								}
								count++
							}
						}
					}
				}
			}
		}
	}
	t.Logf("%d bilingual layouts validated (estimated font widths, not Windows GUI execution)", count)
}
func TestLegacyV3SettingsMigrateToAstral(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("APPDATA", dir)
	t.Setenv("XDG_CONFIG_HOME", dir)
	p := legacySettingsPath()
	if e := os.MkdirAll(filepath.Dir(p), 0700); e != nil {
		t.Fatal(e)
	}
	old := `{"ui_version":3,"gap":132,"font_size":14,"theme":"light","hour_details":false,"day_details":true}`
	if e := os.WriteFile(p, []byte(old), 0600); e != nil {
		t.Fatal(e)
	}
	s := ReadSettings()
	if s.Gap != 132 || s.FontSize != 14 || s.Theme != "light" || s.HourDetails || !s.DayDetails || s.Language != "system" {
		t.Fatalf("V3 migration failed: %+v", s)
	}
	s.Language = "en"
	if e := SaveSettings(s); e != nil {
		t.Fatal(e)
	}
	if got := ReadSettings(); got.Language != "en" {
		t.Fatal("new preferences lost")
	}
	b, _ := os.ReadFile(p)
	if string(b) != old {
		t.Fatal("legacy file changed")
	}
}
