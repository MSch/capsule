package setup

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

const inputPrefix = "→ "

type Prompter interface {
	Select(question string, options []string) (int, error)
	Confirm(question string, defaultYes bool) (bool, error)
	Ask(question, defaultValue string) (string, error)
}

type basePrompter struct {
	reader      *bufio.Reader
	out         io.Writer
	promptCount int
}

type ConsolePrompter struct {
	basePrompter
}

func NewConsolePrompter(in io.Reader, out io.Writer) Prompter {
	if prompt, ok := newTerminalPrompter(in, out); ok {
		return prompt
	}

	return &ConsolePrompter{
		basePrompter: newBasePrompter(in, out),
	}
}

func newBasePrompter(in io.Reader, out io.Writer) basePrompter {
	return basePrompter{
		reader: bufio.NewReader(in),
		out:    out,
	}
}

func (p *ConsolePrompter) Select(question string, options []string) (int, error) {
	return runLineSelect(&p.basePrompter, question, options)
}

func (p *ConsolePrompter) Confirm(question string, defaultYes bool) (bool, error) {
	return runConfirm(&p.basePrompter, question, defaultYes)
}

func (p *ConsolePrompter) Ask(question, defaultValue string) (string, error) {
	return runAsk(&p.basePrompter, question, defaultValue)
}

func runLineSelect(p *basePrompter, question string, options []string) (int, error) {
	p.printPromptHeader(question)
	for index, option := range options {
		fmt.Fprintf(p.out, "  %d. %s\n", index+1, option)
	}

	for {
		fmt.Fprintf(p.out, "Choose an option [1-%d]\n%s", len(options), inputPrefix)
		answer, err := p.readLine()
		if err != nil {
			return 0, err
		}

		switch answer {
		case "1":
			return 0, nil
		case "2":
			if len(options) >= 2 {
				return 1, nil
			}
		}

		fmt.Fprintln(p.out, "Please enter one of the listed numbers.")
	}
}

func runConfirm(p *basePrompter, question string, defaultYes bool) (bool, error) {
	suffix := "[y/N]"
	if defaultYes {
		suffix = "[Y/n]"
	}

	for {
		p.printPromptHeader(fmt.Sprintf("%s %s", question, suffix))
		fmt.Fprint(p.out, inputPrefix)
		answer, err := p.readLine()
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

		fmt.Fprintln(p.out, "Please answer yes or no.")
	}
}

func runAsk(p *basePrompter, question, defaultValue string) (string, error) {
	prompt := question
	if defaultValue != "" {
		prompt = fmt.Sprintf("%s [%s]", question, defaultValue)
	}

	p.printPromptHeader(prompt)
	fmt.Fprint(p.out, inputPrefix)
	answer, err := p.readLine()
	if err != nil {
		return "", err
	}

	if answer == "" {
		return defaultValue, nil
	}

	return answer, nil
}

func (p *basePrompter) readLine() (string, error) {
	line, err := p.reader.ReadString('\n')
	if err != nil && len(line) == 0 {
		return "", err
	}

	return strings.TrimSpace(line), nil
}

func (p *basePrompter) printPromptHeader(prompt string) {
	if p.promptCount > 0 {
		fmt.Fprintln(p.out)
	}

	fmt.Fprintln(p.out, prompt)
	p.promptCount++
}

func newTerminalPrompter(in io.Reader, out io.Writer) (*TerminalPrompter, bool) {
	inFile, ok := in.(*os.File)
	if !ok || !term.IsTerminal(int(inFile.Fd())) {
		return nil, false
	}

	outFile, ok := out.(*os.File)
	if !ok || !term.IsTerminal(int(outFile.Fd())) {
		return nil, false
	}

	return &TerminalPrompter{
		basePrompter: newBasePrompter(inFile, outFile),
		in:           inFile,
		outFile:      outFile,
	}, true
}
