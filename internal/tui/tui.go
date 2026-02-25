package tui

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
	"weektcli/env"
	"weektcli/internal/todo"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"github.com/mattn/go-runewidth"
)

var (

	// colors
	PrimaryColor      = lipgloss.Color("#818cf8")
	PrimaryForeground = lipgloss.Color("#101010")

	SecondaryColor      = lipgloss.Color("#2dd4bf")
	SecondaryForeground = lipgloss.Color("#000000")

	AccentColor = lipgloss.Color("#fcd34d")

	DestructiveColor = lipgloss.Color("#f87171")

	todayDayColor = lipgloss.Color("#f472b6")

	CardBackgroundColor = lipgloss.Color("#1a212b")
	CardForegroundColor = lipgloss.Color("#ffffff")

	columnMaxWidth  = 42
	columnMaxHeight = 19

	// UI Styles
	columnStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(PrimaryColor)).
			Padding(0, 1).
			Width(columnMaxWidth - 2).
			MaxWidth(columnMaxWidth).
			Height(columnMaxHeight - 2).
			MaxHeight(columnMaxHeight)

	todayStyle = columnStyle.
			BorderForeground(todayDayColor)

	titleStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(PrimaryColor)).
			Foreground(lipgloss.Color(PrimaryForeground)).
			Padding(0, 1).
			Bold(true)

	headerStyle = lipgloss.NewStyle().
			Background(SecondaryColor).
			Foreground(PrimaryForeground).
			Padding(0, 1).
			MarginLeft(1).
			Bold(true)

	footerStyle = lipgloss.NewStyle().
			Foreground(AccentColor).
			Align(lipgloss.Center)

	choiceStyle = lipgloss.NewStyle().
			Foreground(SecondaryColor).
			Bold(true)

	highlightedTask = lipgloss.NewStyle().
			Background(SecondaryColor).
			Foreground(PrimaryForeground)

	dialogBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(SecondaryColor).
			Padding(1, 4).
			Width(60)

	deleteBoxStyle = dialogBoxStyle.BorderForeground(DestructiveColor).Align(lipgloss.Center)

	taskDetailsBoxStyle = dialogBoxStyle.Width(80).BorderForeground(SecondaryColor)

	detailBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(SecondaryColor).
			Padding(1, 2).
			Width(60)

	detailHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#1a1a1a")).
				Background(SecondaryColor).
				Bold(true).
				Padding(0, 1).
				MarginBottom(1)

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272a4")).
			Bold(true).
			Width(12)

	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#f8f8f2"))

	notesBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true). // Left border only
			BorderForeground(lipgloss.Color("#44475a")).
			PaddingLeft(2).
			MarginTop(1)

	pickerActiveStyle = lipgloss.NewStyle().
				Background(SecondaryColor).
				Foreground(PrimaryForeground).
				Bold(true).
				Padding(0, 1)

	pickerInactiveStyle = lipgloss.NewStyle().
				Padding(0, 1).
				Foreground(lipgloss.Color("#f8f8f2"))

	calendarBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(SecondaryColor).
				Padding(1, 2)

	monthHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(PrimaryColor).
				MarginBottom(1)

	weekdayStyle = lipgloss.NewStyle().
			Foreground(SecondaryColor).
			Width(4).
			Align(lipgloss.Center)

	// The "Solid" selection look from Shadcn
	selectedDayStyle = lipgloss.NewStyle().
				Background(SecondaryColor).
				Foreground(SecondaryForeground).
				Bold(true).
				Width(4).
				Align(lipgloss.Center)

	todayDayStyle = lipgloss.NewStyle().
			Foreground(todayDayColor).
			Bold(true).
			Width(4).
			Align(lipgloss.Center)

	normalDayStyle = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Width(4).
			Align(lipgloss.Center)

	ruleActiveStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(SecondaryColor)).
			Foreground(lipgloss.Color(SecondaryForeground)).
			Padding(0, 1).Bold(true)

	ruleInactiveStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#27272a")).
				Foreground(lipgloss.Color("#ffffff")).
				Padding(0, 1)

	ruleFocusStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(lipgloss.Color(SecondaryColor)).
			PaddingLeft(2)
)

//---------------------------------------------------------------------------------------------------------------------------------

type Model struct {
	todoList  *todo.List
	cursorDay int
	cursorIdx int
	weekStart time.Time

	textInput   textinput.Model
	noteInput   textarea.Model
	showNewTask bool

	showConfirmDeleteDialog bool

	selectedTask    *todo.Item
	showTaskDetails bool

	editingTaskID uuid.UUID
	showEditTask  bool

	showMoveDialog             bool
	showMoveDialogWithCalender bool
	pickerDay                  int
	pickerMonth                time.Month
	pickerYear                 int

	showRecurrenceRuleDialog bool
	ruleFocus                int
	ruleWeekdayCursor        int
	tempRule                 todo.RecurrenceRule

	terminalW int
	terminalH int

	columnMaxWidth  int
	columnMaxHeight int
}

//---------------------------------------------------------------------------------------------------------------------------------

