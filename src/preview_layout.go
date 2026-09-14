package main

import "strings"

// A keyword is deliberately a preview, not a replacement for the source. Long
// custom text gets an ellipsis; clicking that row reveals the full current cell.
func fitPreviewText(text string, width, size int, bold bool, measure MeasureText) string {
	if measure(text, size, bold) <= width {
		return text
	}
	if measure("…", size, bold) > width {
		return ""
	}
	r := []rune(text)
	lo, hi := 0, len(r)
	for lo < hi {
		mid := (lo + hi + 1) / 2
		if measure(strings.TrimSpace(string(r[:mid]))+"…", size, bold) <= width {
			lo = mid
		} else {
			hi = mid - 1
		}
	}
	return strings.TrimSpace(string(r[:lo])) + "…"
}

// Drawers live INSIDE their parent cards, not in a third card or a footer. The
// label stays visible when closed, but no future identity/advice leaks through.
func appendPreviewDrawer(out *CardLayout, kind string, items []PreviewItem, opt CardOptions, x, y, width int, measure MeasureText) int {
	if len(items) == 0 {
		return y
	}
	dpi := opt.DPI
	if dpi < 1 {
		dpi = 96
	}
	s := func(n int) int { return (n*dpi + 48) / 96 }
	c := darkPalette
	if opt.Light {
		c = lightPalette
	}
	id, rowBase, open := previewDaysButton, previewDayRowBase, opt.Preview.Days
	title := opt.Locale.Text("未来 3 日", "Next 3 days")
	if kind == "hour" {
		id, rowBase, open = previewHoursButton, previewHourRowBase, opt.Preview.Hours
		title = opt.Locale.Text("后续时辰", "Next 4 shichen")
	}
	block := func(text string, rect Rect, size int, bold bool, col uint32) {
		out.Blocks = append(out.Blocks, TextBlock{text, rect, size, bold, col})
	}
	y += s(6)
	out.Panels = append(out.Panels, VisualPanel{R(x, y, width, max(1, s(1))), "preview-divider", 0})
	y += s(3)
	header := R(x, y, width, s(26))
	out.Buttons = append(out.Buttons, Button{true, id, title, header})
	block(title, R(x, y+s(4), width-s(80), s(19)), 11, false, c.Muted)
	action := opt.Locale.Text("展开", "Show")
	if open {
		action = opt.Locale.Text("收起", "Hide")
	}
	aw := measure(action, 10, false) + s(2)
	block(action, R(x+width-s(21)-aw, y+s(5), aw, s(18)), 10, false, c.Muted)
	y += s(28)
	if !open {
		return y
	}
	for i, item := range items {
		top := y
		tx, right := x+s(24), x+width-s(5)
		textW := right - tx
		nameW := measure(item.Name, 12, true) + s(2)
		keywordX := tx + nameW + s(10)
		keywordW := max(1, right-keywordX)
		out.Icons = append(out.Icons, VisualIcon{R(x+s(3), y+s(5), s(16), s(16)), assetName(kind, item.Index, opt.Light, false)})
		block(item.Name, R(tx, y+s(2), nameW, s(19)), 12, true, item.Color)
		keyword := fitPreviewText(item.Keyword, keywordW, 11, false, measure)
		block(keyword, R(keywordX, y+s(3), keywordW, s(18)), 11, false, c.Text)
		y += s(21)
		// Priorities are rendered only when they exist in the source; lunch/sleep
		// never receive invented P-levels. Reserve a separate column for real values.
		priorityW := 0
		if item.Priority != "" {
			priorityW = min(textW/3, measure(item.Priority, 10, false)+s(3))
		}
		whenW := textW
		if priorityW > 0 {
			whenW -= priorityW + s(8)
		}
		whens := WrapText(item.When, whenW, 10, false, measure)
		for j, when := range whens {
			block(when, R(tx, y+j*s(15), whenW, s(16)), 10, false, c.Muted)
		}
		if priorityW > 0 {
			block(fitPreviewText(item.Priority, priorityW, 10, false, measure), R(right-priorityW, y, priorityW, s(16)), 10, false, item.Color)
		}
		y += len(whens)*s(15) + s(5)
		button := Button{true, rowBase + i + 1, item.Name, R(x, top, width, y-top)}
		out.Buttons = append(out.Buttons, button)
		if i == 0 || opt.Preview.Open(button.ID) {
			out.Panels = append(out.Panels, VisualPanel{button.Rect, "preview-" + kind + "-row", item.Index})
		}
		if opt.Preview.Open(button.ID) {
			y += s(2)
			prefix := opt.Locale.Text("宜：", "Best for: ")
			for _, text := range WrapText(prefix+item.Advice, textW, 11, false, measure) {
				block(text, R(tx, y, textW, s(19)), 11, false, c.Text)
				y += s(19)
			}
			y += s(6)
		}
		y += s(2)
	}
	return y + s(2)
}
