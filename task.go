package main

import (
	"encoding/json"
	"github.com/pkg/errors"
	"os"
	"strings"
)

const (
	todoState = "todo"
	openState = "open"
)

type Task struct {
	Title string
	Story string
	State string
}

func read(path string) ([]*Task, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var tasks []*Task
	if err := json.NewDecoder(file).Decode(&tasks); err != nil {
		return nil, errors.Wrap(err, "failed to decode file")
	}
	for _, task := range tasks {
		normalize(task)
	}
	return tasks, nil
}

func write(path string, tasks []*Task) error {
	file, err := os.Create(path)
	if err != nil {
		return errors.Wrap(err, "failed to create file")
	}
	if err := json.NewEncoder(file).Encode(tasks); err != nil {
		return errors.Wrap(err, "failed to encode file")
	}
	return nil
}

func insert(tasks []*Task, newTask *Task, index int) []*Task {
	tasks = append(tasks, nil)
	copy(tasks[index+1:], tasks[index:])
	tasks[index] = newTask
	return tasks
}

func open(tasks []*Task, index int) []*Task {
	openTask := tasks[index]
	openTask.State = openState
	copy(tasks[1:], tasks[:index])
	tasks[0] = openTask
	return tasks
}

func normalize(task *Task) *Task {
	task.Title = strings.ToUpper(trim(task.Title))
	task.Story = trim(task.Story)
	return task
}
