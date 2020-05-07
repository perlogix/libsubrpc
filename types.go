package subrpc

import (
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/yeticloud/airboss"
)

// ProcessOptions allows for passing process options to NewProcess
type ProcessOptions struct {
	Name     string
	Type     string
	Config   map[string]interface{}
	Handler  interface{}
	ExePath  string
	SockPath string
	Env      map[string]string
	Token    string
}

// ProcessInfo holds information about running processes
type ProcessInfo struct {
	Name      string
	Type      string
	Config    map[string]interface{}
	Token     string
	Handler   interface{}
	CMD       *airboss.Subprocess
	Options   ProcessOptions
	Running   bool
	Terminate chan bool
	PID       int
	SockPath  string
	RPC       *rpc.Client
}
