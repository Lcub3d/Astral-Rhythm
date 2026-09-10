package main

import (
	"archive/zip"
	_ "embed"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// All bundled advice comes from the two supplied spreadsheets. No network is used.
//
//go:embed data.json
var builtinJSON []byte

const hourFile = "中国十二时辰_藏干与适宜事项(1).xlsx"
const dayFile = "一周七日_七曜与适宜事项(1).xlsx"

type Day struct {
	Weekday   int    `json:"weekday"`
	WeekLabel string `json:"week_label"`
	Name      string `json:"name"`
	China     string `json:"china"`
	Advice    string `json:"advice"`
	Role      string `json:"role"`
	Sutra     string `json:"sutra"`
	West      string `json:"west"`
}

func (d Day) Short() string { return strings.SplitN(d.Name, "·", 2)[0] }

type Hour struct {
	Index    int    `json:"index"`
	Name     string `json:"name"`
	Period   string `json:"period"`
	Stems    string `json:"stems"`
	Meaning  string `json:"meaning"`
	Advice   string `json:"advice"`
	Priority string `json:"priority"`
}
type Data struct {
	Version    int      `json:"version"`
	Days       []Day    `json:"days"`
	Hours      []Hour   `json:"hours"`
	Notes      []string `json:"notes"`
	Sources    []string `json:"sources"`
	References []string `json:"references"`
}

var dayLabels = []string{"周日", "周一", "周二", "周三", "周四", "周五", "周六"}
var hourLabels = []string{"子时", "丑时", "寅时", "卯时", "辰时", "巳时", "午时", "未时", "申时", "酉时", "戌时", "亥时"}
var shortDayLabels = []string{"日曜", "月曜", "火曜", "水曜", "木曜", "金曜", "土曜"}

func (d Data) Validate() error {
	if len(d.Days) != 7 || len(d.Hours) != 12 {
		return errors.New("需要完整的 7 个七曜和 12 个时辰")
	}
	ds, hs := map[int]bool{}, map[int]bool{}
	for _, v := range d.Days {
		if v.Weekday < 0 || v.Weekday > 6 || ds[v.Weekday] || v.Advice == "" || v.Short() != shortDayLabels[v.Weekday] {
			return errors.New("七曜表存在遗漏、重复或星期与七曜对应不正确")
		}
		ds[v.Weekday] = true
	}
	for _, v := range d.Hours {
		if v.Index < 0 || v.Index > 11 || hs[v.Index] || v.Name != hourLabels[v.Index] || v.Advice == "" {
			return errors.New("时辰表存在遗漏、重复或无法识别的时辰")
		}
		expected := fmt.Sprintf("%02d:00–%02d:00", (23+2*v.Index)%24, (1+2*v.Index)%24)
		normalize := strings.NewReplacer(" ", "", "—", "-", "–", "-", "~", "-", "～", "-", "：", ":")
		if normalize.Replace(v.Period) != normalize.Replace(expected) {
			return fmt.Errorf("%s 的时间须为 %s", v.Name, expected)
		}
		hs[v.Index] = true
	}
	return nil
}
func ParseData(b []byte) (Data, error) {
	var d Data
	e := json.Unmarshal([]byte(strings.TrimPrefix(string(b), "\ufeff")), &d)
	if e == nil {
		e = d.Validate()
	}
	return d, e
}
func DefaultData() Data {
	d, e := ParseData(builtinJSON)
	if e != nil {
		panic(e)
	}
	return d
}
func (d Data) DayAt(t time.Time) Day {
	for _, v := range d.Days {
		if v.Weekday == int(t.Weekday()) {
			return v
		}
	}
	return Day{}
}
func HourIndex(t time.Time) int { return ((t.Hour() + 1) / 2) % 12 }
func (d Data) HourAt(t time.Time) Hour {
	for _, v := range d.Hours {
		if v.Index == HourIndex(t) {
			return v
		}
	}
	return Hour{}
}
func (d Data) HourByIndex(i int) Hour {
	for _, v := range d.Hours {
		if v.Index == i {
			return v
		}
	}
	return Hour{}
}

// Seven-day labels roll over at local midnight. The 子时 interval spans midnight.
// The Windows UI passes the system's current civil time as a UTC-shaped wall clock,
// avoiding a cached timezone after the user changes Windows timezone settings.
func NextBoundary(t time.Time) time.Time {
	midnight := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	next := 2*((t.Hour()+1)/2) + 1
	if next >= 24 {
		return midnight.AddDate(0, 0, 1).Add(time.Hour)
	}
	return time.Date(t.Year(), t.Month(), t.Day(), next, 0, 0, 0, t.Location())
}
func Countdown(t time.Time) string {
	min := int((NextBoundary(t).Sub(t) + time.Minute - 1) / time.Minute)
	if min < 1 {
		min = 1
	}
	if min >= 60 {
		if min%60 == 0 {
			return fmt.Sprintf("%d小时", min/60)
		}
		return fmt.Sprintf("%d小时%d分", min/60, min%60)
	}
	return fmt.Sprintf("%d分", min)
}
func CurrentText(d Data, now time.Time) string {
	day, h := d.DayAt(now), d.HourAt(now)
	var b strings.Builder
	fmt.Fprintf(&b, "七曜工作法 V3.1  %s\r\n\r\n【今日 · 曜日】\r\n%s  %s  [%s]\r\n今日方向：%s\r\n", now.Format("2006-01-02 15:04"), day.WeekLabel, day.Name, day.Role, day.Advice)
	for _, e := range DayExplanations(day) {
		fmt.Fprintf(&b, "%s：%s\r\n", e.Label, e.Value)
	}
	fmt.Fprintf(&b, "\r\n【此刻 · 时辰】\r\n%s  %s\r\n地支：%s\r\n现在适合：%s\r\n", h.Name, h.Period, BranchLabel(h), h.Advice)
	if h.Priority != "" {
		fmt.Fprintf(&b, "工作等级：%s\r\n", strings.ToUpper(h.Priority))
	}
	for _, e := range HourExplanations(h) {
		fmt.Fprintf(&b, "%s：%s\r\n", e.Label, e.Value)
	}
	return b.String()
}
func AllText(d Data) string {
	var b strings.Builder
	b.WriteString("七曜工作法｜全部安排\r\n一周七曜\r\n\r\n")
	for n := 1; n <= 7; n++ {
		for _, x := range d.Days {
			if x.Weekday == n%7 {
				fmt.Fprintf(&b, "%s  %s  [%s]\r\n现代安排：%s\r\n中国传统象征：%s\r\n《宿曜经》倾向：%s\r\n西方传统象征：%s\r\n\r\n", x.WeekLabel, x.Name, x.Role, x.Advice, x.China, x.Sutra, x.West)
			}
		}
	}
	b.WriteString("十二时辰\r\n\r\n")
	for n := 0; n < 12; n++ {
		x := d.HourByIndex(n)
		fmt.Fprintf(&b, "%s  %s\r\n适宜事项：%s\r\n地支藏干：%s\r\n传统含义：%s\r\n", x.Name, x.Period, x.Advice, x.Stems, x.Meaning)
		if x.Priority != "" {
			fmt.Fprintf(&b, "工作等级：%s\r\n", strings.ToUpper(x.Priority))
		}
		b.WriteString("\r\n")
	}
	return b.String()
}

// Read the first worksheet, including inline/shared/rich strings, using only the
// standard library. It does not execute macros, formulas, or external links.
type richText struct {
	Text string `xml:"t"`
	Runs []struct {
		Text string `xml:"t"`
	} `xml:"r"`
}

func (r richText) String() string {
	var b strings.Builder
	b.WriteString(r.Text)
	for _, v := range r.Runs {
		b.WriteString(v.Text)
	}
	return b.String()
}
func firstSheet(file string) ([][]string, error) {
	st, e := os.Stat(file)
	if e != nil {
		return nil, e
	}
	if st.Size() > 16<<20 {
		return nil, errors.New("表格过大（超过16MB）")
	}
	z, e := zip.OpenReader(file)
	if e != nil {
		return nil, e
	}
	defer z.Close()
	files := map[string]*zip.File{}
	for _, f := range z.File {
		files[f.Name] = f
	}
	read := func(name string, out any) error {
		f := files[name]
		if f == nil {
			return fmt.Errorf("缺少 XLSX 条目 %s", name)
		}
		if f.UncompressedSize64 > 8<<20 {
			return errors.New("表格内容过大")
		}
		r, e := f.Open()
		if e != nil {
			return e
		}
		defer r.Close()
		return xml.NewDecoder(io.LimitReader(r, 8<<20)).Decode(out)
	}
	var ss struct {
		Items []richText `xml:"si"`
	}
	if files["xl/sharedStrings.xml"] != nil {
		if e = read("xl/sharedStrings.xml", &ss); e != nil {
			return nil, e
		}
	}
	var book struct {
		Sheets []struct {
			ID string `xml:"id,attr"`
		} `xml:"sheets>sheet"`
	}
	if e = read("xl/workbook.xml", &book); e != nil {
		return nil, e
	}
	if len(book.Sheets) == 0 {
		return nil, errors.New("工作簿没有工作表")
	}
	var rels struct {
		Items []struct {
			ID     string `xml:"Id,attr"`
			Target string `xml:"Target,attr"`
			Mode   string `xml:"TargetMode,attr"`
		} `xml:"Relationship"`
	}
	if e = read("xl/_rels/workbook.xml.rels", &rels); e != nil {
		return nil, e
	}
	target := ""
	for _, r := range rels.Items {
		if r.ID == book.Sheets[0].ID && r.Mode != "External" {
			if strings.HasPrefix(r.Target, "/") {
				target = strings.TrimPrefix(r.Target, "/")
			} else {
				target = path.Join("xl", r.Target)
			}
			break
		}
	}
	if target == "" {
		return nil, errors.New("无法找到第一个工作表")
	}
	var sheet struct {
		Rows []struct {
			Cells []struct {
				Ref    string   `xml:"r,attr"`
				Type   string   `xml:"t,attr"`
				Value  string   `xml:"v"`
				Inline richText `xml:"is"`
			} `xml:"c"`
		} `xml:"sheetData>row"`
	}
	if e = read(target, &sheet); e != nil {
		return nil, e
	}
	var rows [][]string
	for _, row := range sheet.Rows {
		vals := make([]string, 32)
		for _, c := range row.Cells {
			col := 0
			for _, ch := range c.Ref {
				if ch < 'A' || ch > 'Z' {
					break
				}
				col = col*26 + int(ch-'A'+1)
			}
			if col < 1 || col > len(vals) {
				continue
			}
			v := c.Value
			switch c.Type {
			case "s":
				n, err := strconv.Atoi(v)
				if err != nil || n < 0 || n >= len(ss.Items) {
					return nil, errors.New("无效的共享字符串索引")
				}
				v = ss.Items[n].String()
			case "inlineStr":
				v = c.Inline.String()
			}
			vals[col-1] = strings.TrimSpace(v)
		}
		rows = append(rows, vals)
	}
	return rows, nil
}
func headerMap(rows [][]string, required string) (int, map[string]int, error) {
	for i, r := range rows {
		m := map[string]int{}
		for j, c := range r {
			if c != "" {
				m[c] = j
			}
		}
		if _, ok := m[required]; ok {
			return i, m, nil
		}
	}
	return 0, nil, fmt.Errorf("找不到表头“%s”", required)
}
func requireHeaders(m map[string]int, names ...string) error {
	for _, name := range names {
		if _, ok := m[name]; !ok {
			return fmt.Errorf("缺少原表列：%s（请保留表头）", name)
		}
	}
	return nil
}
func tableValue(r []string, m map[string]int, k string) string {
	if i, ok := m[k]; ok && i < len(r) {
		return r[i]
	}
	return ""
}
func LoadExcel(dir string) (Data, error) {
	d := DefaultData()
	d.Days = nil
	d.Hours = nil
	rows, e := firstSheet(filepath.Join(dir, dayFile))
	if e != nil {
		return d, fmt.Errorf("七曜表：%w", e)
	}
	h, m, e := headerMap(rows, "星期")
	if e != nil {
		return d, e
	}
	if e = requireHeaders(m, "星期", "七曜", "中国传统象征", "现代最适合安排", "定位", "中国古代《宿曜经》倾向", "西方传统象征"); e != nil {
		return d, e
	}
	for _, r := range rows[h+1:] {
		label := tableValue(r, m, "星期")
		if label == "星期日" || label == "周天" {
			label = "周日"
		}
		for i, l := range dayLabels {
			if label == l {
				d.Days = append(d.Days, Day{i, label, tableValue(r, m, "七曜"), tableValue(r, m, "中国传统象征"), tableValue(r, m, "现代最适合安排"), tableValue(r, m, "定位"), tableValue(r, m, "中国古代《宿曜经》倾向"), tableValue(r, m, "西方传统象征")})
				break
			}
		}
	}
	rows, e = firstSheet(filepath.Join(dir, hourFile))
	if e != nil {
		return d, fmt.Errorf("时辰表：%w", e)
	}
	h, m, e = headerMap(rows, "时辰")
	if e != nil {
		return d, e
	}
	if e = requireHeaders(m, "时辰", "现代时间", "地支藏干", "传统含义", "比较适合做什么", "对应工作等级"); e != nil {
		return d, e
	}
	for _, r := range rows[h+1:] {
		label := tableValue(r, m, "时辰")
		for i, l := range hourLabels {
			if label == l {
				d.Hours = append(d.Hours, Hour{i, label, tableValue(r, m, "现代时间"), tableValue(r, m, "地支藏干"), tableValue(r, m, "传统含义"), tableValue(r, m, "比较适合做什么"), tableValue(r, m, "对应工作等级")})
				break
			}
		}
	}
	return d, d.Validate()
}
func LoadData(base string) (Data, string, error) {
	tableDir := filepath.Join(base, "表格")
	_, e1 := os.Stat(filepath.Join(tableDir, dayFile))
	_, e2 := os.Stat(filepath.Join(tableDir, hourFile))
	if e1 == nil || e2 == nil {
		d, e := LoadExcel(tableDir)
		if e == nil {
			return d, "原始 Excel 表格", nil
		}
		return DefaultData(), "内置数据", e
	}
	if b, e := os.ReadFile(filepath.Join(base, "data.json")); e == nil {
		d, e := ParseData(b)
		if e == nil {
			return d, "data.json", nil
		}
		return DefaultData(), "内置数据", e
	}
	return DefaultData(), "内置数据", nil
}

type Settings struct {
	Language       string `json:"language"`
	UIVersion      int    `json:"ui_version"`
	Gap            int    `json:"gap"` // distance from notification area, in 96-DPI logical pixels
	FontSize       int    `json:"font_size"`
	Theme          string `json:"theme"`
	Traditional    bool   `json:"traditional,omitempty"` // legacy v1 field; no longer rendered
	HourDetails    bool   `json:"hour_details"`
	DayDetails     bool   `json:"day_details"`
	HideFullscreen bool   `json:"hide_fullscreen"`
	Autostart      bool   `json:"autostart"`
}

func DefaultSettings() Settings {
	return Settings{Language: "system", UIVersion: 3, Gap: 96, FontSize: 13, Theme: "dark", HourDetails: true, DayDetails: false, HideFullscreen: true}
}
func (s *Settings) Normalize() {
	if s.Language != "en" && s.Language != "zh-CN" {
		s.Language = "system"
	}
	if s.Gap < -1200 {
		s.Gap = -1200
	}
	if s.Gap > 12000 {
		s.Gap = 12000
	}
	if s.FontSize < 10 {
		s.FontSize = 10
	}
	if s.FontSize > 18 {
		s.FontSize = 18
	}
	if s.Theme != "dark" && s.Theme != "light" {
		s.Theme = "system"
	}
}
func settingsPath() string {
	d, e := os.UserConfigDir()
	if e != nil {
		d = os.TempDir()
	}
	return filepath.Join(d, appID, "settings.json")
}
func legacySettingsPath() string {
	return filepath.Join(filepath.Dir(filepath.Dir(settingsPath())), "QiyaoShichen", "settings.json")
}
func ReadSettings() Settings {
	s := DefaultSettings()
	b, e := os.ReadFile(settingsPath())
	if os.IsNotExist(e) {
		b, e = os.ReadFile(legacySettingsPath())
	}
	if e == nil {
		candidate := s
		if e = json.Unmarshal(b, &candidate); e == nil {
			s = candidate
			var ver struct {
				UIVersion int `json:"ui_version"`
			}
			_ = json.Unmarshal(b, &ver)
			if ver.UIVersion < 3 {
				s.HourDetails = true
				s.DayDetails = false
			}
		}
	}
	s.UIVersion = 3
	s.Normalize()
	return s
}
func SaveSettings(s Settings) error {
	s.Normalize()
	p := settingsPath()
	if e := os.MkdirAll(filepath.Dir(p), 0700); e != nil {
		return e
	}
	b, e := json.MarshalIndent(s, "", "  ")
	if e != nil {
		return e
	}
	f, e := os.CreateTemp(filepath.Dir(p), "settings-*.tmp")
	if e != nil {
		return e
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	if _, e = f.Write(b); e != nil {
		f.Close()
		return e
	}
	if e = f.Close(); e != nil {
		return e
	}
	// On Windows rename does not atomically replace an existing destination.
	// The original is kept unless writing the new complete file succeeded.
	if e = os.Rename(tmp, p); e == nil {
		return nil
	}
	return os.WriteFile(p, b, 0600)
}
