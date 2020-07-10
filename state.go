package main

import (
	"encoding/json"
	"github.com/pkg/errors"
	"os"
	"strings"
)

type Task struct {
	Title string
	Story string
}

type TaskQueue struct {
	TaskList  []*Task
	OpenIndex int
}

// Opens the file at the given path and deserializes
// a TaskQueue from the file's contents. If the file
// doesn't exist, an empty TaskQueue is returned.
func load(path string) (TaskQueue, error) {
	tq := TaskQueue{OpenIndex: -1}
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return tq, nil
		}
		return tq, err
	}
	if err := json.NewDecoder(file).Decode(&tq); err != nil {
		return tq, errors.Wrap(err, "failed to decode file")
	}
	for _, task := range tq.TaskList {
		task.normalize()
	}
	return tq, nil
}

// Serializes the TaskQueue and writes it to the
// file at the given path. If the file exists it's
// contents is overwritten.
func (tq TaskQueue) save(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return errors.Wrap(err, "failed to create file")
	}
	if err := json.NewEncoder(file).Encode(tq); err != nil {
		return errors.Wrap(err, "failed to encode file")
	}
	return nil
}

// Insert a new Task into the TaskQueue at the given index. The
// index must be between 0 and the current length of the TaskQueue
// (inclusive) or this method will panic.
func (tq TaskQueue) insert(newTask *Task, index int) TaskQueue {
	tq.TaskList = append(tq.TaskList, nil)
	copy(tq.TaskList[index+1:], tq.TaskList[index:])
	tq.TaskList[index] = newTask
	if index <= tq.OpenIndex {
		tq.OpenIndex++
	}
	return tq
}

// Moves the Task found at the given index to the front
// of the TaskQueue. This also causes the moved task to
// enter the "opened" state. The index must be the index
// of an existing Task or this method will panic.
func (tq TaskQueue) front(index int) TaskQueue {
	openTask := tq.TaskList[index]
	copy(tq.TaskList[1:], tq.TaskList[:index])
	tq.TaskList[0] = openTask
	if index > tq.OpenIndex {
		tq.OpenIndex++
	}
	return tq
}

// Removes the Task found at the front of the TaskQueue.
func (tq TaskQueue) pop() TaskQueue {
	tq.TaskList = tq.TaskList[1:]
	if tq.OpenIndex > -1 {
		tq.OpenIndex--
	}
	return tq
}

// Returns an error if the given index shouldn't be used as the index of a new
// task. New tasks are only supposed to be inserted after the opened tasks.
func (tq TaskQueue) validateNewIndex(index int) error {
	if index <= tq.OpenIndex || tq.len() < index {
		return errors.Errorf("'index' must be in (%v, %v]", tq.OpenIndex, tasks.len())
	}
	return nil
}

func (tq TaskQueue) lastOpenedIndex() int {
	return tq.OpenIndex
}

// Returns true if this TaskQueue has at
// least one opened task.
func (tq TaskQueue) hasOpened() bool {
	return tq.OpenIndex >= 0
}

// The number of tasks in this TaskQueue.
func (tq TaskQueue) len() int {
	return len(tq.TaskList)
}

// Normalizes the text of the Task.
func (t *Task) normalize() *Task {
	t.Title = strings.ToUpper(trim(t.Title))
	t.Story = trim(t.Story)
	return t
}
