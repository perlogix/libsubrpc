package subrpc

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/rpc"
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
			}, o.ExePath, "-socket", o.SockPath),
			SockPath:  o.SockPath,
			Terminate: make(chan bool),
		}
		m.Procs[o.Name].CMD.Env = append(m.Procs[o.Name].CMD.Env, o.Env...)
	}
	return nil
}

// StartProcess starts all of the sub processes
func (m *Manager) StartProcess(name string) error {
	if p, ok := m.Procs[name]; ok {
		if !p.Running {
			var err error
			p.StatusChan = p.CMD.Start()
			for i := 0; i <= 10; i++ {
				if p.CMD.Status().StartTs != 0 {
					break
				}
				time.Sleep(250 * time.Millisecond)
			}
			p.PID = p.CMD.Status().PID
			p.Running = true
			p.RPC, err = rpc.Dial(p.SockPath)
			if err != nil {
				return err
			}
			go m.supervise(p)
			go m.log(p)
			return nil
		}
		return fmt.Errorf("process %s is already running", name)
	}
	return fmt.Errorf("process with name %s does not exist", name)
}

// StartAllProcess starts all procs in the manager
func (m *Manager) StartAllProcess() []error {
	errs := []error{}
	for _, v := range m.Procs {
		err := m.StartProcess(v.Name)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errs
}

// RestartProcess restarts a process
func (m *Manager) RestartProcess(name string) error {
	if p, ok := m.Procs[name]; ok {
		err := m.StopProcess(name)
		if err != nil {
			return err
		}
		p.CMD = p.CMD.Clone()
		err = m.StartProcess(name)
		if err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("process with name %s does not exist", name)
}

// StopProcess stopps a process by name
func (m *Manager) StopProcess(name string) error {
	if p, ok := m.Procs[name]; ok {
		p.Running = false
		p.RPC.Close()
		p.Terminate <- true
		err := os.Remove(p.SockPath)
		if err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("process with name %s does not exist", name)
}

// StopAll stopps all procs
func (m *Manager) StopAll() []error {
	errs := []error{}
	for _, p := range m.Procs {
		err := m.StopProcess(p.Name)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) != 0 {
		return errs
	}
	return nil
}

func (m *Manager) supervise(proc *ProcessInfo) {
	for {
		select {
		case <-proc.Terminate:
			proc.CMD.Stop()
			return
		case <-proc.StatusChan:
			err := m.RestartProcess(proc.Name)
			if err != nil {
				fmt.Println(err)
			}
			return
		default:
			st := proc.CMD.Status()
			if st.Complete == false && st.Error != nil {
				err := m.RestartProcess(proc.Name)
				if err != nil {
					fmt.Println(err)
				}
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (m *Manager) log(proc *ProcessInfo) {
	t := time.NewTicker(100 * time.Millisecond)
	for range t.C {
		select {
		case <-proc.Terminate:
			return
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

// Call function calls an RPC service with the supplied "name:function" string
func (m *Manager) Call(urn string, dst interface{}, args ...interface{}) error {
	u := strings.Split(urn, ":")
	if len(u) != 2 {
		return fmt.Errorf("URN must be in format <name>:<function>")
	}
	if p, ok := m.Procs[u[0]]; ok {
		err := p.RPC.Call(&dst, u[1], args...)
		if err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("service with name %s does not exist", u[0])
}
