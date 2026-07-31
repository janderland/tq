package ui

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"unicode"

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
// is returned respectively. If the user presses Ctrl-C, or the
// process is asked to terminate while waiting on input, an error
// is returned and the terminal's state is always restored first.
func (u *UI) QueryYesNo() (bool, error) {
	u.newline()
	fmt.Print("Enter y|n: ")

	fd := int(os.Stdin.Fd())
	termState, err := readline.MakeRaw(fd)
	if err != nil {
		return false, fmt.Errorf("%w: failed to query user", err)
	}
	defer readline.Restore(fd, termState)

	// Raw mode disables the terminal's own Ctrl-C handling, so a
	// keyboard interrupt never reaches us as a signal here - it's
	// caught as a plain byte below instead. This handler is for
	// termination requests that arrive some other way (e.g. `kill`,
	// or the terminal closing), which would otherwise skip the
	// deferred restore above and leave the terminal in raw mode.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	done := make(chan struct{})
	defer func() {
		signal.Stop(sigCh)
		close(done)
	}()
	go func() {
		select {
		case <-sigCh:
			readline.Restore(fd, termState)
			os.Exit(1)
		case <-done:
		}
	}()

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
		case readline.CharInterrupt:
			fmt.Println()
			return false, fmt.Errorf("cancelled by user")
		}
	}
}

// Edit lets the user edit the given task's title inline at the
// terminal prompt, pre-filled with the current title and using
// vim-style modal keybindings for navigation. The task is
// edited in-place.
func (u *UI) Edit(task *state.Task) error {
	rl, err := readline.NewEx(&readline.Config{
		Prompt:  "Task: ",
		VimMode: true,
		FuncFilterInputRune: func(r rune) (rune, bool) {
			return unicode.ToUpper(r), true
		},
	})
	if err != nil {
		return fmt.Errorf("%w: failed to start line editor", err)
	}
	defer rl.Close()

	title, err := rl.ReadlineWithDefault(task.Title)
	if err != nil {
		return fmt.Errorf("%w: failed to read edited title", err)
	}

	task.Title = title
	return nil
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
