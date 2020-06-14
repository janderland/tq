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

func (tq TaskQueue) insert(newTask *Task, index int) TaskQueue {
	tq.TaskList = append(tq.TaskList, nil)
	copy(tq.TaskList[index+1:], tq.TaskList[index:])
	tq.TaskList[index] = newTask
	if index <= tq.OpenIndex {
		tq.OpenIndex++
	}
	return tq
}

func (tq TaskQueue) front(index int) TaskQueue {
	openTask := tq.TaskList[index]
	copy(tq.TaskList[1:], tq.TaskList[:index])
	tq.TaskList[0] = openTask
	tq.OpenIndex++
	return tq
}

func (tq TaskQueue) pop() TaskQueue {
	tq.TaskList = tq.TaskList[1:]
	tq.OpenIndex--
	return tq
}

func (tq TaskQueue) len() int {
	return len(tq.TaskList)
}

func (t *Task) normalize() *Task {
	t.Title = strings.ToUpper(trim(t.Title))
	t.Story = trim(t.Story)
	return t
}
