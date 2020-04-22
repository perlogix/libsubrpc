package subrpc

import (
	"github.com/go-cmd/cmd"
	"github.com/ybbus/jsonrpc"
)

// ProcessOptions allows for passing process options to NewProcess
type ProcessOptions struct {
	Name     string
	Handler  interface{}
	ExePath  string
	SockPath string
	Env      []string
}

// ProcessInfo holds information about running processes
type ProcessInfo struct {
	Name       string
	Handler    interface{}
	CMD        *cmd.Cmd
	Options    ProcessOptions
	Running    bool
	StatusChan <-chan cmd.Status
	Terminate  chan bool
	PID        int
	SockPath   string
	RPC        jsonrpc.RPCClient
}
