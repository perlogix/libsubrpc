package subrpc

import (
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/go-cmd/cmd"
)

// ProcessOptions allows for passing process options to NewProcess
type ProcessOptions struct {
	Name     string
	Type     string
	Config   map[string]interface{}
	Handler  interface{}
	ExePath  string
	SockPath string
	Env      []string
	Token    string
}

// ProcessInfo holds information about running processes
type ProcessInfo struct {
	Name       string
	Type       string
	Config     map[string]interface{}
	Token      string
	Handler    interface{}
	CMD        *cmd.Cmd
	Options    ProcessOptions
	Running    bool
	StatusChan <-chan cmd.Status
	Terminate  chan bool
	PID        int
	SockPath   string
	RPC        *rpc.Client
}
