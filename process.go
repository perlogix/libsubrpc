package subrpc

import (
	"encoding/json"
	"flag"
	"net"
	"os"

	"github.com/ethereum/go-ethereum/rpc"
)

// Process type represents an RPC service
type Process struct {
	SockPath string
	Config   json.RawMessage
	Env      []string
	RPC      *rpc.Server
}

// NewProcess function
func NewProcess() *Process {
	s := flag.String("socket", "", "Socket to bind to")
	c := flag.String("config", "", "Config from plugin manifest")
	flag.Parse()
	p := &Process{
		Env:      os.Environ(),
		SockPath: *s,
		Config:   json.RawMessage(*c),
		RPC:      rpc.NewServer(),
	}
	p.RPC.RegisterName("ping", new(rpcPing))
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

type rpcPing struct{}

func (r *rpcPing) Ping() string {
	return "pong"
}
