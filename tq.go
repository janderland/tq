package main

import (
	"fmt"
	"os"
	"time"

	"github.com/janderland/tq/state"
	"github.com/janderland/tq/ui"
	"github.com/spf13/cobra"
)

var (
	flags struct {
		queue string
		index int
		width int
		force bool
		all   bool
	}

	tasks state.TaskQueue
	ux    ui.UI
)

var rootCmd = &cobra.Command{
	Use:   "tq",
	Short: "Task Queue",
	Args:  cobra.NoArgs,
	PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
		var err error
		tasks, err = state.Load(flags.queue)
		if err != nil {
			return fmt.Errorf("%w: failed to load queue file", err)
		}
		ux = ui.New(flags.width)
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return topCmd.RunE(cmd, args)
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
		for index := 0; index <= tasks.LastOpenedIndex(); index++ {
			ux.Display(tasks, index)
		}
		return nil
	},
}

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Add a new task to the queue.",
	Args:  cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		newTask := state.Task{
			Created: time.Now(),
		}
		ux.Message("New task.\n")
		if err := ux.Edit(&newTask); err != nil {
			return err
		}
		if err := newTask.Normalize(); err != nil {
			return err
		}

		var index int
		if flags.index > -1 {
			if err := tasks.ValidateNewIndex(flags.index); err != nil {
				return err
			}
			index = flags.index
		} else {
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
		if tasks.Len() == 0 && tasks.DoneLen() == 0 {
			ux.Message("No tasks in queue.")
			return nil
		}
		for index := 0; index < tasks.Len(); index++ {
			ux.Display(tasks, index)
			if index == tasks.LastOpenedIndex() {
				ux.Line()
			}
		}
		if tasks.DoneLen() > 0 {
			ux.Line()
			for index := 0; index < tasks.DoneLen(); index++ {
				ux.DisplayDone(tasks, index)
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
			for index = 0; index < tasks.Len(); index++ {
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
		ux.Message("Editing task.\n")
		task := tasks.At(index)
		if err := ux.Edit(task); err != nil {
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
	Short: "Mark the current task as done.",
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
		ux.Message("Marking current task as done.")
		return tasks.Complete().Save(flags.queue)
	},
}

var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear done tasks from the queue.",
	Args:  cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		if flags.all {
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
		}

		if !flags.force {
			ux.Message("Are you sure you want to clear the done tasks?")
			yes, err := ux.QueryYesNo()
			if err != nil {
				return err
			}
			if !yes {
				return nil
			}
		}
		return tasks.ClearDone().Save(flags.queue)
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

	newCmd.Flags().IntVarP(&flags.index, "index", "i", -1, "new task's index in the queue")
	openCmd.Flags().IntVarP(&flags.index, "index", "i", -1, "index of task to open")
	editCmd.Flags().IntVarP(&flags.index, "index", "i", -1, "index of task to edit")
	clearCmd.Flags().BoolVarP(&flags.force, "force", "f", false, "clear without confirmation")
	clearCmd.Flags().BoolVar(&flags.all, "all", false, "clear all tasks, not just done tasks")

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
