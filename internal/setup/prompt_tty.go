package setup

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/chzyer/readline"
	"golang.org/x/term"
)

var errPromptCancelled = errors.New("prompt cancelled")

type TerminalPrompter struct {
	basePrompter
	in      *os.File
	outFile *os.File
}

type selectKey int

const (
	selectKeyUnknown selectKey = iota
	selectKeyUp
	selectKeyDown
	selectKeyConfirm
	selectKeyCancel
)

func (p *TerminalPrompter) Select(question string, options []string) (int, error) {
	if len(options) == 0 {
		return 0, fmt.Errorf("at least one option is required")
	}

	list := newSelectList(stringOptionsToSelectItems(options), 5, defaultSelectListTheme())
	return p.runInteractiveSelect(question, list)
}

func (p *TerminalPrompter) Confirm(question string, defaultYes bool) (bool, error) {
	suffix := "[y/N]"
	if defaultYes {
		suffix = "[Y/n]"
	}

	for {
		p.printPromptHeader(fmt.Sprintf("%s %s", question, suffix))
		answer, err := p.readEditableLine("")
		if err != nil {
			return false, err
		}

		if answer == "" {
			return defaultYes, nil
		}

		switch strings.ToLower(answer) {
		case "y", "yes":
			return true, nil
		case "n", "no":
			return false, nil
		}

		fmt.Fprintln(p.outFile, "Please answer yes or no.")
	}
}

func (p *TerminalPrompter) Ask(question, defaultValue string) (string, error) {
	prompt := question
	if defaultValue != "" {
		prompt = fmt.Sprintf("%s [%s]", question, defaultValue)
	}

	p.printPromptHeader(prompt)
	return p.readEditableLine(defaultValue)
}

func (p *TerminalPrompter) runInteractiveSelect(question string, list *selectList) (int, error) {
	oldState, err := term.MakeRaw(int(p.in.Fd()))
	if err != nil {
		options := make([]string, 0, len(list.items))
		for _, item := range list.items {
			options = append(options, item.label)
		}
		return runLineSelect(&p.basePrompter, question, options)
	}
	defer func() {
		_ = term.Restore(int(p.in.Fd()), oldState)
	}()

	p.printPromptHeader(question)

	fmt.Fprint(p.outFile, "\033[?25l")
	defer fmt.Fprint(p.outFile, "\033[?25h")

	previousLines := 0
	render := func(lines []string) {
		previousLines = redrawSelectBlock(p.outFile, lines, previousLines)
	}

	render(p.selectLines(list))

	for {
		key, err := readSelectKey(p.basePrompter.reader)
		if err != nil {
			fmt.Fprint(p.outFile, "\n")
			return 0, err
		}

		switch key {
		case selectKeyUp:
			list.moveUp()
			render(p.selectLines(list))
		case selectKeyDown:
			list.moveDown()
			render(p.selectLines(list))
		case selectKeyConfirm:
			_, index, _ := list.selectedItem()
			render([]string{list.renderSummary(p.terminalWidth())})
			fmt.Fprint(p.outFile, "\n")
			return index, nil
		case selectKeyCancel:
			render([]string{colorDim + "  Cancelled" + colorReset})
			fmt.Fprint(p.outFile, "\n")
			return 0, errPromptCancelled
		}
	}
}

func (p *TerminalPrompter) selectLines(list *selectList) []string {
	return list.render(p.terminalWidth())
}

func (p *TerminalPrompter) readEditableLine(defaultValue string) (string, error) {
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          inputPrefix,
		HistoryLimit:    -1,
		InterruptPrompt: "\n",
		EOFPrompt:       "\n",
		Stdin:           readline.NewCancelableStdin(p.in),
		Stdout:          p.outFile,
		Stderr:          p.outFile,
	})
	if err != nil {
		fmt.Fprint(p.outFile, inputPrefix)
		return p.basePrompter.readLine()
	}
	defer func() {
		_ = rl.Close()
	}()

	var line string
	if defaultValue != "" {
		line, err = rl.ReadlineWithDefault(defaultValue)
	} else {
		line, err = rl.Readline()
	}
	if err != nil {
		return "", normalizeReadlineError(err)
	}

	return line, nil
}

func normalizeReadlineError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, readline.ErrInterrupt) || errors.Is(err, io.EOF) {
		return errPromptCancelled
	}

	return err
}

func (p *TerminalPrompter) terminalWidth() int {
	width, _, err := term.GetSize(int(p.outFile.Fd()))
	if err != nil || width < 20 {
		return 80
	}

	return width
}

func redrawSelectBlock(out *os.File, lines []string, previousLines int) int {
	if previousLines > 1 {
		fmt.Fprintf(out, "\033[%dA", previousLines-1)
	}

	totalLines := max(previousLines, len(lines))
	for index := 0; index < totalLines; index++ {
		fmt.Fprint(out, "\r\033[K")
		if index < len(lines) {
			fmt.Fprint(out, lines[index])
		}
		if index < totalLines-1 {
			fmt.Fprint(out, "\n")
		}
	}

	if extraLines := totalLines - len(lines); extraLines > 0 {
		fmt.Fprintf(out, "\033[%dA\r", extraLines)
	}

	return len(lines)
}

func readSelectKey(reader *bufio.Reader) (selectKey, error) {
	key, err := reader.ReadByte()
	if err != nil {
		return selectKeyUnknown, err
	}

	switch key {
	case 3:
		return selectKeyCancel, nil
	case '\r', '\n':
		return selectKeyConfirm, nil
	case 'k':
		return selectKeyUp, nil
	case 'j':
		return selectKeyDown, nil
	case 27:
		if reader.Buffered() == 0 {
			return selectKeyCancel, nil
		}

		next, err := reader.ReadByte()
		if err != nil {
			return selectKeyCancel, nil
		}

		if next != '[' && next != 'O' {
			return selectKeyCancel, nil
		}

		if reader.Buffered() == 0 {
			return selectKeyCancel, nil
		}

		final, err := reader.ReadByte()
		if err != nil {
			return selectKeyCancel, nil
		}

		switch final {
		case 'A':
			return selectKeyUp, nil
		case 'B':
			return selectKeyDown, nil
		default:
			return selectKeyCancel, nil
		}
	default:
		return selectKeyUnknown, nil
	}
}

func stringOptionsToSelectItems(options []string) []selectItem {
	items := make([]selectItem, 0, len(options))
	for _, option := range options {
		items = append(items, selectItem{label: option})
	}
	return items
}