func (m Model) getTasksForDay(day int) []todo.Item {
	var filtered []todo.Item

	// 1. Handle Someday
	if day == 7 {
		for _, it := range *m.todoList {
			if it.IsSomeday {
				filtered = append(filtered, it)
			}
		}
		return filtered
	}

	// 2. Normalize Target Date to Midnight (Prevents hour/minute math issues)
	tY, tM, tD := m.weekStart.AddDate(0, 0, day).Date()
	targetDate := time.Date(tY, tM, tD, 0, 0, 0, 0, time.Local)
	targetDateStr := targetDate.Format("2006-01-02")

	for _, it := range *m.todoList {
		if it.IsSomeday {
			continue
		}

		// Normalize Item Start Date to Midnight
		sY, sM, sD := it.Date.Date()
		startDate := time.Date(sY, sM, sD, 0, 0, 0, 0, time.Local)

		isMatch := false

		// --- Case A: One-time Task ---
		if it.RecurrenceRule == nil {
			if startDate.Equal(targetDate) {
				isMatch = true
			}
		} else {
			// --- Case B: Recurring Task ---

			// Never show before the start date
			if targetDate.Before(startDate) {
				continue
			}

			rule := it.RecurrenceRule
			interval := int(rule.Interval)
			if interval <= 0 {
				interval = 1
			}

			// Use integer division on days for DST safety
			daysDiff := int(targetDate.Sub(startDate).Hours() / 24)

			switch rule.Freq {
			case todo.Daily:
				if daysDiff%interval == 0 {
					isMatch = true
				}

			case todo.Weekly:
				weekdayMatch := false

				// If no weekdays were selected (null/empty),
				// default to the weekday of the task's original start date.
				if len(rule.Weekdays) == 0 {
					if targetDate.Weekday() == startDate.Weekday() {
						weekdayMatch = true
					}
				} else {
					// Normal matching logic
					for _, wd := range rule.Weekdays {
						if wd == targetDate.Weekday() {
							weekdayMatch = true
							break
						}
					}
				}

				if weekdayMatch {
					// Calculate weeks diff safely
					weeksDiff := daysDiff / 7
					if weeksDiff%interval == 0 {
						isMatch = true
					}
				}
			case todo.Monthly:
				// 1. Calculate how many months have passed since the start date
				startYear, startMonth, startDay := startDate.Date()
				targetYear, targetMonth, targetDay := targetDate.Date()

				// Formula for total months difference: (YearDiff * 12) + MonthDiff
				monthsDiff := (targetYear-startYear)*12 + int(targetMonth-startMonth)

				// 2. Check if the month hits our interval (e.g., every 1 month, every 3 months)
				if monthsDiff >= 0 && monthsDiff%interval == 0 {

					// 3. Check if today's day (e.g., the 26th) matches the target day
					// If MonthDay is 0 (default), use the day from the task's original start date
					matchDay := int(rule.MonthDay)
					if matchDay <= 0 {
						matchDay = startDay
					}

					// Handle the match
					if targetDay == matchDay {
						isMatch = true
					} else {
						// OPTIONAL: "End of month" safety.
						// If the task is set for the 31st, but this month only has 30 days,
						// show it on the last day of the month.
						lastDayOfMonth := time.Date(targetYear, targetMonth+1, 0, 0, 0, 0, 0, time.Local).Day()
						if matchDay > lastDayOfMonth && targetDay == lastDayOfMonth {
							isMatch = true
						}
					}
				}
			}

			// Handle Completion for Recurring Tasks
			if isMatch {
				it.Done = false // The master 'Done' is ignored for recurring instances
				for _, doneDate := range rule.DoneList {
					if doneDate == targetDateStr {
						it.Done = true
						break
					}
				}
			}
		}

		if isMatch {
			filtered = append(filtered, it)
		}
	}
	return filtered
}

//---------------------------------------------------------------------------------------------------------------------------------

