package todo

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
)
type Frequency uint8

const (
	None Frequency = iota
	Daily
	Weekly
	Monthly
)

type RecurrenceRule struct {
	Freq      Frequency   `json:"freq"`
	Interval  uint8       `json:"interval"`
	Weekdays  []time.Weekday `json:"weekdays"`
	MonthDay  uint8       `json:"month_day"`
	DoneList  []string    `json:"done_list"`
}

type Item struct {
	ID        uuid.UUID `json:"id"`
	Task      string    `json:"task"`
	Notes     string    `json:"notes"`
	Done      bool      `json:"done"`
	Date      time.Time `json:"date"`
	IsSomeday bool      `json:"is_someday"`
	RecurrenceRule *RecurrenceRule `json:"recurrence_rule,omitempty"`
}

type List []Item

func (l *List) Add(task,note string, date time.Time, someday bool) Item {
	item := Item{
		ID:        uuid.New(),
		Task:      task,
		Done:      false,
		Date:      date,
		Notes:     note,
		IsSomeday: someday,
		RecurrenceRule: nil,
	}
	*l = append(*l, item)
	return item
}

func (l *List) DeleteTask(filename, taskId string) error {
	ls := *l
	found := false

	for i, item := range ls {
		// Compare the string version of the UUID
		if item.ID.String() == taskId {
			// Remove the item by joining everything before it and everything after it
			*l = append(ls[:i], ls[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("task with ID %s not found", taskId)
	}

	// Save the changes to the JSON file
	return l.Save(filename)
}

func (l *List) Save(filename string) error {
	data, _ := json.Marshal(l)
	return os.WriteFile(filename, data, 0644)
}

func (l *List) Load(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, l)
}

func (l *List) GetTaskDetails(taskId string) (Item, error) {
	// Loop through the current list in memory
	for _, item := range *l {
		if item.ID.String() == taskId {
			return item, nil
		}
	}
	return Item{}, fmt.Errorf("task with ID %s not found", taskId)
}

func (l *List) UpdateTask(id uuid.UUID, newTaskName, newNotes string) {
	for i := range *l {
		if (*l)[i].ID == id {
			(*l)[i].Task = newTaskName
			(*l)[i].Notes = newNotes
			return
		}
	}
}

func (l *List) MoveTask(id uuid.UUID, newDate time.Time)  {
	for i := range *l {
		if (*l)[i].ID == id {
			(*l)[i].Date = newDate
			return
		}
	}
}

func (l *List) UpdateRecurrenceRule(id uuid.UUID, rule RecurrenceRule) {
	for i := range *l {
		if (*l)[i].ID == id {
			if rule.Freq == None {
				(*l)[i].RecurrenceRule = nil
			} else {
				(*l)[i].RecurrenceRule = &rule
			}
			return
		}
	}
}