package tfcheck

import (
	"fmt"
	"io"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-isatty"
)

var (
	jobNameStyle     = lipgloss.NewStyle().Bold(true)
	taskSpinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	warningStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#E0AF68"))
	failedStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("211"))
	succeededStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
)

const (
	StatusPending Status = iota
	StatusRunning
	StatusWarning
	StatusFailed
	StatusSucceeded
)

type Status int

func (s Status) String() string {
	return map[Status]string{
		StatusWarning:   warningStyle.SetString("").String(),
		StatusFailed:    failedStyle.SetString("✗").String(),
		StatusSucceeded: succeededStyle.SetString("✓").String(),
	}[s]
}

type Config struct {
	Directories   []string
	TFLintConfig  string
	MaxInParallel int
	NoTUI         bool
}

func Run(cfg Config) error {
	var jobs []Job
	for _, d := range cfg.Directories {
		jobs = append(jobs, NewJob(d, cfg.TFLintConfig))
	}

	if cfg.MaxInParallel == 0 {
		cfg.MaxInParallel = len(jobs)
	}

	var (
		writer = io.Discard
		opts   = []tea.ProgramOption{tea.WithOutput(os.Stdout)}
	)

	isTTY := isatty.IsTerminal(os.Stdout.Fd())
	if !isTTY || cfg.NoTUI {
		// Limit parallel executions (since we can't visualise it without the TUI)
		cfg.MaxInParallel = 1

		// Disable bubbletea renderer (View() and tea.Print statements)
		opts = append(opts, tea.WithoutRenderer(), tea.WithInput(nil))

		// Use os.Stdout for the fmt.Fprint statements
		writer = os.Stdout
	}

	semaphore := make(chan struct{}, cfg.MaxInParallel)
	defer close(semaphore)

	m := model{jobs: jobs, sem: semaphore, writer: writer}
	p := tea.NewProgram(m, opts...)

	r, err := p.Run()
	if err != nil {
		return err
	}

	if model, ok := r.(model); ok && model.jobsFailed > 0 {
		return fmt.Errorf("%d job(s) failed", model.jobsFailed)
	}

	return nil
}

type model struct {
	jobs       []Job
	sem        chan struct{}
	writer     io.Writer
	width      int
	height     int
	currentJob int
	jobsDone   int
	jobsFailed int
}

type initMsg struct{}

func (m model) Init() tea.Cmd {
	return func() tea.Msg { return initMsg{} }
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	case tea.KeyMsg:
		// Ctrl-c should quite the program
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		return m, nil
	case initMsg:
		var cmds []tea.Cmd
		for i, job := range m.jobs {
			select {
			case m.sem <- struct{}{}:
				m.currentJob = i
				cmds = append(cmds, job.Init())
			default:
				break
			}
		}
		return m, tea.Batch(cmds...)
	case jobInitMsg:
		var cmds []tea.Cmd
		m.jobs, cmds = m.updateJobs(msg)

		job := m.findJob(msg.id)
		fmt.Fprintln(m.writer, jobNameStyle.SetString(job.name).String())

		return m, tea.Batch(cmds...)
	case taskDoneMsg:
		var cmds []tea.Cmd
		m.jobs, cmds = m.updateJobs(msg)

		job := m.findJob(msg.jobID)
		var task Task
		for _, t := range job.tasks {
			if t.id == msg.id {
				task = t
			}
		}
		fmt.Fprintf(m.writer, task.View()+"\n")

		return m, tea.Batch(cmds...)
	case jobDoneMsg:
		<-m.sem
		if msg.failed {
			m.jobsFailed++ // Keep track of how many jobs have failed
		}

		var cmds []tea.Cmd
		m.jobs, cmds = m.updateJobs(msg)

		m.jobsDone++
		if m.jobsDone >= len(m.jobs) {
			return m, tea.Sequence(tea.Batch(cmds...), tea.Quit)
		}

		if m.currentJob < len(m.jobs)-1 {
			select {
			case m.sem <- struct{}{}:
				m.currentJob++
				cmds = append(cmds, m.jobs[m.currentJob].Init())
			default:
				break
			}
		}
		return m, tea.Batch(cmds...)
	default:
		var cmds []tea.Cmd
		m.jobs, cmds = m.updateJobs(msg)
		return m, tea.Batch(cmds...)
	}
}

func (m model) View() string {
	var s strings.Builder
	for _, j := range m.jobs {
		if j.done {
			continue // Skip jobs that are already done
		}
		s.WriteString("\n" + j.View())
	}
	return s.String()
}

func (m model) findJob(id int) Job {
	for _, job := range m.jobs {
		if job.id == id {
			return job
		}
	}
	panic(fmt.Errorf("job not found: %d", id))
}

func (m model) updateJobs(msg tea.Msg) (jobs []Job, cmds []tea.Cmd) {
	for _, job := range m.jobs {
		j, cmd := job.Update(msg)
		jobs = append(jobs, j)
		cmds = append(cmds, cmd)
	}
	return jobs, cmds
}
