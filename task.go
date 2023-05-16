package tfcheck

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// TaskFunc is the signature for functions that can be executed in a Task.
type TaskFunc func(io.Writer) error

// NewTask returns the initial model for a task.
func NewTask(name string, fn TaskFunc) Task {
	s := spinner.New(spinner.WithSpinner(spinner.MiniDot), spinner.WithStyle(taskSpinnerStyle))
	return Task{
		id:      nextID(),
		name:    name,
		fn:      fn,
		buf:     &Buffer{},
		status:  StatusPending,
		spinner: s,
	}
}

// Task ...
type Task struct {
	id      int
	jobID   int
	name    string
	fn      TaskFunc
	buf     *Buffer
	status  Status
	spinner spinner.Model
}

// taskInitMsg is sent when a task starts.
type taskInitMsg struct {
	id    int
	jobID int
}

// taskDoneMsg is returned when a task is done.
type taskDoneMsg struct {
	id    int
	jobID int
	err   error
}

func (t Task) Init() tea.Cmd {
	init := func() tea.Msg { return taskInitMsg{id: t.id, jobID: t.jobID} }
	return tea.Batch(init, t.spinner.Tick)
}

func (t Task) Update(msg tea.Msg) (Task, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		t.spinner, cmd = t.spinner.Update(msg)
		return t, cmd
	case taskInitMsg:
		if msg.id != t.id {
			return t, nil
		}
		t.status = StatusRunning
		return t, func() tea.Msg {
			err := t.fn(t.buf)
			return taskDoneMsg{id: t.id, jobID: t.jobID, err: err}
		}
	case taskDoneMsg:
		if msg.id != t.id {
			return t, nil
		}
		if msg.err != nil {
			t.status = StatusFailed
		} else {
			t.status = StatusSucceeded
		}
	}
	return t, nil
}

func (t Task) View() string {
	var s string
	switch t.status {
	case StatusPending:
		s = fmt.Sprintf("  %s %s", t.spinner.View(), t.name)
	case StatusRunning:
		s = fmt.Sprintf("  %s %s\n", t.spinner.View(), t.name)
		for _, line := range t.buf.Tail(3) {
			s += fmt.Sprintf("    %s", line)
		}
	case StatusSucceeded:
		s = fmt.Sprintf("  %s %s", t.status, t.name)
	case StatusFailed:
		s = fmt.Sprintf("  %s %s\n", t.status, t.name)
		for _, line := range t.buf.Lines() {
			s += fmt.Sprintf("    %s", line)
		}
	}
	return strings.TrimRight(s, "\n")
}
