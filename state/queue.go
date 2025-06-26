package state

import (
	"encoding/json"
	"fmt"
	"os"
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
		return tq, fmt.Errorf("%w: failed to decode file", err)
	}
	for _, task := range tq.TaskList {
		_ = task.Normalize()
	}
	return tq, nil
}

// Save serializes the TaskQueue and writes it to the
// file at the given path. If the file exists it's
// contents is overwritten.
func (q TaskQueue) Save(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("%w: failed to create file", err)
	}
	if err := json.NewEncoder(file).Encode(q); err != nil {
		return fmt.Errorf("%w: failed to encode file", err)
	}
	return nil
}

// Insert inserts a new Task into the TaskQueue at the given index.
// The index must be between 0 and the current length of the
// TaskQueue (inclusive) or this method will panic.
func (q TaskQueue) Insert(newTask Task, index int) TaskQueue {
	q.TaskList = append(q.TaskList, nil)
	copy(q.TaskList[index+1:], q.TaskList[index:])
	q.TaskList[index] = &newTask
	if index <= q.OpenIndex {
		q.OpenIndex++
	}
	return q
}

// Front moves the Task found at the given index to the
// front of the TaskQueue. This also causes the moved task
// to enter the "opened" state. The index must be the index
// of an existing Task or this method will panic.
func (q TaskQueue) Front(index int) TaskQueue {
	openTask := q.TaskList[index]
	copy(q.TaskList[1:], q.TaskList[:index])
	q.TaskList[0] = openTask
	if index > q.OpenIndex {
		q.OpenIndex++
	}
	return q
}

// Pop removes the Task found at the front of the TaskQueue.
func (q TaskQueue) Pop() TaskQueue {
	q.TaskList = q.TaskList[1:]
	if q.OpenIndex > -1 {
		q.OpenIndex--
	}
	return q
}

// At returns a pointer to the Task at the given index.
func (q TaskQueue) At(index int) *Task {
	return q.TaskList[index]
}

// ValidateNewIndex returns an error if the given index shouldn't be used as the index
// of a new task. New tasks are only supposed to be inserted after the opened tasks.
func (q TaskQueue) ValidateNewIndex(index int) error {
	if index <= q.OpenIndex || q.Len() < index {
		return fmt.Errorf("'index' must be in (%v, %v]", q.OpenIndex, q.Len())
	}
	return nil
}

// LastOpenedIndex returns the index of last
// task that is open.
func (q TaskQueue) LastOpenedIndex() int {
	return q.OpenIndex
}

// HasOpened returns true if this TaskQueue
// has at least one opened task.
func (q TaskQueue) HasOpened() bool {
	return q.OpenIndex >= 0
}

// Len returns the number of tasks in this TaskQueue.
func (q TaskQueue) Len() int {
	return len(q.TaskList)
}
