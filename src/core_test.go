package main

import (
	"archive/zip"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestEmbeddedDataset(t *testing.T) {
	d := DefaultData()
	if e := d.Validate(); e != nil {
		t.Fatal(e)
	}
	if len(d.Days) != 7 || len(d.Hours) != 12 {
		t.Fatal("bad entry count")
	}
}
func TestEveryMinuteOfDay(t *testing.T) {
	counts := [12]int{}
	d := DefaultData()
	for m := 0; m < 1440; m++ {
		now := time.Date(2026, 9, 9, m/60, m%60, 0, 0, time.UTC)
		want := ((m + 60) / 120) % 12
		got := HourIndex(now)
		if got != want {
			t.Fatalf("%02d:%02d got %d want %d", m/60, m%60, got, want)
		}
		if d.HourAt(now).Name != hourLabels[want] {
			t.Fatal("wrong hour name")
		}
		end := NextBoundary(now)
		if !end.After(now) || end.Sub(now) > 2*time.Hour || end.Minute() != 0 || end.Hour()%2 != 1 {
			t.Fatalf("invalid next boundary %v at %v", end, now)
		}
		counts[got]++
	}
	for i, c := range counts {
		if c != 120 {
			t.Fatalf("%s spans %d minutes, expected 120", hourLabels[i], c)
		}
	}
}
func TestBoundarySeconds(t *testing.T) {
	cases := []struct{ input, hour, end string }{
		{"2026-09-08 22:59:59", "亥时", "2026-09-08 23:00:00"},
		{"2026-09-08 23:00:00", "子时", "2026-09-09 01:00:00"},
		{"2026-09-08 23:59:59", "子时", "2026-09-09 01:00:00"},
		{"2026-09-09 00:00:00", "子时", "2026-09-09 01:00:00"},
		{"2026-09-09 00:59:59", "子时", "2026-09-09 01:00:00"},
		{"2026-09-09 01:00:00", "丑时", "2026-09-09 03:00:00"},
		{"2026-12-31 23:30:00", "子时", "2027-01-01 01:00:00"},
		{"2028-02-29 23:59:59", "子时", "2028-03-01 01:00:00"},
	}
	d := DefaultData()
	const layout = "2006-01-02 15:04:05"
	for _, v := range cases {
		now, e := time.Parse(layout, v.input)
		if e != nil {
			t.Fatal(e)
		}
		if d.HourAt(now).Name != v.hour {
			t.Errorf("%s: wrong hour", v.input)
		}
		if NextBoundary(now).Format(layout) != v.end {
			t.Errorf("%s: incorrect midnight boundary", v.input)
		}
	}
}
func TestWeekdaysAndMidnight(t *testing.T) {
	d := DefaultData()
	monday := time.Date(2026, 9, 7, 12, 0, 0, 0, time.UTC)
	expected := []string{"月曜", "火曜", "水曜", "木曜", "金曜", "土曜", "日曜"}
	for i := 0; i < 28; i++ {
		if got := d.DayAt(monday.AddDate(0, 0, i)).Short(); got != expected[i%7] {
			t.Fatalf("day %d: %s", i, got)
		}
	}
	pre := time.Date(2026, 9, 8, 23, 59, 59, 0, time.UTC)
	post := pre.Add(time.Second)
	if d.HourAt(pre).Name != "子时" || d.HourAt(post).Name != "子时" {
		t.Fatal("子时 must span midnight")
	}
	if d.DayAt(pre).Short() != "火曜" || d.DayAt(post).Short() != "水曜" {
		t.Fatal("weekday must change at 00:00, not 23:00")
	}
}
func TestCountdown(t *testing.T) {
	tests := []struct {
		h, m, s int
		want    string
	}{{13, 0, 0, "2小时"}, {13, 40, 0, "1小时20分"}, {14, 59, 59, "1分"}, {23, 30, 0, "1小时30分"}, {0, 0, 0, "1小时"}, {0, 59, 59, "1分"}}
	for _, v := range tests {
		got := Countdown(time.Date(2026, 9, 9, v.h, v.m, v.s, 0, time.UTC))
		if got != v.want {
			t.Errorf("countdown = %q want %q", got, v.want)
		}
	}
}
func TestWorkPriorities(t *testing.T) {
	d := DefaultData()
	expected := map[int]string{5: "p1", 7: "p2、p3", 8: "p1", 9: "p3"}
	for _, h := range d.Hours {
		if h.Priority != expected[h.Index] {
			t.Errorf("%s: priority must come from source: %q", h.Name, h.Priority)
		}
	}
}
func TestActualSpreadsheetsMatchEmbeddedData(t *testing.T) {
	dir := filepath.Join("..", "表格")
	if _, e := os.Stat(filepath.Join(dir, dayFile)); e != nil {
		t.Skip("place the supplied 表格 folder beside src for the source-data integration test")
	}
	got, e := LoadExcel(dir)
	if e != nil {
		t.Fatal(e)
	}
	want := DefaultData()
	if !reflect.DeepEqual(got.Days, want.Days) {
		t.Fatalf("weekday Excel content differs from embedded snapshot:\ngot %#v\nwant %#v", got.Days, want.Days)
	}
	if !reflect.DeepEqual(got.Hours, want.Hours) {
		t.Fatal("hour Excel content differs from embedded snapshot")
	}
	t.Log("All 7 weekday records and all 12 hour records match the supplied Excel cells, including blank priority cells.")
}
func TestRejectInvalidData(t *testing.T) {
	d := DefaultData()
	d.Days[1] = d.Days[0]
	if d.Validate() == nil {
		t.Fatal("duplicate weekdays accepted")
	}
	d = DefaultData()
	d.Hours = d.Hours[:11]
	if d.Validate() == nil {
		t.Fatal("missing hour accepted")
	}
	d = DefaultData()
	d.Hours[0].Period = "00:00–02:00"
	if d.Validate() == nil {
		t.Fatal("wrong interval accepted")
	}
	d = DefaultData()
	d.Days[0].Name = "火曜·荧惑"
	if d.Validate() == nil {
		t.Fatal("wrong weekday mapping accepted")
	}
}
func TestDataFallbackAndMalformedWorkbook(t *testing.T) {
	dir := t.TempDir()
	d, s, e := LoadData(dir)
	if e != nil || s != "内置数据" || len(d.Days) != 7 {
		t.Fatal("missing files should use builtin snapshot")
	}
	if e = os.WriteFile(filepath.Join(dir, "data.json"), []byte("broken"), 0600); e != nil {
		t.Fatal(e)
	}
	d, s, e = LoadData(dir)
	if e == nil || s != "内置数据" || d.Validate() != nil {
		t.Fatal("broken JSON must fall back with a reported error")
	}
	table := filepath.Join(dir, "表格")
	_ = os.MkdirAll(table, 0700)
	_ = os.WriteFile(filepath.Join(table, dayFile), []byte("not a ZIP"), 0600)
	_, _, e = LoadData(dir)
	if e == nil {
		t.Fatal("malformed XLSX silently accepted")
	}
}
func TestXLSXSharedInlineAndRelationshipTarget(t *testing.T) {
	file := filepath.Join(t.TempDir(), "sample.xlsx")
	f, e := os.Create(file)
	if e != nil {
		t.Fatal(e)
	}
	z := zip.NewWriter(f)
	parts := map[string]string{
		"xl/workbook.xml":            `<workbook xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheets><sheet name="数据" r:id="rId42"/></sheets></workbook>`,
		"xl/_rels/workbook.xml.rels": `<Relationships><Relationship Id="rId42" Target="/xl/worksheets/custom.xml"/></Relationships>`,
		"xl/sharedStrings.xml":       `<sst><si><r><t>水曜</t></r><r><t>·辰星</t></r></si></sst>`,
		"xl/worksheets/custom.xml":   `<worksheet><sheetData><row r="1"><c r="A1" t="s"><v>0</v></c><c r="C1" t="inlineStr"><is><t>中文&amp;文本</t></is></c></row></sheetData></worksheet>`,
	}
	for n, content := range parts {
		w, e := z.Create(n)
		if e != nil {
			t.Fatal(e)
		}
		if _, e = w.Write([]byte(content)); e != nil {
			t.Fatal(e)
		}
	}
	_ = z.Close()
	_ = f.Close()
	rows, e := firstSheet(file)
	if e != nil {
		t.Fatal(e)
	}
	if rows[0][0] != "水曜·辰星" || rows[0][1] != "" || rows[0][2] != "中文&文本" {
		t.Fatalf("incorrect XML decoding: %v", rows)
	}
}
func TestCopyTextIncludesEverySourceField(t *testing.T) {
	d := DefaultData()
	now := time.Date(2026, 9, 9, 13, 40, 0, 0, time.UTC)
	text := CurrentText(d, now)
	day, h := d.DayAt(now), d.HourAt(now)
	for _, v := range []string{day.Name, day.Role, day.Advice, day.China, day.Sutra, day.West, h.Name, h.Period, h.Advice, h.Stems, h.Meaning, "P2、P3"} {
		if !strings.Contains(text, v) {
			t.Errorf("missing current detail: %s", v)
		}
	}
	all := AllText(d)
	for _, x := range d.Days {
		if !strings.Contains(all, x.Name) {
			t.Fatal("missing weekday in all arrangements")
		}
	}
	for _, x := range d.Hours {
		if !strings.Contains(all, x.Name) {
			t.Fatal("missing hour in all arrangements")
		}
	}
}
func TestSettingsNormalization(t *testing.T) {
	s := Settings{Gap: 999999, FontSize: 2, Theme: "invalid"}
	s.Normalize()
	if s.Gap != 12000 || s.FontSize != 10 || s.Theme != "system" {
		t.Fatal("bad settings normalization")
	}
}
