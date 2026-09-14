//go:build windows

package main

import (
	"fmt"
	"strings"
	"time"
)

var previewGetAsyncKeyState = user32.NewProc("GetAsyncKeyState")

type previewInput struct{ Left, Right, Escape bool }

func (a *App) applyPreviewOptions(o *CardOptions, now time.Time) {
	// Keep an opened drawer across clock/theme refreshes, but don't silently
	// retarget an expanded row when the relative day/hour positions advance.
	key := fmt.Sprintf("%s/%d", now.Format("2006-01-02"), HourIndex(now))
	if a.PreviewClock != key {
		a.Preview.DayRow, a.Preview.HourRow = 0, 0
		a.PreviewClock = key
	}
	o.Preview = a.Preview
	o.UpcomingDays = UpcomingDays(a.Data, now, a.L, a.isLight())
	o.UpcomingHours = UpcomingHours(a.Data, now, a.L, a.isLight())
}

// Only the instantaneous left/right/Escape states needed to dismiss our popup
// are sampled. There are no hooks, key recording, background logs or activation
// of the main window; opening a drawer must not steal typing focus.
func (a *App) pollPreviewDismissal() {
	down := func(key uintptr) bool { v, _, _ := previewGetAsyncKeyState.Call(key); return v&0x8000 != 0 }
	in := previewInput{down(1), down(2), down(27)}
	before := a.PreviewInput
	a.PreviewInput = in
	if !a.CardShown || a.MenuOpen || a.MouseDown {
		return
	}
	if in.Escape && !before.Escape {
		a.hideCard()
		a.HoverSince = time.Now().Add(24 * time.Hour)
		return
	}
	clicked := in.Left && !before.Left || in.Right && !before.Right
	if clicked && a.Preview.Active() && !a.Pinned {
		p := cursorPoint()
		if !a.WidgetRect.Contains(p) && !windowRect(a.Popup).Contains(p) {
			a.hideCard()
		}
	}
}

func (a *App) paintPreviewPanel(dst *PixelBuffer, panel VisualPanel, rect, clip Rect) {
	if panel.Kind == "preview-divider" {
		dst.RoundRect(rect, 0, blendRGB(a.Colors.Border, a.Colors.Panel, .35), clip)
		return
	}
	if panel.Kind != "preview-day-row" && panel.Kind != "preview-hour-row" {
		return
	}
	col := DayTextColor(panel.Index, a.isLight())
	kind := "day"
	if panel.Kind == "preview-hour-row" {
		col = HourTextColor(panel.Index, a.isLight())
		kind = "hour"
	}
	surface := a.Colors.Panel
	for _, p := range a.Panels {
		if p.Kind == kind {
			if kind == "day" {
				surface = visualDesign.Days[p.Index].Surface(a.isLight())
			} else {
				surface = visualDesign.Hours[p.Index].Surface(a.isLight())
			}
			break
		}
	}
	dst.RoundRect(rect, float64(a.S(5)), blendRGB(col, surface, .95), clip)
	// Nearest/selected rows use a hairline, never a saturated full-width plate.
	dst.RoundRect(R(int(rect.Left), int(rect.Top)+a.S(6), max(1, a.S(2)), rect.H()-a.S(12)), float64(a.S(1)), blendRGB(col, surface, .2), clip)
}

func isPreviewButton(id int) bool {
	return id == previewDaysButton || id == previewHoursButton || id > previewDayRowBase && id <= previewDayRowBase+3 || id > previewHourRowBase && id <= previewHourRowBase+4
}
func (a *App) cardHit(p Point) int {
	for _, b := range a.Buttons {
		if (!b.InBody || a.Body.Contains(p)) && a.buttonRect(b).Contains(p) {
			return b.ID
		}
	}
	return 0
}
func (a *App) previewControlCursor() bool {
	p := cursorPoint()
	wr := windowRect(a.Popup)
	p.X -= wr.Left
	p.Y -= wr.Top
	if a.cardHit(p) == 0 {
		return false
	}
	hand, _, _ := loadCursor.Call(0, 32649) // IDC_HAND, system-owned cursor
	setCursor.Call(hand)
	return true
}

// Preview header chevrons remain visible without hovering so the new affordance
// is discoverable. Legacy detail chevrons retain their hover-only appearance.
func (a *App) paintCardControls(dst *PixelBuffer, clip Rect) {
	p := cursorPoint()
	wr := windowRect(a.Popup)
	inside := wr.Contains(p)
	p.X -= wr.Left
	p.Y -= wr.Top
	for _, b := range a.Buttons {
		preview := isPreviewButton(b.ID)
		if !preview && !inside {
			continue
		}
		r := a.buttonRect(b)
		if r.Bottom <= a.Body.Top || r.Top >= a.Body.Bottom {
			continue
		}
		col := a.Colors.Muted
		hovered := inside && r.Contains(p)
		if hovered {
			// Outline only: controls are drawn after text, so a fill would hide it.
			dst.RoundFrame(r, float64(a.S(5)), 1, a.Colors.Border, clip)
			col = a.Colors.Text
		}
		if b.ID > previewDayRowBase {
			continue
		}
		open := a.Settings.HourDetails
		if b.ID == 6 {
			open = a.Settings.DayDetails
		}
		if preview {
			open = a.Preview.Open(b.ID)
		}
		cx, cy := float64(r.Left)+float64(r.W())/2, float64(r.Top)+float64(r.H())/2
		if preview {
			cx = float64(r.Right) - float64(a.S(9))
		}
		d := float64(a.S(3))
		dy := d / 2
		if open {
			dy = -dy
		}
		dst.Line(cx-d, cy-dy, cx, cy+dy, 1, col, clip)
		dst.Line(cx, cy+dy, cx+d, cy-dy, 1, col, clip)
	}
}

// Testable classification; docs and UI never introduce a mixed third card.
func previewPanelKind(kind string) bool { return strings.HasPrefix(kind, "preview-") }
