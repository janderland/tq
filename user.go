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
	rd *bufio.Reader
	nl bool

	Width int
}

func (u *UI) reader() *bufio.Reader {
	if u.rd == nil {
		u.rd = bufio.NewReader(os.Stdin)
	}
	return u.rd
}

func (u *UI) newline() {
	if u.nl {
		fmt.Println()
	}
	u.nl = true
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
	fmt.Printf("+ "+format+"\n", args...)
}

func (u *UI) display(tasks []*Task, index int) {
	task := tasks[index]
	title := fmt.Sprintf("%d. ", index)
	if index < 10 {
		title += " "
	}
	count := len(title)
	titleWords := append(
		[]string{fmt.Sprintf("[%s]", task.State)},
		strings.Split(task.Title, " ")...)
	for i, word := range titleWords {
		title += word
		count += len(word)
		if i != len(titleWords)-1 {
			if count > u.Width {
				title += "\n    "
				count = 4
			} else {
				title += " "
				count++
			}
		}
	}

	story := "    "
	count = len(story)
	storyWords := strings.Split(task.Story, " ")
	for i, word := range storyWords {
		story += word
		count += len(word)
		if i != len(storyWords)-1 {
			if count > u.Width {
				story += "\n    "
				count = 4
			} else {
				story += " "
				count++
			}
		}
	}

	u.newline()
	fmt.Println(title)
	fmt.Println(story)
}
