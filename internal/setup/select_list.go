package setup

import (
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	defaultPrimaryColumnWidth = 32
	primaryColumnGap          = 2
	minDescriptionWidth       = 10
)

const colorDim = "\033[2m"

type selectItem struct {
	label       string
	description string
}

type selectListTheme struct {
	selectedLine func(string) string
	description  func(string) string
	scrollInfo   func(string) string
	noMatch      func(string) string
}

type selectList struct {
	items         []selectItem
	selectedIndex int
	maxVisible    int
	theme         selectListTheme
}

func newSelectList(items []selectItem, maxVisible int, theme selectListTheme) *selectList {
	if maxVisible < 1 {
		maxVisible = 5
	}

	return &selectList{
		items:      append([]selectItem(nil), items...),
		maxVisible: maxVisible,
		theme:      theme,
	}
}

func defaultSelectListTheme() selectListTheme {
	return selectListTheme{
		selectedLine: func(text string) string {
			return colorCyan + text + colorReset
		},
		description: func(text string) string {
			return colorDim + text + colorReset
		},
		scrollInfo: func(text string) string {
			return colorDim + text + colorReset
		},
		noMatch: func(text string) string {
			return colorDim + text + colorReset
		},
	}
}

func (l *selectList) moveUp() {
	if len(l.items) == 0 {
		return
	}

	if l.selectedIndex == 0 {
		l.selectedIndex = len(l.items) - 1
		return
	}

	l.selectedIndex--
}

func (l *selectList) moveDown() {
	if len(l.items) == 0 {
		return
	}

	if l.selectedIndex == len(l.items)-1 {
		l.selectedIndex = 0
		return
	}

	l.selectedIndex++
}

func (l *selectList) selectedItem() (selectItem, int, bool) {
	if len(l.items) == 0 {
		return selectItem{}, 0, false
	}

	return l.items[l.selectedIndex], l.selectedIndex, true
}

func (l *selectList) render(width int) []string {
	if len(l.items) == 0 {
		return []string{l.theme.noMatch("  No options available")}
	}

	if width < 20 {
		width = 20
	}

	lines := make([]string, 0, l.maxVisible+1)
	start, end := l.visibleRange()
	primaryWidth := l.primaryColumnWidth()

	for index := start; index < end; index++ {
		item := l.items[index]
		lines = append(lines, l.renderItem(item, index == l.selectedIndex, width, primaryWidth))
	}

	if start > 0 || end < len(l.items) {
		lines = append(lines, l.theme.scrollInfo(truncateWidth("  ("+itoa(l.selectedIndex+1)+"/"+itoa(len(l.items))+")", width-2)))
	}

	return lines
}

func (l *selectList) renderSummary(width int) string {
	item, _, ok := l.selectedItem()
	if !ok {
		return ""
	}

	prefix := "→ "
	if width < 10 {
		width = 10
	}

	summary := prefix + truncateWidth(item.label, width-stringWidth(prefix))
	return l.theme.selectedLine(summary)
}

func (l *selectList) visibleRange() (int, int) {
	if len(l.items) <= l.maxVisible {
		return 0, len(l.items)
	}

	start := l.selectedIndex - l.maxVisible/2
	if start < 0 {
		start = 0
	}

	maxStart := len(l.items) - l.maxVisible
	if start > maxStart {
		start = maxStart
	}

	return start, min(start+l.maxVisible, len(l.items))
}

func (l *selectList) primaryColumnWidth() int {
	widest := 0
	for _, item := range l.items {
		widest = max(widest, stringWidth(item.label)+primaryColumnGap)
	}

	return clamp(widest, 1, defaultPrimaryColumnWidth)
}

func (l *selectList) renderItem(item selectItem, selected bool, width, primaryWidth int) string {
	prefix := "  "
	if selected {
		prefix = "→ "
	}

	prefixWidth := stringWidth(prefix)
	label := truncateWidth(item.label, max(1, width-prefixWidth-2))
	description := normalizeSingleLine(item.description)

	if description != "" && width > 40 {
		effectivePrimaryWidth := max(1, min(primaryWidth, width-prefixWidth-4))
		maxLabelWidth := max(1, effectivePrimaryWidth-primaryColumnGap)
		label = truncateWidth(item.label, maxLabelWidth)
		labelWidth := stringWidth(label)
		spacing := strings.Repeat(" ", max(1, effectivePrimaryWidth-labelWidth))
		remainingWidth := width - prefixWidth - labelWidth - len(spacing) - 2
		if remainingWidth > minDescriptionWidth {
			desc := truncateWidth(description, remainingWidth)
			if selected {
				return l.theme.selectedLine(prefix + label + spacing + desc)
			}

			return prefix + label + l.theme.description(spacing+desc)
		}
	}

	if selected {
		return l.theme.selectedLine(prefix + label)
	}

	return prefix + label
}

func normalizeSingleLine(text string) string {
	text = strings.ReplaceAll(text, "\r", " ")
	text = strings.ReplaceAll(text, "\n", " ")
	return strings.TrimSpace(text)
}

func truncateWidth(text string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}

	if stringWidth(text) <= maxWidth {
		return text
	}

	var (
		builder strings.Builder
		width   int
	)
	for _, r := range text {
		if width+1 > maxWidth {
			break
		}
		builder.WriteRune(r)
		width++
	}

	return builder.String()
}

func stringWidth(text string) int {
	return utf8.RuneCountInString(text)
}

func clamp(value, minValue, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func itoa(value int) string {
	return strconv.Itoa(value)
}