func InitialModel(l *todo.List) Model {
	now := time.Now()
	offset := int(now.Weekday()) - int(time.Monday)
	if offset < 0 {
		offset += 7
	}

	ti := textinput.New()
	ti.Placeholder = "New task..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 30

	ta := textarea.New()
	ta.Placeholder = "Add some notes here..."
	ta.SetWidth(30)
	ta.SetHeight(3)

	return Model{
		todoList:                   l,
		weekStart:                  now.AddDate(0, 0, -offset),
		cursorDay:                  offset,
		textInput:                  ti,
		noteInput:                  ta,
		showNewTask:                false,
		cursorIdx:                  0,
		showConfirmDeleteDialog:    false,
		showTaskDetails:            false,
		showEditTask:               false,
		showMoveDialog:             false,
		showMoveDialogWithCalender: false,
		ruleFocus:                  0,
		ruleWeekdayCursor:          1,
		columnMaxWidth:             columnMaxWidth,
		columnMaxHeight:            columnMaxHeight,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.terminalW = msg.Width
		m.terminalH = msg.Height
		return m, nil

	case tea.KeyMsg:
		tasks := m.getTasksForDay(m.cursorDay)

		if m.showNewTask {
			// add new task
			switch msg.String() {

			case "tab":
				if m.textInput.Focused() {
					m.textInput.Blur()
					m.noteInput.Focus()
				} else {
					m.noteInput.Blur()
					m.textInput.Focus()
				}
				return m, nil

			case "enter":
				if m.textInput.Focused() {
					taskName := m.textInput.Value()
					noteText := m.noteInput.Value()
					if taskName != "" {
						isSomeday := m.cursorDay == 7
						taskDate := m.weekStart.AddDate(0, 0, m.cursorDay)
						m.todoList.Add(taskName, noteText, taskDate, isSomeday)
						m.todoList.Save(env.TodoFileName)
					}
					m.showNewTask = false
					m.textInput.Reset()
					m.noteInput.Reset()
					return m, nil
				}
			case "ctrl+s":
				taskName := m.textInput.Value()
				if taskName != "" {
					m.todoList.Add(taskName, m.noteInput.Value(), m.weekStart.AddDate(0, 0, m.cursorDay), m.cursorDay == 7)
					m.todoList.Save(env.TodoFileName)
				}
				m.showNewTask = false
				m.textInput.Reset()
				m.noteInput.Reset()
				return m, nil

			case "esc":
				m.showNewTask = false
				m.textInput.Reset()
				m.noteInput.Reset()
				return m, cmd
			}
		} else if m.showConfirmDeleteDialog {

			// delete task
			switch msg.String() {
			case "enter":
				tasks := m.getTasksForDay(m.cursorDay)

				if len(tasks) > 0 && m.cursorIdx < len(tasks) {
					idToDelete := tasks[m.cursorIdx].ID.String()

					m.todoList.DeleteTask(env.TodoFileName, idToDelete)

					if m.cursorIdx > 0 && m.cursorIdx >= len(tasks)-1 {
						m.cursorIdx--
					}
				}
				m.showConfirmDeleteDialog = false
				return m, cmd
			case "q", "esc":
				m.showConfirmDeleteDialog = false
				return m, cmd
			}
		} else if m.showTaskDetails {

			//task details
			switch msg.String() {
			case "q", "esc":
				m.showTaskDetails = false
				return m, cmd
			case "e": // go to edit
				tasks := m.getTasksForDay(m.cursorDay)
				if len(tasks) > 0 && m.cursorIdx < len(tasks) {
					selected := tasks[m.cursorIdx]

					m.editingTaskID = selected.ID

					m.textInput.SetValue(selected.Task)
					m.noteInput.SetValue(selected.Notes)

					m.showTaskDetails = false
					m.showEditTask = true
					m.textInput.Focus()
					m.noteInput.Blur()

					return m, nil
				}
			}

		} else if m.showEditTask {

			// edit task title and note
			switch msg.String() {
			case "esc":
				m.showEditTask = false
				return m, cmd
			case "tab":
				if m.textInput.Focused() {
					m.textInput.Blur()
					m.noteInput.Focus()
				} else {
					m.noteInput.Blur()
					m.textInput.Focus()
				}
				return m, nil
			case "enter", "ctrl+s":
				if msg.String() == "ctrl+s" || (msg.String() == "enter" && m.textInput.Focused()) {
					taskName := m.textInput.Value()
					noteText := m.noteInput.Value()

					if taskName != "" {
						m.todoList.UpdateTask(m.editingTaskID, taskName, noteText)

						m.todoList.Save(env.TodoFileName)
					}

					m.showEditTask = false
					m.textInput.Reset()
					m.noteInput.Reset()
					return m, nil
				}
			}

		} else if m.showMoveDialog {

			// move task dialog
			switch msg.String() {
			case "esc":
				m.showMoveDialog = false
				return m, cmd
			case "s":
				for i := range *m.todoList {
					if (*m.todoList)[i].ID == m.editingTaskID {
						(*m.todoList)[i].IsSomeday = true
						(*m.todoList)[i].Date = time.Time{}
						break
					}
				}
				m.todoList.Save(env.TodoFileName)
				m.showMoveDialog = false
				return m, nil
			case "t":
				now := time.Now()
				for i := range *m.todoList {
					if (*m.todoList)[i].ID == m.editingTaskID {
						(*m.todoList)[i].IsSomeday = false
						(*m.todoList)[i].Date = now
						break
					}
				}
				m.todoList.Save(env.TodoFileName)
				m.showMoveDialog = false
				return m, nil
			case "c":
				now := time.Now()
				m.pickerDay = now.Day()
				m.pickerMonth = now.Month()
				m.pickerYear = now.Year()
				m.showMoveDialog = false
				m.showMoveDialogWithCalender = true

				return m, nil
			}

		} else if m.showMoveDialogWithCalender {
			daysInMonth := time.Date(m.pickerYear, m.pickerMonth+1, 0, 0, 0, 0, 0, time.Local).Day()

			switch msg.String() {
			case "esc":
				m.showMoveDialogWithCalender = false
				return m, nil

			// --- MONTH NAVIGATION ---
			case "]": // Next Month
				m.pickerMonth++
				if m.pickerMonth > 12 {
					m.pickerMonth = 1
					m.pickerYear++
				}
			case "[": // Previous Month
				m.pickerMonth--
				if m.pickerMonth < 1 {
					m.pickerMonth = 12
					m.pickerYear--
				}

			// --- YEAR NAVIGATION ---
			case "pgup": // Next Year
				m.pickerYear++
			case "pgdown": // Previous Year
				m.pickerYear--

			// --- DAY NAVIGATION ---
			case "left", "h":
				if m.pickerDay > 1 {
					m.pickerDay--
				}
			case "right", "l":
				if m.pickerDay < daysInMonth {
					m.pickerDay++
				}
			case "up", "k":
				if m.pickerDay > 7 {
					m.pickerDay -= 7
				}
			case "down", "j":
				if m.pickerDay <= daysInMonth-7 {
					m.pickerDay += 7
				}

			case "enter":
				targetDate := time.Date(m.pickerYear, m.pickerMonth, m.pickerDay, 0, 0, 0, 0, time.Local)

				for i := range *m.todoList {
					if (*m.todoList)[i].ID == m.editingTaskID {
						item := &(*m.todoList)[i]

						item.Date = targetDate
						item.IsSomeday = false

						if item.RecurrenceRule != nil {
							switch item.RecurrenceRule.Freq {

							case todo.Weekly:
								item.RecurrenceRule.Weekdays = []time.Weekday{targetDate.Weekday()}

							case todo.Monthly:
								item.RecurrenceRule.MonthDay = uint8(targetDate.Day())
							}
						}
						break
					}
				}

				m.todoList.Save(env.TodoFileName)
				m.showMoveDialogWithCalender = false
				return m, nil
			}

			newDaysInMonth := time.Date(m.pickerYear, m.pickerMonth+1, 0, 0, 0, 0, 0, time.Local).Day()
			if m.pickerDay > newDaysInMonth {
				m.pickerDay = newDaysInMonth
			}
		} else if m.showRecurrenceRuleDialog {
			switch msg.String() {
			case "esc":
				m.showRecurrenceRuleDialog = false
				return m, nil

			case "tab":
				m.ruleFocus = (m.ruleFocus + 1) % 2
				return m, nil

			case "left", "h":
				switch m.ruleFocus {
				case 0: // Change Frequency
					if m.tempRule.Freq > 0 {
						m.tempRule.Freq--
					}
				case 2: // Move Weekday Cursor
					if m.ruleWeekdayCursor > 0 {
						m.ruleWeekdayCursor--
					}
				}

			case "right", "l":
				switch m.ruleFocus {
				case 0: // Change Frequency
					if m.tempRule.Freq < 3 {
						m.tempRule.Freq++
					}
				case 2: // Move Weekday Cursor
					if m.ruleWeekdayCursor < 6 {
						m.ruleWeekdayCursor++
					}
				}

			case "up", "k":
				if m.ruleFocus == 1 { // Increase Interval
					m.tempRule.Interval++
				}

			case "down", "j":
				if m.ruleFocus == 1 { // Decrease Interval
					if m.tempRule.Interval > 1 {
						m.tempRule.Interval--
					}
				}

			case " ": // Toggle Weekday (Space)
				if m.ruleFocus == 2 && m.tempRule.Freq == todo.Weekly {
					// Map cursor (0-6) to time.Weekday (Mon=1 ... Sun=0)
					// Go's Sunday is 0, Monday is 1
					targetWD := time.Weekday((m.ruleWeekdayCursor + 1) % 7)

					// Toggle logic
					found := -1
					for i, wd := range m.tempRule.Weekdays {
						if wd == targetWD {
							found = i
							break
						}
					}

					if found != -1 {
						// Remove weekday
						m.tempRule.Weekdays = append(m.tempRule.Weekdays[:found], m.tempRule.Weekdays[found+1:]...)
					} else {
						// Add weekday
						m.tempRule.Weekdays = append(m.tempRule.Weekdays, targetWD)
					}
				}

			case "enter":
				// SAFETY: If Weekly mode is selected but NO weekdays are picked,
				// add the weekday of the original task date automatically.
				if m.tempRule.Freq == todo.Weekly && len(m.tempRule.Weekdays) == 0 {
					// Find the original task in the list
					for _, it := range *m.todoList {
						if it.ID == m.editingTaskID {
							m.tempRule.Weekdays = []time.Weekday{it.Date.Weekday()}
							break
						}
					}
				}

				m.todoList.UpdateRecurrenceRule(m.editingTaskID, m.tempRule)
				m.todoList.Save(env.TodoFileName)
				m.showRecurrenceRuleDialog = false
				return m, nil
			}
		} else {

			// all actions on key

			switch msg.String() {
			case "ctrl+c", "q", "esc":
				return m, tea.Quit

			case "left", "h":
				if m.cursorDay > 0 {
					m.cursorDay--
					m.cursorIdx = 0
				}
			case "right", "l":
				if m.cursorDay < 7 {
					m.cursorDay++
					m.cursorIdx = 0
				}
			case "up", "k":
				if m.cursorIdx > 0 {
					m.cursorIdx--
				}
			case "down", "j":
				if m.cursorIdx < len(tasks)-1 {
					m.cursorIdx++
				}
			case "n":
				if !m.showNewTask {
					m.showNewTask = true
					m.textInput.Focus()
					return m, nil
				}
			case "x", "delete", "backspace":
				m.showConfirmDeleteDialog = true

			case "e": // go to edit
				tasks := m.getTasksForDay(m.cursorDay)
				if len(tasks) > 0 && m.cursorIdx < len(tasks) {
					selected := tasks[m.cursorIdx]

					m.editingTaskID = selected.ID

					m.textInput.SetValue(selected.Task)
					m.noteInput.SetValue(selected.Notes)

					m.showEditTask = true
					m.textInput.Focus()
					m.noteInput.Blur()

					return m, nil
				}
			case "r": // Open Recurrence Settings
				tasks := m.getTasksForDay(m.cursorDay)
				if len(tasks) > 0 {
					selected := tasks[m.cursorIdx]
					m.editingTaskID = selected.ID

					if selected.RecurrenceRule != nil {
						m.tempRule = *selected.RecurrenceRule // Copy existing
					} else {
						// Create a default new rule
						m.tempRule = todo.RecurrenceRule{
							Freq:     todo.None,
							Interval: 1,
							Weekdays: []time.Weekday{},
							DoneList: []string{},
							MonthDay: 0,
						}
					}

					m.showRecurrenceRuleDialog = true
					m.ruleFocus = 0 // Start focus on Frequency
				}
			case "m":
				tasks := m.getTasksForDay(m.cursorDay)
				if len(tasks) > 0 && m.cursorIdx < len(tasks) {
					selected := tasks[m.cursorIdx]

					m.editingTaskID = selected.ID

					m.showMoveDialog = true

					return m, nil
				}

			case " ": // Toggle Task Done
				tasks := m.getTasksForDay(m.cursorDay)
				if len(tasks) > 0 && m.cursorIdx < len(tasks) {
					selectedTask := tasks[m.cursorIdx]
					targetID := selectedTask.ID

					// Calculate the date string for the current day being toggled
					// This is important because we need to know WHICH day of the recurring task we are finishing
					targetDateStr := m.weekStart.AddDate(0, 0, m.cursorDay).Format("2006-01-02")

					for i := range *m.todoList {
						if (*m.todoList)[i].ID == targetID {
							item := &(*m.todoList)[i]

							// --- CASE 1: Recurring Task ---
							if item.RecurrenceRule != nil {
								// Check if this date is already in the DoneList
								foundIdx := -1
								for idx, date := range item.RecurrenceRule.DoneList {
									if date == targetDateStr {
										foundIdx = idx
										break
									}
								}

								if foundIdx != -1 {
									// Date found: Uncheck it (Remove from list)
									item.RecurrenceRule.DoneList = append(
										item.RecurrenceRule.DoneList[:foundIdx],
										item.RecurrenceRule.DoneList[foundIdx+1:]...,
									)
								} else {
									// Date not found: Check it (Add to list)
									item.RecurrenceRule.DoneList = append(item.RecurrenceRule.DoneList, targetDateStr)
								}

							} else {
								// --- CASE 2: One-time Task ---
								item.Done = !item.Done
							}

							// Save immediately to persist the change
							m.todoList.Save(env.TodoFileName)
							break
						}
					}
				}
			case "i", "enter": // View task details
				tasks := m.getTasksForDay(m.cursorDay)
				if len(tasks) > 0 && m.cursorIdx < len(tasks) {

					task, err := m.todoList.GetTaskDetails(tasks[m.cursorIdx].ID.String())
					if err == nil {
						m.selectedTask = &task
						m.showTaskDetails = true
					}
				}

			case "[":
				m.weekStart = m.weekStart.AddDate(0, 0, -7)
			case "]":
				m.weekStart = m.weekStart.AddDate(0, 0, 7)
			}
		}

	}
	m.textInput, cmd = m.textInput.Update(msg)
	cmds = append(cmds, cmd)

	m.noteInput, cmd = m.noteInput.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

//---------------------------------------------------------------------------------------------------------------------------------

func overlay(background, foreground string, x, y int) string {

	//https://lmika.org/2022/09/24/overlay-composition-using.html
	bgLines := strings.Split(background, "\n")
	fgLines := strings.Split(foreground, "\n")
	fgWidth := lipgloss.Width(fgLines[0])

	for i, fgLine := range fgLines {
		targetY := y + i
		if targetY < 0 || targetY >= len(bgLines) {
			continue
		}

		bgLine := bgLines[targetY]

		//left part (before the dialog)
		leftPart, _ := splitAtVisual(bgLine, x)

		//right part (after the dialog)
		_, rightPart := splitAtVisual(bgLine, x+fgWidth)

		//combine
		padding := ""
		leftWidth := lipgloss.Width(leftPart)
		if leftWidth < x {
			padding = strings.Repeat(" ", x-leftWidth)
		}

		bgLines[targetY] = leftPart + padding + fgLine + rightPart
	}

	return strings.Join(bgLines, "\n")
}

func splitAtVisual(s string, width int) (string, string) {
	var left, right strings.Builder
	visualWidth := 0
	inAnsi := false // Flag to track if we are currently reading an escape code like \x1b[31m

	i := 0
	for i < len(s) {
		// --- 1. DETECT START OF ANSI SEQUENCE ---
		// If we hit the ESC character (\x1b), we have started an invisible color/style code.
		if s[i] == '\x1b' {
			inAnsi = true
		}

		// --- 2. HANDLE INVISIBLE ANSI BYTES ---
		if inAnsi {
			char := s[i]
			// We append the ANSI bytes to whichever side we are currently building,
			// but we DO NOT increment visualWidth because these are invisible to the user.
			if visualWidth < width {
				left.WriteByte(char)
			} else {
				right.WriteByte(char)
			}

			// ANSI sequences usually end with a letter (like 'm' in \x1b[0m).
			// Once we hit a letter, the invisible sequence is finished.
			if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') {
				inAnsi = false
			}
			i++
			continue // Skip to the next byte
		}

		// --- 3. HANDLE VISIBLE CHARACTERS (RUNES) ---
		// We decode the next UTF-8 character (rune). size = how many bytes it takes (1-4).
		r, size := utf8.DecodeRuneInString(s[i:])

		// runewidth.RuneWidth determines how many columns the character takes on screen.
		// Examples: 'A' = 1, '界' = 2, '😀' = 2.
		w := runewidth.RuneWidth(r)

		// --- 4. DECIDE WHICH SIDE TO PUT THE CHARACTER ---
		if visualWidth < width {
			// If adding this character would push us past the requested split width,
			// we must put the entire character into the 'right' side to avoid
			// "ghost" spaces or misaligned columns.
			if visualWidth+w > width {
				right.WriteString(string(r))
			} else {
				left.WriteString(string(r))
			}
		} else {
			// We have already reached or passed the split point, everything goes to the right.
			right.WriteString(string(r))
		}

		// --- 5. INCREMENT COUNTERS ---
		visualWidth += w // Track total visual space used so far
		i += size        // Jump the byte index forward by the size of the rune
	}

	return left.String(), right.String()
}

