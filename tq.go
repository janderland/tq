package main

import (
	"github.com/janderland/tq/state"
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

var tasks state.TaskQueue
var ui UI

var rootCmd = &cobra.Command{
	Use:   "tq",
	Short: "Task Queue",
	PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
		var err error
		tasks, err = state.Load(flags.queue)
		if err != nil {
			return errors.Wrap(err, "failed to load queue file")
		}
		ui = UI{width: flags.width}
		return nil
	},
}

var topCmd = &cobra.Command{
	Use:   "top",
	Short: "View the current task.",
	Args:  cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		if tasks.Len() == 0 {
			ui.message("No tasks in queue.")
			return nil
		}
		if !tasks.HasOpened() {
			ui.message("No tasks are opened.")
			return nil
		}
		ui.display(tasks, 0)
		return nil
	},
}

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Add a new task to the queue.",
	Args:  cobra.NoArgs,
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

		var newTask state.Task
		var index int
		var err error

		if flagCount != 0 {
			if err := tasks.ValidateNewIndex(flags.index); err != nil {
				return err
			}
			newTask.Title = flags.title
			newTask.Story = flags.story
			index = flags.index
		} else {
			ui.message("Input the title.")
			newTask.Title, err = ui.queryLine()
			if err != nil {
				return errors.Wrap(err, "failed to read title")
			}

			ui.message("Input the story.")
			newTask.Story, err = ui.queryLine()
			if err != nil {
				return errors.Wrap(err, "failed to read story")
			}

			for index = tasks.Len(); index > tasks.LastOpenedIndex()+1; index-- {
				ui.message("Should the new task be opened before this one?")
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

		ui.message("Inserting new task at index %d.", index)
		return tasks.Insert(newTask.Normalize(), index).Save(flags.queue)
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tasks in the queue.",
	Args:  cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		if tasks.Len() == 0 {
			ui.message("No tasks in queue.")
			return nil
		}
		for index := 0; index < tasks.Len(); index++ {
			ui.display(tasks, index)
			if index == tasks.LastOpenedIndex() {
				ui.line()
			}
		}
		return nil
	},
}

var openCmd = &cobra.Command{
	Use:   "open",
	Short: "Change your current task.",
	Args:  cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		var index int
		if flags.index != -1 {
			index = flags.index
		} else {
			start := 0
			if tasks.HasOpened() {
				start = 1
			}
			for index = start; index < tasks.Len(); index++ {
				ui.message("Would you like to open this task?")
				ui.display(tasks, index)
				yes, err := ui.queryYesNo()
				if err != nil {
					return err
				}
				if yes {
					break
				}
			}
			if index == tasks.Len() {
				ui.message("End of queue. No task opened.")
				return nil
			}
		}
		ui.message("Opening task.")
		return tasks.Front(index).Save(flags.queue)
	},
}

var editCmd = &cobra.Command{
	Use: "edit",
	Short: "Edit a task.",
	Args: cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		var index int
		if flags.index != -1 {
			index = flags.index
		} else {
			start := 0
			if tasks.HasOpened() {
				start = 1
			}
			for index = start; index < tasks.Len(); index++ {
				ui.message("Would you like to edit this task?")
				ui.display(tasks, index)
				yes, err := ui.queryYesNo()
				if err != nil {
					return err
				}
				if yes {
					break
				}
			}
			if index == tasks.Len() {
				ui.message("End of queue. No task edited.")
				return nil
			}
		}
		ui.message("Editing task.")
		if err := tasks.At(index).Edit(); err != nil {
			return err
		}
		return tasks.Save(flags.queue)
	},
}

var doneCmd = &cobra.Command{
	Use:   "done",
	Short: "Remove the current task from the queue.",
	Args:  cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		if tasks.Len() == 0 {
			ui.message("No tasks in queue.")
			return nil
		}
		ui.message("Is this task done?")
		ui.display(tasks, 0)
		yes, err := ui.queryYesNo()
		if err != nil {
			return err
		}
		if !yes {
			ui.message("Keeping current task.")
			return nil
		}
		ui.message("Removing current task.")
		return tasks.Pop().Save(flags.queue)
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
	editCmd.Flags().IntVarP(&flags.index, "index", "i", -1, "index of task to edit")

	rootCmd.AddCommand(topCmd)
	rootCmd.AddCommand(newCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(openCmd)
	rootCmd.AddCommand(editCmd)
	rootCmd.AddCommand(doneCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
