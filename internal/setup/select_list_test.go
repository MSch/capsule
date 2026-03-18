package setup

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
)

func TestSelectListWrapsSelection(t *testing.T) {
	t.Parallel()

	list := newSelectList([]selectItem{
		{label: "First"},
		{label: "Second"},
		{label: "Third"},
	}, 5, plainSelectListTheme())

	list.moveUp()
	_, index, ok := list.selectedItem()
	if !ok {
		t.Fatal("expected a selected item")
	}
	if index != 2 {
		t.Fatalf("expected selection to wrap to last item, got %d", index)
	}

	list.moveDown()
	_, index, _ = list.selectedItem()
	if index != 0 {
		t.Fatalf("expected selection to wrap back to first item, got %d", index)
	}
}

func TestSelectListRendersDescriptionColumn(t *testing.T) {
	t.Parallel()

	list := newSelectList([]selectItem{
		{label: "Install locally", description: "Use the current machine"},
		{label: "Connect over SSH", description: "Bootstrap a remote server"},
	}, 5, plainSelectListTheme())

	lines := list.render(80)
	if len(lines) < 2 {
		t.Fatalf("expected at least two rendered lines, got %d", len(lines))
	}

	if !strings.Contains(lines[0], "→ Install locally") {
		t.Fatalf("expected selected line to include the highlighted label, got %q", lines[0])
	}
	if !strings.Contains(lines[0], "Use the current machine") {
		t.Fatalf("expected selected line to include the description, got %q", lines[0])
	}
	if !strings.Contains(lines[1], "Bootstrap a remote server") {
		t.Fatalf("expected secondary line to include its description, got %q", lines[1])
	}
}

func TestSelectListAddsScrollInfo(t *testing.T) {
	t.Parallel()

	items := make([]selectItem, 0, 8)
	for index := 0; index < 8; index++ {
		items = append(items, selectItem{label: "Option"})
	}

	list := newSelectList(items, 5, plainSelectListTheme())
	for range 6 {
		list.moveDown()
	}

	lines := list.render(60)
	lastLine := lines[len(lines)-1]
	if lastLine != "  (7/8)" {
		t.Fatalf("expected scroll info for the current position, got %q", lastLine)
	}
}

func TestReadSelectKey(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input []byte
		want  selectKey
	}{
		{name: "vim up", input: []byte("k"), want: selectKeyUp},
		{name: "vim down", input: []byte("j"), want: selectKeyDown},
		{name: "enter", input: []byte{'\r'}, want: selectKeyConfirm},
		{name: "escape", input: []byte{27}, want: selectKeyCancel},
		{name: "arrow up", input: []byte{27, '[', 'A'}, want: selectKeyUp},
		{name: "arrow down", input: []byte{27, '[', 'B'}, want: selectKeyDown},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			key, err := readSelectKey(bufio.NewReader(bytes.NewReader(testCase.input)))
			if err != nil {
				t.Fatalf("readSelectKey returned an error: %v", err)
			}
			if key != testCase.want {
				t.Fatalf("expected key %v, got %v", testCase.want, key)
			}
		})
	}
}

func plainSelectListTheme() selectListTheme {
	return selectListTheme{
		selectedLine: func(text string) string { return text },
		description:  func(text string) string { return text },
		scrollInfo:   func(text string) string { return text },
		noMatch:      func(text string) string { return text },
	}
}
