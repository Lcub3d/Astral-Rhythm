package main

// These are interface color assignments, not replacements for the workbook's
// own traditional descriptions. Sun follows Fire; Moon follows Water.
var dayElements = [7]string{"火", "水", "火", "水", "木", "金", "土"}
var hourElements = [12]string{"水", "土", "木", "木", "土", "火", "火", "土", "金", "金", "土", "水"}
var branchAnimals = [12]string{"鼠", "牛", "虎", "兔", "龙", "蛇", "马", "羊", "猴", "鸡", "狗", "猪"}

func DayElement(i int) string {
	if i < 0 || i >= len(dayElements) {
		return ""
	}
	return dayElements[i]
}
func HourElement(i int) string {
	if i < 0 || i >= len(hourElements) {
		return ""
	}
	return hourElements[i]
}
func BranchLabel(h Hour) string {
	if h.Index < 0 || h.Index >= len(branchAnimals) {
		return h.Name
	}
	r := []rune(h.Name)
	if len(r) == 0 {
		return ""
	}
	return string(r[0]) + "（" + branchAnimals[h.Index] + "） · " + HourElement(h.Index)
}
func DayTextColor(i int, light bool) uint32 {
	if i < 0 || i >= len(visualDesign.Days) {
		return 0xBFCBCD
	}
	return visualDesign.Days[i].Accent(light)
}
func HourTextColor(i int, light bool) uint32 {
	if i < 0 || i >= len(visualDesign.Hours) {
		return 0xBFCBCD
	}
	return visualDesign.Hours[i].Accent(light)
}

func blendRGB(a, b uint32, t float64) uint32 {
	var result uint32
	for _, shift := range []uint{0, 8, 16} {
		v := float64((a>>shift)&255)*(1-t) + float64((b>>shift)&255)*t
		result |= uint32(v+.5) << shift
	}
	return result
}