//---------------------------------------------------------------------------------------------------------------------------------

// ui components

func (m Model) renderDay(dayIdx int) string {
	var dateLabel string
	style := columnStyle.Width(m.columnMaxWidth - 2).Height(m.columnMaxHeight - 2)

	if dayIdx == 7 {
		dateLabel = "SOMEDAY"
	} else {
		d := m.weekStart.AddDate(0, 0, dayIdx)
		dateLabel = d.Format("Monday, Jan 02")
		if d.Format("2006-01-02") == time.Now().Format("2006-01-02") {
			style = todayStyle.Height(m.columnMaxHeight - 2)
		}
	}

	// Active column highlight
	if m.cursorDay == dayIdx {
		style = style.BorderForeground(SecondaryColor)
	}

	tasks := m.getTasksForDay(dayIdx)

	var maxVisibleTasks = m.columnMaxHeight - (2 + 4)
	var taskList strings.Builder

	if len(tasks) == 0 {
		taskList.WriteString("\n  (No tasks)")
	} else {

		start, end := 0, len(tasks)
		if len(tasks) > maxVisibleTasks {
			if m.cursorDay == dayIdx {
				start = m.cursorIdx - (maxVisibleTasks / 2)
				if start < 0 {
					start = 0
				}

				end = start + maxVisibleTasks
				if end > len(tasks) {
					end = len(tasks)
					start = end - maxVisibleTasks
				}
			} else {
				start = 0
				end = maxVisibleTasks
			}
		}
		//"More tasks up" indicator
		if start > 0 {
			taskList.WriteString(lipgloss.NewStyle().Faint(true).Render(fmt.Sprintf("\n  ↑ +%d more", start)))
		} else {
			taskList.WriteString("\n")
		}

		//only the tasks in the window
		for i := start; i < end; i++ {
			t := tasks[i]
			check := "[ ]"
			if t.Done {
				check = "[✔]"
			}
			line := fmt.Sprintf("%s %s", check, t.Task)

			//truncate long task names
			line = runewidth.Truncate(line, m.columnMaxWidth-(2+5), "…")

			if m.cursorDay == dayIdx && m.cursorIdx == i {
				taskList.WriteString(highlightedTask.Render("> "+line) + "\n")
			} else {
				taskList.WriteString("  " + line + "\n")
			}
		}

		//"More tasks below" indicator
		if end < len(tasks) {
			taskList.WriteString(lipgloss.NewStyle().Faint(true).Render(fmt.Sprintf("  ↓ +%d more", len(tasks)-end)))
		}
	}

	content := fmt.Sprintf("%s\n%s", titleStyle.Render(dateLabel), taskList.String())
	return style.Render(content)
}

