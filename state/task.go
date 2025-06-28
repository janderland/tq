package state

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type Task struct {
	Title string
	Story string
}

// Normalize cleans up the title & story of the
// task. Lead & trailing whitespace is removed
// and any substring of adjacent whitespace is
// replaced by a single space. The title is also
// changed to all caps. After these transformations,
// if either the title or story is an empty string,
// an error is returned.
func (t *Task) Normalize() error {
	title := strings.ToUpper(trim(t.Title))
	story := trim(t.Story)
	if title == "" {
		return fmt.Errorf("title is empty")
	}
	if story == "" {
		return fmt.Errorf("story is empty")
	}
	t.Title = title
	t.Story = story
	return nil
}

// Edit writes the task to a temporary file and opens the file in
// a text editor. The format is: first line is the title, remaining
// lines are the story. If the editor exits with 0 exit code, then the
// temporary file is parsed back into the task. If the editor fails,
// the task remains unedited. The editor is run in the shell specified
// with the environment variable $SHELL (defaulting to "sh"). The
// editor is specified with the environment variable $EDITOR (defaulting
// to "vim").
func (t *Task) Edit() error {
	namePattern := "*_" + strings.ReplaceAll(t.Title, " ", "_")
	file, err := os.CreateTemp("", namePattern)
	if err != nil {
		return fmt.Errorf("%w: failed to open temp file", err)
	}

	_, err = fmt.Fprintf(file, "%s\n%s", t.Title, t.Story)
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
		if len(lines) > 1 {
			t.Story = strings.Join(lines[1:], "\n")
		}
	}

	return nil
}

// trim transforms any adjacent whitespace into a single
// space and removes any leading or trailing whitespace.
func trim(str string) string {
	return regexp.MustCompile(`\s+`).ReplaceAllString(strings.TrimSpace(str), " ")
}
