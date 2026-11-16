# tq

A simple task queue (todo list).

![demo](demo.gif)

## Install

```sh
go install github.com/janderland/tq@latest
```

## Usage

```
Task Queue

Usage:
  tq [flags]
  tq [command]

Available Commands:
  clear       Clear done tasks from the queue.
  done        Mark the current task as done.
  edit        Edit a task.
  help        Help about any command
  list        List all tasks in the queue.
  new         Add a new task to the queue.
  open        Change your current task.
  top         View the current task.

Flags:
  -h, --help           help for tq
  -q, --queue string   file path for the queue's contents
  -w, --width int      width of displayed tasks (default 60)

Use "tq [command] --help" for more information about a command.
```

## Demo

The GIF above is generated with [VHS](https://github.com/charmbracelet/vhs):

```sh
go build -o tq . && vhs demo.tape
```
