package state

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
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

// Edit writes the task's title to a temporary file and opens the file
// in a text editor. If the editor exits with 0 exit code, then the
// first line of the temporary file is parsed back into the task's
// title. If the editor fails, the task remains unedited. The editor
// is run in the shell specified with the environment variable $SHELL
// (defaulting to "sh"). The editor is specified with the environment
// variable $EDITOR (defaulting to "vim").
func (t *Task) Edit() error {
	namePattern := "*_" + strings.ReplaceAll(t.Title, " ", "_")
	file, err := os.CreateTemp("", namePattern)
	if err != nil {
		return fmt.Errorf("%w: failed to open temp file", err)
	}

	_, err = fmt.Fprint(file, t.Title)
	if err != nil {
		return fmt.Errorf("%w: failed to write task", err)
	}
	if err = file.Close(); err != nil {
		return fmt.Errorf("%w: failed to close file", err)
	}

	shell := strings.TrimSpace(os.Getenv("SHELL"))
	if len(shell) == 0 {
		shell = "sh"
	}
	editor := strings.TrimSpace(os.Getenv("EDITOR"))
	if len(editor) == 0 {
		editor = "vim"
	}
	cmd := exec.Command(shell, "-c", fmt.Sprintf("%s %s", editor, file.Name()))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	if err = cmd.Run(); err != nil {
		return fmt.Errorf("%w: failed to execute editor", err)
	}

	file, err = os.Open(file.Name())
	if err != nil {
		return fmt.Errorf("%w: failed to open file", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lines := []string{}
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err = scanner.Err(); err != nil {
		return fmt.Errorf("%w: failed to read file", err)
	}

	if len(lines) > 0 {
		t.Title = lines[0]
	}

	return nil
}

// trim transforms any adjacent whitespace into a single
// space and removes any leading or trailing whitespace.
func trim(str string) string {
	return regexp.MustCompile(`\s+`).ReplaceAllString(strings.TrimSpace(str), " ")
}
