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
	Socket       string
	Config       []byte
	Env          []string
	RPC          *rpc.Server
	Token        string
	ServerSocket string
	Srv          *rpc.Client
}

// NewProcess function
func NewProcess() *Process {
	f, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		writeLog(err)
		panic(err)
	}
	var opts ProcessInput
	err = json.Unmarshal(f, &opts)
	if err != nil {
		writeLog(err)
		panic(err)
	}
	p := &Process{
		Env:          os.Environ(),
		Socket:       opts.Socket,
		Config:       opts.Config,
		RPC:          rpc.NewServer(),
		Token:        opts.Token,
		ServerSocket: opts.ServerSocket,
	}
	srv, err := rpc.Dial(p.ServerSocket)
	if err != nil {
		writeLog(err)
		panic(err)
	}
	p.Srv = srv
	err = p.RPC.RegisterName("ping", new(rpcPing))
	if err != nil {
		panic(err)
	}
	return p
}

// Start starts a new process instance
func (p *Process) Start() error {
	conn, err := net.Listen("unix", p.Socket)
	if err != nil {
		writeLog(err)
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
		writeLog(err)
		return err
	}
	return nil
}

type rpcPing struct{}

func (r *rpcPing) Ping() string {
	return "pong"
}

func writeLog(msg ...interface{}) {
	f, err := os.OpenFile("./libsubrpc.log", os.O_APPEND, 0600)
	if err != nil {
		fmt.Println(err)
	}
	_, err = f.WriteString(fmt.Sprint(msg...))
	if err != nil {
		fmt.Println(err)
	}
	err = f.Close()
	if err != nil {
		fmt.Println(err)
	}
}
