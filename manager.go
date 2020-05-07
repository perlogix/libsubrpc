package subrpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/yeticloud/airboss"
)

// Manager type instantiates a new Manager instance
type Manager struct {
	ServerPort int
	Procs      map[string]map[string]*ProcessInfo
	OutBuffer  *bytes.Buffer
	ErrBuffer  *bytes.Buffer
	Errors     chan error
	Metrics    chan Metrics
	RPC        *rpc.Server
	mgr        *airboss.ProcessManager
}

// Metrics type
type Metrics struct {
	URN      string
	CallTime time.Duration
	Error    bool
}

// NewManager function returns a new instance of the Manager object
func NewManager() (*Manager, error) {
	m := &Manager{
		ServerPort: 0,
		Procs:      make(map[string]map[string]*ProcessInfo),
		OutBuffer:  bytes.NewBuffer([]byte{}),
		ErrBuffer:  bytes.NewBuffer([]byte{}),
		Errors:     make(chan error, 64),
		Metrics:    make(chan Metrics, 1024),
		RPC:        rpc.NewServer(),
		mgr:        airboss.NewProcessManager(),
	}
	conn, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		return nil, err
	}
	m.ServerPort, err = strconv.Atoi(strings.Split(conn.Addr().String(), ":")[1])
	if err != nil {
		return m, nil
	}
	go m.RPC.ServeListener(conn)
	m.RPC.RegisterName("ping", new(ManagerService))
	return m, nil
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
		if o.Port == 0 {
			o.Port = findAPort()
		}
		byt, err := json.Marshal(o.Config)
		if err != nil {
			return err
		}
		opts := ProcessInput{
			Port:       o.Port,
			ServerPort: m.ServerPort,
			Config:     byt,
			Token:      o.Token,
		}
		bopts, err := json.Marshal(opts)
		if err != nil {
			return err
		}
		if _, ok := m.Procs[o.Type]; !ok {
			m.Procs[o.Type] = map[string]*ProcessInfo{}
		}
		p, err := m.mgr.Fork(o.ExePath)
		if err != nil {
			return err
		}
		_, err = p.Stdin.Write(bopts)
		if err != nil {
			return err
		}
		m.Procs[o.Type][o.Name] = &ProcessInfo{
			Name:      o.Name,
			Options:   o,
			Running:   false,
			CMD:       p,
			Port:      o.Port,
			Terminate: make(chan bool),
		}
		m.Procs[o.Type][o.Name].CMD.Env = o.Env
	}
	return nil
}

// StartProcess starts all of the sub processes
func (m *Manager) StartProcess(name string, typ string) error {
	if p, ok := m.Procs[typ][name]; ok {
		if !p.Running {
			var err error
			_, err = p.CMD.Start()
			if err != nil {
				return err
			}
			go m.log(p)
			p.PID = p.CMD.PID
			p.Running = true
			p.RPC, err = rpc.DialHTTP(fmt.Sprintf("http://127.0.0.1:%v", p.Port))
			if err != nil {
				return err
			}
			return nil
		}
		return fmt.Errorf("process %s is already running", name)
	}
	return fmt.Errorf("process with name %s does not exist", name)
}

// StartAllProcess starts all procs in the manager
func (m *Manager) StartAllProcess() []error {
	errs := []error{}
	for k, v := range m.Procs {
		for _, j := range v {
			err := m.StartProcess(j.Name, k)
			if err != nil {
				errs = append(errs, err)
			}
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errs
}

// RestartProcess restarts a process
func (m *Manager) RestartProcess(name string, typ string) error {
	if p, ok := m.Procs[typ][name]; ok {
		_, err := p.CMD.Restart()
		if err != nil {
			return err
		}
		p.PID = p.CMD.PID
		return nil
	}
	return fmt.Errorf("process with name %s does not exist", name)
}

// StopProcess stopps a process by name
func (m *Manager) StopProcess(name string, typ string) error {
	if p, ok := m.Procs[typ][name]; ok {
		err := p.CMD.Stop()
		if err != nil {
			return err
		}
		p.Running = false
		p.RPC.Close()
		return nil
	}
	return fmt.Errorf("process with name %s does not exist", name)
}

// StopAll stopps all procs
func (m *Manager) StopAll() []error {
	errs := []error{}
	for k, v := range m.Procs {
		for _, j := range v {
			err := m.StopProcess(j.Name, k)
			if err != nil {
				errs = append(errs, err)
			}
		}
	}
	if len(errs) != 0 {
		return errs
	}
	return nil
}

func (m *Manager) log(proc *ProcessInfo) {
	t := time.NewTicker(100 * time.Millisecond)
	for range t.C {
		select {
		case <-proc.Terminate:
			return
		case e := <-proc.CMD.Errors:
			m.Errors <- e
		}
		if proc.CMD.Stdout.Len() > 0 {
			_, err := m.OutBuffer.ReadFrom(proc.CMD.Stdout)
			if err != nil {
				m.Errors <- err
			}
		}
		if proc.CMD.Stderr.Len() > 0 {
			_, err := m.ErrBuffer.ReadFrom(proc.CMD.Stderr)
			if err != nil {
				m.Errors <- err
			}
		}
	}
}

// Call function calls an RPC service with the supplied "name:function" string
func (m *Manager) Call(urn string, dst interface{}, args ...interface{}) error {
	start := time.Now()
	u := strings.Split(urn, ":")
	if len(u) != 3 {
		m.Metrics <- Metrics{
			URN:      urn,
			Error:    true,
			CallTime: time.Now().Sub(start),
		}
		return fmt.Errorf("URN must be in format <type>:<name>:<function>")
	}
	if p, ok := m.Procs[u[0]][u[1]]; ok {
		err := p.RPC.Call(&dst, u[2], args...)
		if err != nil {
			m.Metrics <- Metrics{
				URN:      urn,
				Error:    true,
				CallTime: time.Now().Sub(start),
			}
			return err
		}
		m.Metrics <- Metrics{
			URN:      urn,
			Error:    false,
			CallTime: time.Now().Sub(start),
		}
		return nil
	}
	m.Metrics <- Metrics{
		URN:      urn,
		Error:    true,
		CallTime: time.Now().Sub(start),
	}
	return fmt.Errorf("service with name %s does not exist", u[0])
}

// ManagerService type
type ManagerService struct{}

// Ping function
func (ms *ManagerService) Ping() string {
	return "pong"
}

func findAPort() int {
	start := 8000
	for {
		conn, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%v", start))
		defer conn.Close()
		if err != nil {
			start++
		} else {
			return start
		}
	}
}