func (m Model) renderNewTaskDialog() string {

	dayName := "Someday"
	if m.cursorDay < 7 {
		dayName = m.weekStart.AddDate(0, 0, m.cursorDay).Format("Monday")
	}

	activeStyle := lipgloss.NewStyle().Foreground(SecondaryColor).Bold(true)
	inactiveStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	titleLabel := " Title "
	notesLabel := " Notes "

	if m.textInput.Focused() {
		titleLabel = activeStyle.Render("▶ Title")
	} else {
		titleLabel = inactiveStyle.Render("  Title")
	}

	if m.noteInput.Focused() {
		notesLabel = activeStyle.Render("▶ Notes")
	} else {
		notesLabel = inactiveStyle.Render("  Notes")
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Bold(true).Render("Add task to "+dayName),
		"",
		titleLabel,
		m.textInput.View(),
		"",
		notesLabel,
		m.noteInput.View(),
		"",
		footerStyle.MarginTop(1).Render("\nTab: Switch, 󰆓 Enter (on Title): Save, 󰆓 Ctrl+S: Save, 󰜺 Esc: Cancel"),
	)

	return dialogBoxStyle.Render(content)
}

func (m Model) renderEditTaskDialog() string {
	// 1. Logic for context title
	dayName := "Someday"
	if m.cursorDay < 7 {
		dayName = m.weekStart.AddDate(0, 0, m.cursorDay).Format("Monday, Jan 02")
	}

	// 2. Styling
	activeStyle := lipgloss.NewStyle().Foreground(SecondaryColor).Bold(true)
	inactiveStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	// Different header color for "Edit" vs "Add"
	headerStyle := lipgloss.NewStyle().
		Foreground(PrimaryForeground).
		Background(AccentColor).
		Padding(0, 1).
		MarginBottom(1).
		Bold(true)

	titleLabel := " Title "
	notesLabel := " Notes "

	// Visual indicators for focus
	if m.textInput.Focused() {
		titleLabel = activeStyle.Render("▶ Title")
	} else {
		titleLabel = inactiveStyle.Render("  Title")
	}

	if m.noteInput.Focused() {
		notesLabel = activeStyle.Render("▶ Notes")
	} else {
		notesLabel = inactiveStyle.Render("  Notes")
	}

	// 3. Assemble Content
	content := lipgloss.JoinVertical(lipgloss.Left,
		headerStyle.Render("EDITING TASK"),
		lipgloss.NewStyle().Faint(true).Render("On "+dayName),
		"",
		titleLabel,
		m.textInput.View(),
		"",
		notesLabel,
		m.noteInput.View(),
		"",
		footerStyle.MarginTop(1).Render("Tab: Switch, 󰆓 Enter/Ctrl+S: Save, 󰜺 Esc: Cancel"),
	)

	return dialogBoxStyle.Render(content)
}

