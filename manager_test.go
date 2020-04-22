package subrpc

import "testing"

// TestNewManager func
func TestNewManager(t *testing.T) {
	m := NewManager()
	if m.SockPath == "" {
		t.Error("SockPath not set")
	}
	if len(m.Procs) != 0 {
		t.Error("Procs initialized with contents")
	}
	if m.OutBuffer == nil || m.ErrBuffer == nil {
		t.Error("buffers must be instantiated")
	}
}
