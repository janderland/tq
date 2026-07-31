package state

import (
	"fmt"
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

// trim transforms any adjacent whitespace into a single
// space and removes any leading or trailing whitespace.
func trim(str string) string {
	return regexp.MustCompile(`\s+`).ReplaceAllString(strings.TrimSpace(str), " ")
}
