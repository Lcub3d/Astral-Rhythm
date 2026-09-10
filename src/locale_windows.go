//go:build windows

package main

import (
	"fmt"
	"unsafe"
)

var getUserDefaultUILanguage = kernel32.NewProc("GetUserDefaultUILanguage")
var setWindowText = user32.NewProc("SetWindowTextW")
var startupLocale = NewLocalizer("zh-CN")

const WM_SETLANG = WM_APP + 5

func systemLanguage() string {
	id, _, _ := getUserDefaultUILanguage.Call()
	if id&0x3ff == 0x04 {
		return "zh-CN"
	}
	return "en"
}
func activeLocale() Localizer {
	if app != nil {
		return app.L
	}
	return startupLocale
}
func loadLocalizer(preference, base string) Localizer {
	l := NewLocalizer(ResolveLanguage(preference, systemLanguage()))
	if e := l.LoadOverrides(base); e != nil {
		logError(e)
	}
	return l
}
func (a *App) setLanguage(preference string) {
	a.Settings.Language = preference
	a.Settings.Normalize()
	a.L = loadLocalizer(a.Settings.Language, a.Base)
	a.clearFonts()
	a.Scroll = 0
	a.LastClock = ""
	setWindowText.Call(a.HWND, uintptr(unsafe.Pointer(w(a.L.Title()))))
	setWindowText.Call(a.Popup, uintptr(unsafe.Pointer(w(a.L.Name()+a.L.Text(" · 当前建议", " · Current guidance")))))
	a.save()
	a.positionWidget()
	invalidate(a.HWND)
	a.tray(1)
	if a.CardShown {
		a.positionCard()
	}
}
func (a *App) widgetWidth() int {
	// Reserve enough width for every label in the selected language: switching
	// days never makes the taskbar companion jump or clips 'Mercury'.
	dc, _, _ := getDC.Call(a.HWND)
	defer releaseDC.Call(a.HWND, dc)
	width := 0
	for _, d := range a.Data.Days {
		width = max(width, a.textWidth(dc, a.L.DayShort(d), a.Settings.FontSize, true))
	}
	for _, h := range a.Data.Hours {
		width = max(width, a.textWidth(dc, a.L.Content(h.Name), a.Settings.FontSize, true))
	}
	return max(a.S(64), width+a.S(32))
}
func (l Localizer) About(source string) string {
	return l.Text(
		"第一行：曜星；第二行：时辰，文字随五行着色。\n\n悬停：查看工作建议和传统参考。\n单击：固定／收起详情。拖动：调整位置。\n右键：语言、字号、主题、位置、开机启动、退出。\n表格保存后约 5 秒自动重新读取。\n\n语言：跟随系统／简体中文／English，即时切换。\n英文建议来自本地译文，不联网翻译。用户自改内容没有对应译文时保留原文，可编辑 locales/en.json。\n\n按 Windows 本地时间显示，曜星在 00:00 换日。\n子时为 23:00—01:00，不使用真太阳时。\n仅跟随主任务栏，不会为自身挤开系统图标。\n\n数据来源：",
		"Top row: day star. Bottom row: shichen. Text follows the five-element palette.\n\nHover for guidance; click to pin or close; drag to move.\nRight-click for language, font, theme, position, startup and exit.\nWorkbooks reload about 5 seconds after saving.\n\nLanguage: System / Simplified Chinese / English.\nEnglish guidance is translated locally, without a network service. Custom source text without a matching translation stays unchanged; edit locales/en.json to add one.\n\nUses Windows local time. The day star changes at midnight; Zi spans 23:00-01:00. No solar-time correction.\nFollows the primary taskbar and does not move existing tray icons.\n\nData source: ") + l.Content(source) + l.Text("\n纯本地运行，不联网、不上传数据。", "\nRuns offline. No data is uploaded.")
}
func (l Localizer) ColorRules() string {
	return l.Text(
		"木：青绿；火：朱红；土：土黄；金：银白；水：蓝黑。\n\n深色任务栏上，金用银白，水用灰蓝；浅色任务栏使用同色系的深色。\n日曜按火色、月曜按水色，是界面配色约定。\n\n时辰文字按地支本五行着色：\n子亥属水；寅卯属木；巳午属火；申酉属金；辰戌丑未属土。\n\n曜星卡片随七曜切换底板；时辰卡片保留独立山岚底板。\n任务栏按钮不绘制背景、边框或底板。",
		"Wood: green. Fire: vermilion. Earth: ochre. Metal: silver. Water: slate blue.\n\nDark taskbars use lighter text; light taskbars use darker variants. Sun uses Fire and Moon uses Water as interface color conventions.\n\nShichen text follows the branch's primary element, not a secondary hidden stem.\nWater: Zi, Hai. Wood: Yin, Mao. Fire: Si, Wu. Metal: Shen, You. Earth: Chen, Xu, Chou, Wei.\n\nThe top card follows the day star's artwork. The lower card has independent shichen artwork.\nThe taskbar companion has no visible background or border.")
}
func (l Localizer) PixelLabel(n int) string { return fmt.Sprintf(l.Text("%d 像素", "%d px"), n) }
