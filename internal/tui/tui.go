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

	columnMaxWidth = 42

	// UI Styles
	columnStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(PrimaryColor)).
			Padding(0, 1).
			Width(columnMaxWidth-2).
			MaxWidth(columnMaxWidth).
			Height(22)

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

	terminalW int
	terminalH int
}

//---------------------------------------------------------------------------------------------------------------------------------

func (m Model) getTasksForDay(day int) []todo.Item {
	var filtered []todo.Item
	if day == 7 {
		for _, it := range *m.todoList {
			if it.IsSomeday {
				filtered = append(filtered, it)
			}
		}
		return filtered
	}

	targetDate := m.weekStart.AddDate(0, 0, day).Format("2006-01-02")
	for _, it := range *m.todoList {
		if !it.IsSomeday && it.Date.Format("2006-01-02") == targetDate {
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
						(*m.todoList)[i].Date = targetDate
						(*m.todoList)[i].IsSomeday = false
						break
					}
				}
				m.todoList.Save(env.TodoFileName)
				m.showMoveDialogWithCalender = false
				return m, nil
			}

			// Important: If we changed months, ensure the cursor isn't on day 31 of a 30-day month
			newDaysInMonth := time.Date(m.pickerYear, m.pickerMonth+1, 0, 0, 0, 0, 0, time.Local).Day()
			if m.pickerDay > newDaysInMonth {
				m.pickerDay = newDaysInMonth
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
			case "m":
				tasks := m.getTasksForDay(m.cursorDay)
				if len(tasks) > 0 && m.cursorIdx < len(tasks) {
					selected := tasks[m.cursorIdx]

					m.editingTaskID = selected.ID

					m.showMoveDialog = true

					return m, nil
				}

			case " ": // Toggle Task Done
				if len(tasks) > 0 && m.cursorIdx < len(tasks) {
					targetID := tasks[m.cursorIdx].ID

					for i := range *m.todoList {
						if (*m.todoList)[i].ID == targetID {
							(*m.todoList)[i].Done = !(*m.todoList)[i].Done
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
	style := columnStyle

	if dayIdx == 7 {
		dateLabel = "SOMEDAY"
	} else {
		d := m.weekStart.AddDate(0, 0, dayIdx)
		dateLabel = d.Format("Monday, Jan 02")
		if d.Format("2006-01-02") == time.Now().Format("2006-01-02") {
			style = todayStyle
		}
	}

	// Active column highlight
	if m.cursorDay == dayIdx {
		style = style.BorderForeground(SecondaryColor)
	}

	tasks := m.getTasksForDay(dayIdx)

	const maxVisibleTasks = 18
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
			line = runewidth.Truncate(line, columnMaxWidth-(2+5), "…")

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

func (m Model) renderMoveTaskDialog() string {
	content := lipgloss.JoinVertical(lipgloss.Center,
		lipgloss.NewStyle().Bold(true).Render("MOVE TASK"),
		"",
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

	// Grid
	row1 := lipgloss.JoinHorizontal(lipgloss.Top, m.renderDay(0), m.renderDay(1), m.renderDay(2), m.renderDay(3), m.renderDay(4), m.renderDay(5))
	row2 := lipgloss.JoinHorizontal(lipgloss.Top, m.renderDay(6), m.renderDay(7))

	grid := lipgloss.JoinVertical(lipgloss.Left, row1, row2)

	// Footer
	helpText := "• ←/→: Day, ↑/↓: Task, Space: Toggle , n: Add Task, e: Edit Task, m:Move task, Delete/x: Delete task, [: Prev Week, ]: Next Week, Esc/q: Quit •"
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
	} else {
		return mainView
	}

}
