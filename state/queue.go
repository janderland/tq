package state

import (
	"encoding/json"
	"os"

	"github.com/pkg/errors"
)

type TaskQueue struct {
	TaskList  []*Task
	OpenIndex int
}

func NewQueue() TaskQueue {
	return TaskQueue{OpenIndex: -1}
}

// Load opens the file at the given path and deserializes
// a TaskQueue from the file's contents. If the file
// doesn't exist, an empty TaskQueue is returned.
func Load(path string) (TaskQueue, error) {
	tq := NewQueue()
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
		_ = task.Normalize()
	}
	return tq, nil
}

// Save serializes the TaskQueue and writes it to the
// file at the given path. If the file exists it's
// contents is overwritten.
func (tq TaskQueue) Save(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return errors.Wrap(err, "failed to create file")
	}
	if err := json.NewEncoder(file).Encode(tq); err != nil {
		return errors.Wrap(err, "failed to encode file")
	}
	return nil
}

// Insert inserts a new Task into the TaskQueue at the given index.
// The index must be between 0 and the current length of the
// TaskQueue (inclusive) or this method will panic.
func (tq TaskQueue) Insert(newTask *Task, index int) TaskQueue {
	tq.TaskList = append(tq.TaskList, nil)
	copy(tq.TaskList[index+1:], tq.TaskList[index:])
	tq.TaskList[index] = newTask
	if index <= tq.OpenIndex {
		tq.OpenIndex++
	}
	return tq
}

// Front moves the Task found at the given index to the
// front of the TaskQueue. This also causes the moved task
// to enter the "opened" state. The index must be the index
// of an existing Task or this method will panic.
func (tq TaskQueue) Front(index int) TaskQueue {
	openTask := tq.TaskList[index]
	copy(tq.TaskList[1:], tq.TaskList[:index])
	tq.TaskList[0] = openTask
	if index > tq.OpenIndex {
		tq.OpenIndex++
	}
	return tq
}

// Pop removes the Task found at the front of the TaskQueue.
func (tq TaskQueue) Pop() TaskQueue {
	tq.TaskList = tq.TaskList[1:]
	if tq.OpenIndex > -1 {
		tq.OpenIndex--
	}
	return tq
}

// At returns a pointer to the Task at the given index.
func (tq TaskQueue) At(index int) *Task {
	return tq.TaskList[index]
}

// ValidateNewIndex returns an error if the given index shouldn't be used as the index
// of a new task. New tasks are only supposed to be inserted after the opened tasks.
func (tq TaskQueue) ValidateNewIndex(index int) error {
	if index <= tq.OpenIndex || tq.Len() < index {
		return errors.Errorf("'index' must be in (%v, %v]", tq.OpenIndex, tq.Len())
	}
	return nil
}

// LastOpenedIndex returns the index of last
// task that is open.
func (tq TaskQueue) LastOpenedIndex() int {
	return tq.OpenIndex
}

// HasOpened returns true if this TaskQueue
// has at least one opened task.
func (tq TaskQueue) HasOpened() bool {
	return tq.OpenIndex >= 0
}

// Len returns the number of tasks in this TaskQueue.
func (tq TaskQueue) Len() int {
	return len(tq.TaskList)
}
