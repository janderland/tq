package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/chzyer/readline"
	"github.com/janderland/tq/state"
)

// UI provides a uniform set of IO functions. Ensures
// there is an empty line between every statement.
type UI struct {
	nl    bool
	width int
}

func New(width int) UI {
	return UI{width: width}
}

// QueryYesNo queries the user for a 'y' or 'n'. The response is
// read as soon as a single character is pressed, without waiting
// for enter. If the user presses another key, the prompt is
// repeated. If the user presses 'y' or 'n', then true or false
// is returned respectively.
func (u *UI) QueryYesNo() (bool, error) {
	u.newline()
	fmt.Print("Enter y|n: ")

	fd := int(os.Stdin.Fd())
	termState, err := readline.MakeRaw(fd)
	if err != nil {
		return false, fmt.Errorf("%w: failed to query user", err)
	}
	defer readline.Restore(fd, termState)

	key := make([]byte, 1)
	for {
		if _, err := os.Stdin.Read(key); err != nil {
			return false, fmt.Errorf("%w: failed to query user", err)
		}
		switch key[0] {
		case 'y':
			fmt.Println("y")
			return true, nil
		case 'n':
			fmt.Println("n")
			return false, nil
		case 3: // Ctrl+C
			fmt.Println()
			os.Exit(130)
		}
	}
}

// Message prints a message to the user.
func (u *UI) Message(format string, args ...any) {
	u.newline()
	fmt.Println(paragraph(fmt.Sprintf("+ "+format, args...), u.width, 2))
}

// Display prints the task found at the given index in the given TaskQueue.
func (u *UI) Display(tasks state.TaskQueue, index int) {
	header := fmt.Sprintf("%d. ", index)
	if index < 10 {
		header += " "
	}
	if index <= tasks.LastOpenedIndex() {
		header += "[" + green("open") + "] "
	} else {
		header += "[" + yellow("todo") + "] "
	}
	header += tasks.At(index).Created.Format("2006-01-02")
	title := strings.Repeat(" ", 4) + bold(tasks.At(index).Title)

	u.newline()
	fmt.Println(paragraph(header, u.width, 4))
	fmt.Println(paragraph(title, u.width, 4))
}

// DisplayDone prints the done task found at the given index.
func (u *UI) DisplayDone(tasks state.TaskQueue, index int) {
	header := "x.  [" + red("done") + "] "
	header += tasks.DoneAt(index).Created.Format("2006-01-02")
	title := strings.Repeat(" ", 4) + bold(tasks.DoneAt(index).Title)

	u.newline()
	fmt.Println(paragraph(header, u.width, 4))
	fmt.Println(paragraph(title, u.width, 4))
}

// Line prints a horizontal separator.
func (u *UI) Line() {
	u.newline()
	fmt.Println("---")
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
// are added between words after a line becomes "width" wide.
// Also, each line (except for the first) is prefixed with
// "indent" number of spaces.
func paragraph(str string, width, indent int) string {
	wordList := strings.Split(str, " ")
	count := 0
	str = ""
	for i, word := range wordList {
		str += word
		count += len(word)

		if i == len(wordList)-1 {
			break
		}

		if count > width {
			str += "\n" + strings.Repeat(" ", indent)
			count = indent
			continue
		}

		str += " "
		count++
	}
	return str
}

// bold wraps a string in the ANSI escape codes
// which render it in bold on the terminal.
func bold(str string) string {
	return "\033[1m" + str + "\033[0m"
}

// green wraps a string in the ANSI escape codes
// which render it in green on the terminal.
func green(str string) string {
	return "\033[32m" + str + "\033[0m"
}

// yellow wraps a string in the ANSI escape codes
// which render it in yellow on the terminal.
func yellow(str string) string {
	return "\033[33m" + str + "\033[0m"
}

// red wraps a string in the ANSI escape codes
// which render it in red on the terminal.
func red(str string) string {
	return "\033[31m" + str + "\033[0m"
}
