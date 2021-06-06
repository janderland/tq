package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/janderland/tq/state"
	"github.com/pkg/errors"
)

// UI provides a uniform set of IO functions. Ensures
// there is an empty line between every statement.
type UI struct {
	rd    *bufio.Reader
	nl    bool
	width int
}

func New(width int) UI {
	return UI{width: width}
}

// QueryYesNo queries the user for a 'y' or 'n'.
// If the user enters another character, the prompt
// is repeated. If the user enters 'y' or 'n',
// then true or false is returned respectively.
func (u *UI) QueryYesNo() (bool, error) {
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

// Message prints a message to the user.
func (u *UI) Message(format string, args ...interface{}) {
	u.newline()
	fmt.Println(u.paragraph(fmt.Sprintf("+ "+format, args...), 2))
}

// Display prints the task found at the given index in the given TaskQueue.
func (u *UI) Display(tasks state.TaskQueue, index int) {
	title := fmt.Sprintf("%d. ", index)
	if index < 10 {
		title += " "
	}
	if index <= tasks.LastOpenedIndex() {
		title += "[open] "
	} else {
		title += "[todo] "
	}
	title += tasks.At(index).Title
	story := spaces(4) + tasks.At(index).Story
	u.newline()
	fmt.Println(u.paragraph(title, 4))
	fmt.Println(u.paragraph(story, 4))
}

// Line prints a horizontal separator.
func (u *UI) Line() {
	u.newline()
	fmt.Println("---")
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

func spaces(count int) string {
	str := ""
	for i := 0; i < count; i++ {
		str += " "
	}
	return str
}
