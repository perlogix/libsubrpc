package subrpc

import (
	"flag"
	"os"

	"github.com/valyala/gorpc"
)

// Process type represents an RPC service
type Process struct {
	SockPath    string
	Env         []string
	RPCDispatch *gorpc.Dispatcher
	server      *gorpc.Server
}

// NewProcess function
func NewProcess() *Process {
	p := &Process{
		Env:         os.Environ(),
		SockPath:    *flag.String("socket", "", "Sets the socket to listen on"),
		RPCDispatch: gorpc.NewDispatcher(),
	}
	p.server = gorpc.NewUnixServer(p.SockPath, p.RPCDispatch.NewHandlerFunc())
	p.RPCDispatch.AddFunc("ping", rpcPing)
	return p
}

// Start starts a new process instance
func (p *Process) Start() {
	p.server.Serve()
}

// AddFunction adds a function to the RPC handler
func (p *Process) AddFunction(name string, f interface{}) {
	p.RPCDispatch.AddFunc(name, f)
}

func rpcPing() string {
	return "pong"
}
