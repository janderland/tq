package main

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"os"
)

var flags struct {
	queue string
	title string
	story string
	index int
	width int
}

var tasks []*Task
var ui *UI

var rootCmd = &cobra.Command{
	Use:   "tq",
	Short: "Task Queue",
	PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
		var err error
		tasks, err = read(flags.queue)
		if err != nil {
			return errors.Wrap(err, "failed to read queue file")
		}
		ui = &UI{Width: flags.width}
		return nil
	},
}

var topCmd = &cobra.Command{
	Use:   "top",
	Short: "View the current task.",
	Long: trim(`
		View the current task. This is the task at the front
		of the queue.`),
	RunE: func(_ *cobra.Command, _ []string) error {
		if len(tasks) == 0 {
			ui.print("No tasks in queue.")
			return nil
		}
		ui.display(tasks, 0)
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
		var err error

		if flagCount != 0 {
			if flags.index > len(tasks) || flags.index < 0 {
				return errors.Errorf(
					"flag 'index' must be in (0, %v)", len(tasks))
			}
			newTask.Title = flags.title
			newTask.Story = flags.story
			index = flags.index
		} else {
			ui.print("Input the title.")
			newTask.Title, err = ui.queryLine()
			if err != nil {
				return errors.Wrap(err, "failed to read title")
			}

			ui.print("Input the story.")
			newTask.Story, err = ui.queryLine()
			if err != nil {
				return errors.Wrap(err, "failed to read title")
			}

			for index = len(tasks); index > 0; index-- {
				ui.print("Should the new task be opened before this one?")
				ui.display(tasks, index-1)
				yes, err := ui.queryYesNo()
				if err != nil {
					return err
				}
				if !yes {
					break
				}
			}
		}

		ui.print("Inserting new task at index %d.", index)
		return write(flags.queue, insert(tasks, normalize(&newTask), index))
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tasks in the queue.",
	RunE: func(_ *cobra.Command, _ []string) error {
		if len(tasks) == 0 {
			ui.print("No tasks in queue.")
			return nil
		}
		for index := range tasks {
			ui.display(tasks, index)
		}
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
		var index int
		if flags.index != -1 {
			index = flags.index
		} else {
			for index = 0; index < len(tasks); index++ {
				ui.print("Would you like to open this task?")
				ui.display(tasks, index)
				yes, err := ui.queryYesNo()
				if err != nil {
					return err
				}
				if yes {
					break
				}
			}
			if index == len(tasks) {
				ui.print("End of queue. No task opened.")
				return nil
			}
		}
		ui.print("Opening task.")
		return write(flags.queue, open(tasks, index))
	},
}

var doneCmd = &cobra.Command{
	Use:   "done",
	Short: "Remove the current task from the queue.",
	RunE: func(_ *cobra.Command, _ []string) error {
		if len(tasks) == 0 {
			ui.print("No tasks in queue.")
			return nil
		}
		ui.print("Is this task done?")
		ui.display(tasks, 0)
		yes, err := ui.queryYesNo()
		if err != nil {
			return err
		}
		if !yes {
			ui.print("Keeping current task.")
			return nil
		}
		ui.print("Removing current task.")
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
