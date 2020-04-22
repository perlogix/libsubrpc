# libsubrpc
libsubrpc provides subprocess management and RPC communication in a single package

## Purpose
The purpose is to provide a framework for allowing a plugin/dll-style system for Go given its lack of plugin or dynamic symbol loading. 

## Basic Usage

#### Main App
```go

manager := subrpc.NewManager() // make a new Manager instance

// add a few subprocesses to the manager
manager.NewProcess(subrpc.ProcessOptions{
    // Name should be unique
    Name: "foo",

    // Absolute path is best
    Exepath: "/path/to/process/binary/foo",
})

manager.NewProcess(subrpc.ProcessOptions{
    // Name should be unique
    Name: "bar",
    
    // Absolute path is best
    Exepath: "/path/to/process/binary/bar",

    // You can set a custom socket path if you want, otherwise
    // a random one is generated
    SockPath: "/path/to/custom/socket",

    // Custom environment variables to pass into the process
    Env: []string{
        "MYVAR=myvalue",
        "OTHERVAR=other_value",
    }
})

// Read Stdin * Stderr from our subprocesses for logging purposes
// the Manager struct exposes OutBuffer and ErrBuffer buffers that
// combine the stdout and stderr of all processes, respectively.
go func() {
    for {
		l, err := manager.OutBuffer.ReadString('\n')
		if err != nil && err != io.EOF {
			fmt.Println(err)
		}
		if l != "" {
			fmt.Println(l)
		}
		l, err = manager.ErrBuffer.ReadString('\n')
		if err != nil && err != io.EOF {
			fmt.Print(err)
		}
		if l != "" {
			fmt.Println(l)
		}
		time.Sleep(1 * time.Second)
	}
}()

// Start all the processes (returns error)
_ = manager.StartAllProcess()


// Create a variable to put the results into
var result json.RawMessage

// Call the `foo` process and request the `test_service` endpoint.
// Endpoints are underscore delimited (see subprocess implementation)
_ = manager.Call("foo:test_service", &result)

// Call the `bar` process with some args and save it to `result`
_ = manager.Call("bar:other_service", &result, "something", 1, 2, 5)

// Stops all running processes
manager.StopAll()

```

#### Subprocess App
```go

// Define a service here
type Test struct {}

// Service handler; to call this you would call `test_service`
func (t *Test) Service() string {
    return "this is a test"
}

proc := subrpc.Process()

// add the Test object as an RPC handler under the `test_` prefix
proc.AddFunction("test", new(Test))

proc.Start() // start listening for connections; blocking


```