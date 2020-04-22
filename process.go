package subrpc

import (
	"flag"
	"net"
	"os"

	"github.com/ethereum/go-ethereum/rpc"
)

// Process type represents an RPC service
type Process struct {
	SockPath string
	Env      []string
	RPC      *rpc.Server
}

// NewProcess function
func NewProcess() *Process {
	p := &Process{
		Env:      os.Environ(),
		SockPath: *flag.String("socket", "", "Sets the socket to listen on"),
		RPC:      rpc.NewServer(),
	}
	return p
}

// Start starts a new process instance
func (p *Process) Start() error {
	conn, err := net.Listen("unix", p.SockPath)
	if err != nil {
		return err
	}
	return p.RPC.ServeListener(conn)
}

// AddFunction adds a function to the RPC handler
func (p *Process) AddFunction(name string, f interface{}) error {
	return p.RPC.RegisterName(name, f)
}

func rpcPing() string {
	return "pong"
}
