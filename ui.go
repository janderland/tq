package main

import (
	"bufio"
	"fmt"
	"github.com/pkg/errors"
	"os"
	"strings"
)

// Provides a uniform set of UI functions. Ensures
// there is an empty line between every statement.
type UI struct {
	rd    *bufio.Reader
	nl    bool
	width int
}

func (u *UI) reader() *bufio.Reader {
	if u.rd == nil {
		u.rd = bufio.NewReader(os.Stdin)
	}
	return u.rd
}

// Prints a newline, if necessary,
// between each user interaction.
func (u *UI) newline() {
	if u.nl {
		fmt.Println()
	}
	u.nl = true
}

// Formats a string to be printed to the console. Newlines
// are added between words where needed to ensure no single
// line exceeds "u.width" characters. Also, each line
// (except for the first) is prefixed with "indent" number
// of spaces.
func (u *UI) paragraph(str string, indent int) string {
	wordList := strings.Split(str, " ")
	count := 0
	str = ""
	for i, word := range wordList {
		str += word
		count += len(word)
		if i != len(wordList)-1 {
			if count > u.width {
				str += "\n" + spaces(indent)
				count = indent
			} else {
				str += " "
				count++
			}
		}
	}
	return str
}

func (u *UI) queryYesNo() (bool, error) {
	u.newline()
	for {
		fmt.Print("Enter y|n: ")
		resp, err := u.reader().ReadString('\n')
		if err != nil {
			return false, errors.Wrap(err, "failed to query user")
		}
		switch strings.TrimSpace(resp) {
		case "y":
			return true, nil
		case "n":
			return false, nil
		}
	}
}

func (u *UI) queryLine() (string, error) {
	u.newline()
	return u.reader().ReadString('\n')
}

func (u *UI) message(format string, args ...interface{}) {
	u.newline()
	fmt.Println(u.paragraph(fmt.Sprintf("+ "+format, args...), 2))
}

func (u *UI) display(tasks TaskQueue, index int) {
	title := fmt.Sprintf("%d. ", index)
	if index < 10 {
		title += " "
	}
	if index <= tasks.lastOpenedIndex() {
		title += "[open] "
	} else {
		title += "[todo] "
	}
	title += tasks.at(index).Title
	story := spaces(4) + tasks.at(index).Story
	u.newline()
	fmt.Println(u.paragraph(title, 4))
	fmt.Println(u.paragraph(story, 4))
}

func (u *UI) line() {
	u.newline()
	fmt.Println("---")
}