func (m Model) renderUpdateRecurrenceRuleDialog() string {
	// --- 1. FREQUENCY ROW ---
	freqs := []string{"None", "Daily", "Weekly", "Monthly"}
	var freqButtons []string
	for i, f := range freqs {
		str := f
		style := ruleInactiveStyle

		// Highlight if this is the currently selected frequency
		if int(m.tempRule.Freq) == i {
			style = ruleActiveStyle
		}
		// Add a special indicator if this specific row is focused
		if m.ruleFocus == 0 && int(m.tempRule.Freq) == i {
			str = "▶ " + str
		}
		freqButtons = append(freqButtons, style.Render(str))
	}
	freqRow := lipgloss.JoinHorizontal(lipgloss.Left, freqButtons...)
	if m.ruleFocus == 0 {
		freqRow = ruleFocusStyle.Render(freqRow)
	}

	// --- 2. INTERVAL ROW ---
	unit := "days"
	switch m.tempRule.Freq {
	case todo.Weekly:
		unit = "weeks"
	case todo.Monthly:
		unit = "months"
	}
	intervalStr := fmt.Sprintf("Repeat every: [ %d ] %s", m.tempRule.Interval, unit)
	if m.ruleFocus == 1 {
		intervalStr = pickerActiveStyle.Render(intervalStr)
	}

	// --- 4. ASSEMBLE ---
	content := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Bold(true).MarginBottom(1).Render("RECURRENCE SETTINGS"),
		"Frequency:",
		freqRow,
		"",
		intervalStr,
		"",
		footerStyle.Render("Tab: Move, ←→↑↓: Change Value, 󰆓 Enter: Save, 󰜺 Esc: Cancel"),
	)

	return dialogBoxStyle.Render(content)
}

