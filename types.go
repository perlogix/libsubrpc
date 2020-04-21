package subrpc

import "github.com/go-cmd/cmd"

// ProcessOptions allows for passing process options to NewProcess
type ProcessOptions struct {
	Name     string
	ExePath  string
	SockPath string
}

// ProcessInfo holds information about running processes
type ProcessInfo struct {
	Name       string
	CMD        *cmd.Cmd
	Options    ProcessOptions
	Running    bool
	StatusChan <-chan cmd.Status
	Terminate  chan bool
	PID        int
	SockPath   string
}
