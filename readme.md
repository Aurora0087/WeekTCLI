# WeekTCLI

WeekTCLI is a minimalist, weekly calendar-based task manager designed for the terminal. Inspired by the Tweek.so workflow, it moves away from traditional vertical lists in favor of a horizontal weekly grid, allowing users to visualize their entire week and manage an unscheduled "Someday" drawer.

## Overview

Unlike standard todo applications that focus on a single list, WeekTCLI provides a structured environment where tasks are anchored to specific days. This layout helps users balance their workload across the week and provides a dedicated space for long-term tasks that haven't yet been assigned a date.

## Features

- Weekly Grid Layout: View and manage tasks across a seven-day spread (Monday to Sunday) plus a dedicated Someday list.
- Interactive TUI: A full-screen terminal user interface built with the Bubble Tea framework.
- Quick Entry: Add tasks directly into specific days using an integrated modal dialog without leaving the weekly view.
- Pager-style Details: View full task metadata and multi-line notes in a dedicated inspector view.
- Shadcn-inspired Date Picker: Move tasks between days or weeks using a clean, grid-based calendar selector.
- Persistence: All data is stored locally in a JSON format for easy backup and portability.
- CLI Integration: Support for standard command-line arguments to add, delete, or toggle tasks quickly without opening the full UI.

## Keyboard Controls

### Navigation
- h / l or Arrow Left / Right: Move between days.
- j / k or Arrow Up / Down: Navigate tasks within the selected day.
- PgUp / PgDn: Navigate between previous and next weeks.

### Task Management
- n: Create a new task on the selected day.
- e: Edit the selected task's title and notes.
- m: Open the move menu to reschedule a task or send it to Someday.
- Space: Toggle task completion status.
- Delete / Backspace: Remove the selected task.
- i / Enter: Open the task details inspector.

### General
- Esc: Exit the application or close active modals.

## Technical Details

WeekTCLI is written in Go and utilizes the following libraries:
- Cobra: For CLI command structure and flag parsing.
- Bubble Tea: For the functional TUI architecture.
- Lipgloss: For terminal styling and layout management.
- Bubbles: For interactive components like text inputs and text areas.

## Installation

Ensure you have Go installed on your system, then clone the repository and build the binary:

```bash
go build -o weektcli
```

To run the application:

```bash
./weektcli tui
```