func (m Model) renderMoveTaskDialog() string {
	content := lipgloss.JoinVertical(lipgloss.Center,
		lipgloss.NewStyle().Bold(true).Render("MOVE TASK"),
		"",
		fmt.Sprintf("Press %s to move to Today", choiceStyle.Render("t")),
		fmt.Sprintf("Press %s to move to Someday", choiceStyle.Render("s")),
		fmt.Sprintf("Press %s to choose a Date", choiceStyle.Render("c")),
		"",
		lipgloss.NewStyle().Faint(true).Render("(esc to cancel)"),
	)

	return dialogBoxStyle.Align(lipgloss.Center).Render(content)
}

func (m Model) renderDeleteTaskConfirmDialog() string {
	content := lipgloss.JoinVertical(lipgloss.Center,
		lipgloss.NewStyle().Bold(true).Foreground(DestructiveColor).Render("󰆴 Enter to Delete Task"),
		lipgloss.NewStyle().Bold(true).Render("󰜺 Esc to Cancel"),
	)

	return deleteBoxStyle.Render(content)
}

func (m Model) renderTaskDetails() string {
	if m.selectedTask == nil {
		return ""
	}

	t := m.selectedTask

	// 1. Header (Task Name)
	header := detailHeaderStyle.Render(" TASK DETAILS ")
	taskTitle := lipgloss.NewStyle().Bold(true).Render(t.Task)

	// 2. Metadata Section (Status, Date, ID)
	status := lipgloss.NewStyle().Foreground(DestructiveColor).Render("In Progress")
	if t.Done {
		status = lipgloss.NewStyle().Foreground(SecondaryColor).Render("Completed ✔")
	}

	dateStr := t.Date.Format("Monday, Jan 02, 2006")
	if t.IsSomeday {
		dateStr = lipgloss.NewStyle().Foreground(AccentColor).Render("Someday Drawer")
	}

	// Create rows of metadata
	metaRows := lipgloss.JoinVertical(lipgloss.Left,
		fmt.Sprintf("%s %s", labelStyle.Render("Status:"), status),
		fmt.Sprintf("%s %s", labelStyle.Render("Scheduled:"), dateStr),
	)

	// 3. Notes Section
	notesTitle := labelStyle.Render("Notes:")
	notesContent := t.Notes
	if notesContent == "" {
		notesContent = lipgloss.NewStyle().Faint(true).Render("No notes provided.")
	}
	notesBody := notesBoxStyle.Render(notesContent)

	// 4. Footer hints
	footer := footerStyle.MarginTop(1).
		Render("\n e to edit, 󰜺 Esc to close")

	// Assemble everything
	content := lipgloss.JoinVertical(lipgloss.Left,
		header,
		taskTitle,
		"", // Spacer
		metaRows,
		"", // Spacer
		notesTitle,
		notesBody,
		footer,
	)

	return detailBoxStyle.Render(content)
}

func (m Model) renderMoveTaskDatePicker() string {
	// Header: "October 2023"
	header := monthHeaderStyle.Render(fmt.Sprintf("%s %d", m.pickerMonth.String(), m.pickerYear))

	// Weekday Headers: Su Mo Tu We Th Fr Sa
	daysOfWeek := []string{"Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"}
	var dayHeader strings.Builder
	for _, d := range daysOfWeek {
		dayHeader.WriteString(weekdayStyle.Render(d))
	}

	// Calculate Month Grid
	firstDay := time.Date(m.pickerYear, m.pickerMonth, 1, 0, 0, 0, 0, time.Local)
	startOffset := int(firstDay.Weekday())
	daysInMonth := time.Date(m.pickerYear, m.pickerMonth+1, 0, 0, 0, 0, 0, time.Local).Day()

	var calendar strings.Builder
	column := 0

	// Padding for the first week
	for i := 0; i < startOffset; i++ {
		calendar.WriteString(lipgloss.NewStyle().Width(4).Render(""))
		column++
	}

	for day := 1; day <= daysInMonth; day++ {
		currDate := time.Date(m.pickerYear, m.pickerMonth, day, 0, 0, 0, 0, time.Local)
		style := normalDayStyle

		if m.pickerDay == day {
			style = selectedDayStyle
		} else if currDate.Format("2006-01-02") == time.Now().Format("2006-01-02") {
			style = todayDayStyle
		}

		calendar.WriteString(style.Render(fmt.Sprintf("%d", day)))
		column++

		if column == 7 {
			calendar.WriteString("\n")
			column = 0
		}
	}

	footer := footerStyle.MarginTop(1).Render(
		"←→↑↓: Nav, [:Pre Month, ]: Next Month, PgDn:Pre Year,PgUp: Next Year\nEnter: Pick, Esc: Close",
	)

	return calendarBoxStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
		header,
		dayHeader.String(),
		calendar.String(),
		footer,
	))
}

