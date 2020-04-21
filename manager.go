package subrpc

import (
	"bytes"
	"fmt"
	"time"

	"github.com/go-cmd/cmd"
	"github.com/google/uuid"
)

// Manager type instantiates a new Manager instance
type Manager struct {
	SockPath  string
	Procs     map[string]*ProcessInfo
	OutBuffer *bytes.Buffer
	ErrBuffer *bytes.Buffer
}

// NewManager function returns a new instance of the Manager object
func NewManager() *Manager {
	return &Manager{
		SockPath:  fmt.Sprintf("/tmp/rpc-%s", uuid.New().String()),
		Procs:     map[string]*ProcessInfo{},
		OutBuffer: bytes.NewBuffer([]byte{}),
		ErrBuffer: bytes.NewBuffer([]byte{}),
	}
}

// NewProcess instantiates new processes
func (m *Manager) NewProcess(options ...ProcessOptions) error {
	for _, o := range options {
		if o.Name == "" {
			return fmt.Errorf("name cannot be blank")
		}
		if o.ExePath == "" {
			return fmt.Errorf("exepath cannot be blank")
		}
		if o.SockPath == "" {
			o.SockPath = fmt.Sprintf("/tmp/rpc-%s", uuid.New().String())
		}
		m.Procs[o.Name] = &ProcessInfo{
			Name:    o.Name,
			Options: o,
			Running: false,
			CMD: cmd.NewCmdOptions(cmd.Options{
				Buffered:  false,
				Streaming: true,
			}, o.ExePath, "--socket", o.SockPath),
			SockPath: o.SockPath,
		}
	}
	return nil
}

// StartProcess starts all of the sub processes
func (m *Manager) StartProcess(name string) error {
	if p, ok := m.Procs[name]; ok {
		if !p.Running {
			p.StatusChan = p.CMD.Start()
			p.PID = p.CMD.Status().PID
			p.Running = true
			go m.supervise(p)
			go m.log(p)
			return nil
		}
		return fmt.Errorf("process %s is already running", name)
	}
	return fmt.Errorf("process with name %s does not exist", name)
}

// StartAllProcess starts all procs in the manager
func (m *Manager) StartAllProcess() error {
	for _, v := range m.Procs {
		if !v.Running {
			v.StatusChan = v.CMD.Start()
			v.PID = v.CMD.Status().PID
			v.Running = true
			go m.supervise(v)
			go m.log(v)
		}
	}
	return nil
}

// Stop stopps a process by name
func (m *Manager) Stop(name string) error {
	if p, ok := m.Procs[name]; ok {
		if p.Running {
			p.Terminate <- true
			return nil
		}
		return fmt.Errorf("process %s is not running, cannot stop", name)
	}
	return fmt.Errorf("process with name %s does not exist", name)
}

// StopAll stopps all procs
func (m *Manager) StopAll(name string) {
	for _, p := range m.Procs {
		p.Terminate <- true
	}
}

func (m *Manager) supervise(proc *ProcessInfo) {
	for {
		select {
		case <-proc.StatusChan:
			m.StartProcess(proc.Name)
			return
		case <-proc.Terminate:
			proc.CMD.Stop()
			return
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (m *Manager) log(proc *ProcessInfo) {
	t := time.NewTicker(100 * time.Millisecond)
	for range t.C {
		select {
		case line := <-proc.CMD.Stdout:
			_, err := m.OutBuffer.WriteString(line)
			if err != nil {
				fmt.Println(err)
			}
		case line := <-proc.CMD.Stderr:
			_, err := m.ErrBuffer.WriteString(line)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}
