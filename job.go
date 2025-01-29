package tfcheck

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/stopwatch"
	tea "github.com/charmbracelet/bubbletea"
)

func NewJob(dir, tflintConfig string) Job {
	job := Job{
		id:        nextID(),
		name:      dir,
		status:    StatusPending,
		stopwatch: stopwatch.NewWithInterval(time.Millisecond),
	}

	tf := newTerraformRunner(dir, tflintConfig)
	job.tasks = []Task{
		NewTask("terraform:fmt", tf.fmt),
		NewTask("terraform:init", tf.init),
		NewTask("terraform:validate", tf.validate),
	}
	_, err := exec.LookPath("tflint")
	if err != nil && !errors.Is(err, exec.ErrNotFound) {
		panic(fmt.Errorf("failed to look up 'tflint' in PATH: %v", err))
	}
	if err == nil {
		job.tasks = append(job.tasks, NewTask("terraform:tflint", tf.tflint))
	}
	for i := range job.tasks {
		job.tasks[i].jobID = job.id
	}
	return job
}

type Job struct {
	id          int
	name        string
	tasks       []Task
	currentTask int
	status      Status
	stopwatch   stopwatch.Model
	done        bool
	failed      bool
}

type jobInitMsg struct {
	id int
}

type jobDoneMsg struct {
	id     int
	failed bool
}

func (j Job) Init() tea.Cmd {
	init := func() tea.Msg { return jobInitMsg{id: j.id} }
	return tea.Batch(init, j.stopwatch.Init())
}

func (j Job) Update(msg tea.Msg) (Job, tea.Cmd) {
	switch msg := msg.(type) {
	case stopwatch.StartStopMsg, stopwatch.TickMsg:
		var cmd tea.Cmd
		j.stopwatch, cmd = j.stopwatch.Update(msg)
		return j, cmd
	case jobInitMsg:
		if msg.id != j.id {
			return j, nil
		}
		j.status = StatusRunning
		return j, j.tasks[0].Init()
	case jobDoneMsg:
		if msg.id != j.id {
			return j, nil
		}
		j.done = true
		if j.failed {
			j.status = StatusFailed
		} else {
			j.status = StatusSucceeded
		}
		return j, tea.Printf(j.View())
	case taskDoneMsg:
		if msg.jobID != j.id {
			return j, nil
		}
		if msg.err != nil {
			j.failed = true // Consider job failed if any task has failed
		}

		var cmds []tea.Cmd
		j.tasks, cmds = j.updateTasks(msg)

		j.currentTask++
		if j.currentTask >= len(j.tasks) {
			// We are done processing and need to send a done msg.
			done := func() tea.Msg { return jobDoneMsg{id: j.id, failed: j.failed} }
			return j, tea.Sequence(tea.Batch(cmds...), done)
		}
		cmds = append(cmds, j.tasks[j.currentTask].Init())
		return j, tea.Batch(cmds...)
	default:
		var cmds []tea.Cmd
		j.tasks, cmds = j.updateTasks(msg)
		return j, tea.Batch(cmds...)
	}
}

func (j Job) View() string {
	var s strings.Builder
	switch j.status {
	case StatusPending, StatusRunning:
		s.WriteString(fmt.Sprintf("%s (%s)", jobNameStyle.SetString(j.name), j.stopwatch.View()))
	case StatusSucceeded, StatusFailed:
		s.WriteString(fmt.Sprintf("%s %s (%s)", j.status, jobNameStyle.SetString(j.name), j.stopwatch.View()))
	}
	for _, t := range j.tasks {
		s.WriteString("\n" + t.View())
	}
	return s.String()
}

func (j Job) updateTasks(msg tea.Msg) (tasks []Task, cmds []tea.Cmd) {
	for _, task := range j.tasks {
		t, cmd := task.Update(msg)
		tasks = append(tasks, t)
		cmds = append(cmds, cmd)
	}
	return tasks, cmds
}