//---------------------------------------------------------------------------------------------------------------------------------

func (m Model) View() string {
	// Header
	weekRange := fmt.Sprintf(" %s - %s ",
		m.weekStart.Format("Jan 02"),
		m.weekStart.AddDate(0, 0, 6).Format("Jan 02, 2006"))
	header := headerStyle.Render(env.AppName) + "  " + weekRange

	// grid
	unitWidth := m.columnMaxWidth

	colsPerRow := m.terminalW / unitWidth
	if colsPerRow <= 2 {
		m.columnMaxHeight = 14
		colsPerRow = 2
	} else {
		m.columnMaxHeight = 19
	}
	if colsPerRow > 8 {
		colsPerRow = 8
	}

	var rows []string
	var currentRow []string

	for i := 0; i < 8; i++ {
		currentRow = append(currentRow, m.renderDay(i))

		//if current row is full / we r at the very last box
		if (i+1)%colsPerRow == 0 || i == 7 {
			rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, currentRow...))
			currentRow = []string{}
		}
	}

	grid := lipgloss.JoinVertical(lipgloss.Left, rows...)

	// Footer
	helpText := "• ← →: Day, ↑ ↓: Task, Space: Toggle, n:  Add Task, e:  Edit Task, m:  Move task, r:  Recurrence Setting, Delete/x: 󰆴 Delete task, [: Prev Week, ]: Next Week, Esc/q: Quit •"
	footer := footerStyle.Width(m.terminalW).MarginTop(1).Render(helpText)

	mainView := lipgloss.JoinVertical(lipgloss.Left, header, grid, footer)

	dimmedBG := lipgloss.NewStyle().Faint(true).Render(mainView)
	bgWidth := lipgloss.Width(dimmedBG)
	bgHeight := lipgloss.Height(dimmedBG)

	if m.showNewTask {

		dialog := m.renderNewTaskDialog()

		// Calculate the center position
		fgWidth := lipgloss.Width(dialog)
		fgHeight := lipgloss.Height(dialog)

		// Calculate top-left corner for the dialog to be centered
		x := (bgWidth - fgWidth) / 2
		y := (bgHeight - fgHeight) / 2

		// Return the overlaid result
		return overlay(dimmedBG, dialog, x, y)
	} else if m.showConfirmDeleteDialog {

		dialog := m.renderDeleteTaskConfirmDialog()

		// Calculate the center position
		fgWidth := lipgloss.Width(dialog)
		fgHeight := lipgloss.Height(dialog)

		// Calculate top-left corner for the dialog to be centered
		x := (bgWidth - fgWidth) / 2
		y := (bgHeight - fgHeight) / 2

		// Return the overlaid result
		return overlay(dimmedBG, dialog, x, y)
	} else if m.showTaskDetails && m.selectedTask != nil {

		taskDetails := m.renderTaskDetails()

		// Calculate the center position
		fgWidth := lipgloss.Width(taskDetails)
		fgHeight := lipgloss.Height(taskDetails)

		// Calculate top-left corner for the dialog to be centered
		x := (bgWidth - fgWidth) / 2
		y := (bgHeight - fgHeight) / 2

		// Return the overlaid result
		return overlay(dimmedBG, taskDetails, x, y)
	} else if m.showEditTask {

		taskEditDialog := m.renderEditTaskDialog()

		// Calculate the center position
		fgWidth := lipgloss.Width(taskEditDialog)
		fgHeight := lipgloss.Height(taskEditDialog)

		// Calculate top-left corner for the dialog to be centered
		x := (bgWidth - fgWidth) / 2
		y := (bgHeight - fgHeight) / 2

		// Return the overlaid result
		return overlay(dimmedBG, taskEditDialog, x, y)
	} else if m.showMoveDialog {

		taskMoveDialog := m.renderMoveTaskDialog()

		// Calculate the center position

		fgWidth := lipgloss.Width(taskMoveDialog)
		fgHeight := lipgloss.Height(taskMoveDialog)

		// Calculate top-left corner for the dialog to be centered
		x := (bgWidth - fgWidth) / 2
		y := (bgHeight - fgHeight) / 2

		// Return the overlaid result
		return overlay(dimmedBG, taskMoveDialog, x, y)
	} else if m.showMoveDialogWithCalender {
		taskMoveDialog := m.renderMoveTaskDatePicker()

		// Calculate the center position
		fgWidth := lipgloss.Width(taskMoveDialog)
		fgHeight := lipgloss.Height(taskMoveDialog)

		// Calculate top-left corner for the dialog to be centered
		x := (bgWidth - fgWidth) / 2
		y := (bgHeight - fgHeight) / 2

		// Return the overlaid result
		return overlay(dimmedBG, taskMoveDialog, x, y)
	} else if m.showRecurrenceRuleDialog {
		taskMoveDialog := m.renderUpdateRecurrenceRuleDialog()

		// Calculate the center position
		fgWidth := lipgloss.Width(taskMoveDialog)
		fgHeight := lipgloss.Height(taskMoveDialog)

		// Calculate top-left corner for the dialog to be centered
		x := (bgWidth - fgWidth) / 2
		y := (bgHeight - fgHeight) / 2

		// Return the overlaid result
		return overlay(dimmedBG, taskMoveDialog, x, y)
	} else {
		return mainView
	}

}
