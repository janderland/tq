package state

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/chzyer/readline"
)

type Task struct {
	Title   string
	Created time.Time
}

// Normalize cleans up the title of the task. Leading
// & trailing whitespace is removed and any substring of
// adjacent whitespace is replaced by a single space. The
// title is also changed to all caps. If the resulting
// title is an empty string, an error is returned.
func (t *Task) Normalize() error {
	title := strings.ToUpper(trim(t.Title))
	if title == "" {
		return fmt.Errorf("title is empty")
	}
	t.Title = title
	return nil
}

// Edit lets the user edit the task's title inline at the terminal
// prompt, pre-filled with the current title and using vim-style
// modal keybindings for navigation.
func (t *Task) Edit() error {
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

	title, err := rl.ReadlineWithDefault(t.Title)
	if err != nil {
		return fmt.Errorf("%w: failed to read edited title", err)
	}

	t.Title = title
	return nil
}

// trim transforms any adjacent whitespace into a single
// space and removes any leading or trailing whitespace.
func trim(str string) string {
	return regexp.MustCompile(`\s+`).ReplaceAllString(strings.TrimSpace(str), " ")
}
