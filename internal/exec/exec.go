package exec

import (
	"context"
	"os/exec"
	"sync"
)

type Status int

const (
	Pending Status = iota
	Running
	Failed
	Succeeded
)

func New(dir string) *Executor {
	return &Executor{dir: dir, buf: &Buffer{}}
}

type Executor struct {
	dir  string
	buf  *Buffer
	cmds []command

	mu     sync.Mutex
	status Status
}

type command struct {
	name string
	args []string
}

func (e *Executor) Command(name string, args ...string) {
	e.cmds = append(e.cmds, command{name: name, args: args})
}

func (e *Executor) Run(ctx context.Context) error {
	e.setStatus(Running)

	for _, cmd := range e.cmds {
		c := exec.CommandContext(ctx, cmd.name, cmd.args...)

		c.Dir = e.dir
		c.Stdout = e.buf
		c.Stderr = e.buf

		if err := c.Run(); err != nil {
			e.setStatus(Failed)
			return err
		}
	}

	e.setStatus(Succeeded)
	return nil
}

func (e *Executor) Status() Status {
	e.mu.Lock()
	defer e.mu.Unlock()

	return e.status
}

func (e *Executor) setStatus(s Status) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.status = s
}

func (e *Executor) Output() string {
	return ""
}
