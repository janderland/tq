package main

import (
	"fmt"
	"os"

	"github.com/janderland/tq/state"
	"github.com/janderland/tq/ui"
	"github.com/spf13/cobra"
)

var (
	flags struct {
		queue string
		title string
		story string
		index int
		width int
		force bool
	}

	tasks state.TaskQueue
	ux    ui.UI
)

var rootCmd = &cobra.Command{
	Use:   "tq",
	Short: "Task Queue",
	PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
		var err error
		tasks, err = state.Load(flags.queue)
		if err != nil {
			return fmt.Errorf("%w: failed to load queue file", err)
		}
		ux = ui.New(flags.width)
		return nil
	},
}

var topCmd = &cobra.Command{
	Use:   "top",
	Short: "View the current task.",
	Args:  cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		if tasks.Len() == 0 {
			ux.Message("No tasks in queue.")
			return nil
		}
		if !tasks.HasOpened() {
			ux.Message("No tasks are opened.")
			return nil
		}
		ux.Display(tasks, 0)
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
		if flags.index > -1 {
			flagCount++
		}
		if flagCount > 0 && flagCount < 3 {
			return fmt.Errorf("flags 'title', 'story', & 'index' must be all set or none")
		}

		newTask := state.Task{Title: "NEW"}
		var index int

		if flagCount != 0 {
			if err := tasks.ValidateNewIndex(flags.index); err != nil {
				return err
			}
			newTask.Title = flags.title
			newTask.Story = flags.story
			index = flags.index
		} else {
			if err := newTask.Edit(); err != nil {
				return err
			}
			if err := newTask.Normalize(); err != nil {
				return err
			}
			for index = tasks.Len(); index > tasks.LastOpenedIndex()+1; index-- {
				ux.Message("Should the new task be opened before this one?")
				ux.Display(tasks, index-1)
				yes, err := ux.QueryYesNo()
				if err != nil {
					return err
				}
				if !yes {
					break
				}
			}
		}

		ux.Message("Inserting new task at index %d.", index)
		return tasks.Insert(newTask, index).Save(flags.queue)
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tasks in the queue.",
	Args:  cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		if tasks.Len() == 0 {
			ux.Message("No tasks in queue.")
			return nil
		}
		for index := 0; index < tasks.Len(); index++ {
			ux.Display(tasks, index)
			if index == tasks.LastOpenedIndex() {
				ux.Line()
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
		if flags.index > -1 {
			index = flags.index
		} else {
			start := 0
			if tasks.HasOpened() {
				start = 1
			}
			for index = start; index < tasks.Len(); index++ {
				ux.Message("Would you like to open this task?")
				ux.Display(tasks, index)
				yes, err := ux.QueryYesNo()
				if err != nil {
					return err
				}
				if yes {
					break
				}
			}
			if index == tasks.Len() {
				ux.Message("End of queue. No task opened.")
				return nil
			}
		}
		ux.Message("Opening task.")
		return tasks.Front(index).Save(flags.queue)
	},
}

var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit a task.",
	Args:  cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		var index int
		if flags.index > -1 {
			index = flags.index
		} else {
			start := 0
			if tasks.HasOpened() {
				start = 1
			}
			for index = start; index < tasks.Len(); index++ {
				ux.Message("Would you like to edit this task?")
				ux.Display(tasks, index)
				yes, err := ux.QueryYesNo()
				if err != nil {
					return err
				}
				if yes {
					break
				}
			}
			if index == tasks.Len() {
				ux.Message("End of queue. No task edited.")
				return nil
			}
		}
		ux.Message("Editing task.")
		task := tasks.At(index)
		if err := task.Edit(); err != nil {
			return err
		}
		if err := task.Normalize(); err != nil {
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
			ux.Message("No tasks in queue.")
			return nil
		}
		ux.Message("Is this task done?")
		ux.Display(tasks, 0)
		yes, err := ux.QueryYesNo()
		if err != nil {
			return err
		}
		if !yes {
			ux.Message("Keeping current task.")
			return nil
		}
		ux.Message("Removing current task.")
		return tasks.Pop().Save(flags.queue)
	},
}

var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear all tasks from the queue.",
	Args:  cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		if !flags.force {
			ux.Message("Are you sure you want to clear all tasks?")
			yes, err := ux.QueryYesNo()
			if err != nil {
				return err
			}
			if !yes {
				return nil
			}
		}
		return state.NewQueue().Save(flags.queue)
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
	clearCmd.Flags().BoolVarP(&flags.force, "force", "f", false, "clear without confirmation")

	rootCmd.AddCommand(topCmd)
	rootCmd.AddCommand(newCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(openCmd)
	rootCmd.AddCommand(editCmd)
	rootCmd.AddCommand(doneCmd)
	rootCmd.AddCommand(clearCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}