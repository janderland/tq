package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"os"
	"regexp"
	"strings"
)

var flags struct {
	queue string
	title string
	story string
	index int
	width int
}

var rootCmd = &cobra.Command{
	Use:   "tq",
	Short: "Task Queue",
}

var topCmd = &cobra.Command{
	Use:   "top",
	Short: "View the current task.",
	Long: trim(`
		View the current task. This is the task at the front
		of the queue.`),
	RunE: func(_ *cobra.Command, _ []string) error {
		tasks, err := read(flags.queue)
		if err != nil {
			return errors.Wrap(err, "failed to read queue file")
		}
		if len(tasks) == 0 {
			fmt.Println("no tasks in queue")
			return nil
		}
		display(tasks, 0)
		return nil
	},
}

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Add a new task to the queue.",
	Long: trim(`
		Add a new task to the queue. Iterate through the
		tasks in the queue from back to front, evaluating
		whether each task has higher priority than the new
		task. When a higher priority task is found, insert
		the new task behind this task.`),
	RunE: func(_ *cobra.Command, _ []string) error {
		tasks, err := read(flags.queue)
		if err != nil {
			return errors.Wrap(err, "failed to read queue file")
		}

		flagCount := 0
		if flags.title != "" {
			flagCount++
		}
		if flags.story != "" {
			flagCount++
		}
		if flags.index != -1 {
			flagCount++
		}
		if flagCount > 0 && flagCount < 3 {
			return errors.New("flags 'title', 'story', & 'index' must be all set or none")
		}

		newTask := Task{State: todoState}
		var index int

		if flagCount != 0 {
			if flags.index > len(tasks) || flags.index < 0 {
				return errors.Errorf(
					"flag 'index' must be in (0, %v)", len(tasks))
			}
			newTask.Title = flags.title
			newTask.Story = flags.story
			index = flags.index
		} else {
			reader := bufio.NewReader(os.Stdin)

			fmt.Print("+ Input the title.\n\n")
			newTask.Title, err = reader.ReadString('\n')
			if err != nil {
				return errors.Wrap(err, "failed to read title")
			}

			fmt.Print("\n+ Input the story.\n\n")
			newTask.Story, err = reader.ReadString('\n')
			if err != nil {
				return errors.Wrap(err, "failed to read title")
			}

			for index = len(tasks); index > 0; index-- {
				fmt.Println()
				fmt.Println("+ Should the new task be opened before this one?")
				fmt.Println()
				display(tasks, index-1)
				fmt.Println()
				yes, err := queryYesNo()
				if err != nil {
					return err
				}
				if !yes {
					break
				}
			}
		}

		fmt.Printf("\n+ Inserting new task at index %d.\n", index)
		return write(flags.queue, insert(tasks, normalize(&newTask), index))
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tasks in the queue.",
	RunE: func(_ *cobra.Command, _ []string) error {
		tasks, err := read(flags.queue)
		if err != nil {
			return errors.Wrap(err, "failed to read queue file")
		}
		if len(tasks) == 0 {
			fmt.Println("no tasks in queue")
			return nil
		}
		lastIndex := len(tasks) - 1
		for index := range tasks[:lastIndex] {
			display(tasks, index)
			fmt.Println()
		}
		display(tasks, lastIndex)
		return nil
	},
}

var openCmd = &cobra.Command{
	Use:   "open",
	Short: "Change your current task.",
	Long: trim(`
		Change the current task. Iterate through the tasks in
		the queue from front to back. Select a task, change
		it's state to "open", and move it to the front of the
		queue.`),
	RunE: func(_ *cobra.Command, _ []string) error {
		tasks, err := read(flags.queue)
		if err != nil {
			return errors.Wrap(err, "failed to read queue file")
		}

		var index int

		if flags.index != -1 {
			index = flags.index
		} else {
			for index = 0; index < len(tasks); index++ {
				if index != 0 {
					fmt.Println()
				}
				fmt.Println("+ Would you like to open this task?")
				fmt.Println()
				display(tasks, index)
				fmt.Println()
				yes, err := queryYesNo()
				if err != nil {
					return err
				}
				if yes {
					break
				}
			}
		}

		return write(flags.queue, open(tasks, index))
	},
}

var doneCmd = &cobra.Command{
	Use:   "done",
	Short: "Remove the current task from the queue.",
	RunE: func(_ *cobra.Command, _ []string) error {
		tasks, err := read(flags.queue)
		if err != nil {
			return errors.Wrap(err, "failed to read queue file")
		}
		if len(tasks) == 0 {
			fmt.Println("no tasks in queue")
			return nil
		}
		return write(flags.queue, tasks[1:])
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(
		&flags.queue, "queue", "q", "", "file path for the queue's contents")
	if err := rootCmd.MarkPersistentFlagRequired("queue"); err != nil {
		panic(err)
	}

	rootCmd.PersistentFlags().IntVarP(
		&flags.width, "width", "w", 60, "width of displayed tasks")

	newCmd.Flags().StringVarP(&flags.title, "title", "t", "", "new task's title")
	newCmd.Flags().StringVarP(&flags.story, "story", "s", "", "new task's story")
	newCmd.Flags().IntVarP(&flags.index, "index", "i", -1, "new task's index in the queue")
	openCmd.Flags().IntVarP(&flags.index, "index", "i", -1, "index of task to open")

	rootCmd.AddCommand(topCmd)
	rootCmd.AddCommand(newCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(openCmd)
	rootCmd.AddCommand(doneCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

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

func display(tasks []*Task, index int) {
	task := tasks[index]
	title := fmt.Sprintf("%d. ", index)
	if index < 10 {
		title += " "
	}
	count := len(title)
	words := append(strings.Split(task.Title, " "), fmt.Sprintf("[%s]", task.State))
	for _, word := range words {
		title += word
		count += len(word)
		if count > flags.width {
			title += "\n    "
			count = 4
		} else {
			title += " "
			count++
		}
	}

	story := "    "
	count = len(story)
	for _, word := range strings.Split(task.Story, " ") {
		story += word
		count += len(word)
		if count > flags.width {
			story += "\n    "
			count = 4
		} else {
			story += " "
			count++
		}
	}

	fmt.Println(title)
	fmt.Println(story)
}

func trim(str string) string {
	return regexp.MustCompile(`\s+`).ReplaceAllString(strings.TrimSpace(str), " ")
}

func queryYesNo() (bool, error) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter y|n: ")
		resp, err := reader.ReadString('\n')
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
