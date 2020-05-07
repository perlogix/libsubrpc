package subrpc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"

	"github.com/ethereum/go-ethereum/rpc"
)

// Process type represents an RPC service
type Process struct {
	Port       int
	Config     []byte
	Env        []string
	RPC        *rpc.Server
	Token      string
	ServerPort int
	Srv        *rpc.Client
}

// NewProcess function
func NewProcess() *Process {
	f, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}
	var opts ProcessInput
	err = json.Unmarshal(f, &opts)
	if err != nil {
		panic(err)
	}
	p := &Process{
		Env:        os.Environ(),
		Port:       opts.Port,
		Config:     opts.Config,
		RPC:        rpc.NewServer(),
		Token:      opts.Token,
		ServerPort: opts.ServerPort,
	}
	srv, err := rpc.DialHTTP(fmt.Sprintf("127.0.0.1:%v", opts.ServerPort))
	if err != nil {
		panic(err)
	}
	p.Srv = srv
	p.RPC.RegisterName("ping", new(rpcPing))
	return p
}

// Start starts a new process instance
func (p *Process) Start() error {
	conn, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%v", p.Port))
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
