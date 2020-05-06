package subrpc

import (
	"encoding/base64"
	"flag"
	"net"
	"os"

	"github.com/ethereum/go-ethereum/rpc"
)

// Process type represents an RPC service
type Process struct {
	SockPath   string
	Config     []byte
	Env        []string
	RPC        *rpc.Server
	Token      string
	ServerSock string
	Srv        *rpc.Client
}

// NewProcess function
func NewProcess() *Process {
	s := flag.String("socket", "", "Socket to bind to")
	c := flag.String("config", "", "Config from plugin manifest")
	t := flag.String("token", "", "Option for passing in a trust token")
	u := flag.String("serversocket", "", "Socket the server is listening on")
	flag.Parse()
	config, err := base64.StdEncoding.DecodeString(*c)
	if err != nil {
		panic(err)
	}
	p := &Process{
		Env:        os.Environ(),
		SockPath:   *s,
		Config:     config,
		RPC:        rpc.NewServer(),
		Token:      *t,
		ServerSock: *u,
	}
	srv, err := rpc.Dial(p.ServerSock)
	if err != nil {
		panic(err)
	}
	p.Srv = srv
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

// Call function
func (p *Process) Call(method string, dst interface{}, args ...interface{}) error {
	err := p.Srv.Call(&dst, method, args...)
	if err != nil {
		return err
	}
	return nil
}

type rpcPing struct{}

func (r *rpcPing) Ping() string {
	return "pong"
}
