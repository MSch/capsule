package setup

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type Prompter interface {
	Select(question string, options []string) (int, error)
	Confirm(question string, defaultYes bool) (bool, error)
	Ask(question, defaultValue string) (string, error)
}

type ConsolePrompter struct {
	reader *bufio.Reader
	out    io.Writer
}

func NewConsolePrompter(in io.Reader, out io.Writer) Prompter {
	return &ConsolePrompter{
		reader: bufio.NewReader(in),
		out:    out,
	}
}

func (p *ConsolePrompter) Select(question string, options []string) (int, error) {
	fmt.Fprintln(p.out, question)
	for index, option := range options {
		fmt.Fprintf(p.out, "  %d. %s\n", index+1, option)
	}

	for {
		fmt.Fprintf(p.out, "Choose an option [1-%d]: ", len(options))
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

func (p *ConsolePrompter) Confirm(question string, defaultYes bool) (bool, error) {
	suffix := "[y/N]"
	if defaultYes {
		suffix = "[Y/n]"
	}

	for {
		fmt.Fprintf(p.out, "%s %s: ", question, suffix)
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

func (p *ConsolePrompter) Ask(question, defaultValue string) (string, error) {
	prompt := question
	if defaultValue != "" {
		prompt = fmt.Sprintf("%s [%s]", question, defaultValue)
	}

	fmt.Fprintf(p.out, "%s: ", prompt)
	answer, err := p.readLine()
	if err != nil {
		return "", err
	}

	if answer == "" {
		return defaultValue, nil
	}

	return answer, nil
}

func (p *ConsolePrompter) readLine() (string, error) {
	line, err := p.reader.ReadString('\n')
	if err != nil && len(line) == 0 {
		return "", err
	}

	return strings.TrimSpace(line), nil
}
