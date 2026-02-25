package main

import (
	"fmt"
	"weektcli/env"
	"weektcli/internal/todo"
	"weektcli/internal/tui"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var todoList todo.List

func main() {
	todoList.Load(env.TodoFileName)

	var rootCmd = &cobra.Command{Use: "weektcli"}

	// --- EXISTING ADD COMMAND ---
	var someday bool
	var dateStr string
	var addCmd = &cobra.Command{
		Use:   "add [task]",
		Short: "Add a task to a day or Someday",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var taskDate time.Time
			if dateStr != "" {
				parsedDate, _ := time.Parse("2006-01-02", dateStr)
				taskDate = parsedDate
			} else {
				taskDate = time.Now()
			}
			todoList.Add(args[0], "", taskDate, someday)
			todoList.Save(env.TodoFileName)
			fmt.Printf("Added: %s\n", args[0])
		},
	}
	addCmd.Flags().BoolVarP(&someday, "someday", "s", false, "Add to Someday list")
	addCmd.Flags().StringVarP(&dateStr, "date", "d", "", "Specific date (YYYY-MM-DD)")

	// --- NEW: DELETE COMMAND ---
	var deleteCmd = &cobra.Command{
		Use:   "delete [id]",
		Short: "Delete a task by ID",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := todoList.DeleteTask(env.TodoFileName, args[0])
			if err != nil {
				fmt.Println("Error:", err)
				return
			}
			fmt.Println("Task deleted.")
		},
	}

	// --- NEW: TOGGLE COMMAND ---
	var toggleCmd = &cobra.Command{
		Use:   "toggle [id]",
		Short: "Toggle task done/undone",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Logic: Find task, flip boolean, save
			found := false
			for i := range todoList {
				if todoList[i].ID.String() == args[0] {
					todoList[i].Done = !todoList[i].Done
					found = true
					break
				}
			}
			if found {
				todoList.Save(env.TodoFileName)
				fmt.Println("Task status toggled.")
			} else {
				fmt.Println("Task not found.")
			}
		},
	}

	// --- NEW: EDIT COMMAND ---
	var notes string
	var editCmd = &cobra.Command{
		Use:   "edit [id] [new title]",
		Short: "Edit a task title and notes",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			todoList.UpdateTask(uuid.MustParse(args[0]), args[1], notes)
			todoList.Save(env.TodoFileName)
			fmt.Println("Task updated.")
		},
	}
	editCmd.Flags().StringVarP(&notes, "notes", "n", "", "Update notes for the task")

	// --- NEW: DETAILS COMMAND ---
	var getCmd = &cobra.Command{
		Use:   "get [id]",
		Short: "Get full details of a task",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			for _, t := range todoList {
				if t.ID.String() == args[0] {
					fmt.Printf("ID:      %s\n", t.ID)
					fmt.Printf("Task:    %s\n", t.Task)
					fmt.Printf("Done:    %v\n", t.Done)
					fmt.Printf("Notes:   %s\n", t.Notes)
					fmt.Printf("Date:    %s\n", t.Date.Format("2006-01-02"))
					return
				}
			}
			fmt.Println("Task not found.")
		},
	}

	// --- TUI COMMANDS ---
	var tuiCmd = &cobra.Command{
		Use:   "tui",
		Short: "Open Weekly View",
		Run: func(cmd *cobra.Command, args []string) {
			p := tea.NewProgram(tui.InitialModel(&todoList), tea.WithAltScreen())
			p.Run()
			todoList.Save(env.TodoFileName)
		},
	}

	rootCmd.AddCommand(addCmd, deleteCmd, toggleCmd, editCmd, getCmd, tuiCmd)
	rootCmd.Execute()
}