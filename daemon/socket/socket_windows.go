// +build windows

package socket

import (
	"net"

	"github.com/natefinch/npipe"
)

func makeSocket(transportAddress string, groupShared bool) (net.Listener, error) {
	// We're using the npipe functionality here to create a net.Listener for
	// Named Pipes on Windows.
	// The security model for a `CreateNamedPipe` allows us to set an optional
	// SecurityAttribute, however, npipe doesn't allow us to do this. It is
	// however not setting it's own, resulting on the default behaviour:
	//
	// From the Microsoft Documentation (https://msdn.microsoft.com/en-us/library/windows/desktop/aa365150(v=vs.85).aspx)
	// A pointer to a SECURITY_ATTRIBUTES structure that specifies a security
	// descriptor for the new named pipe and determines whether child processes
	// can inherit the returned handle. If lpSecurityAttributes is NULL, the
	// named pipe gets a default security descriptor and the handle cannot be
	// inherited. The ACLs in the default security descriptor for a named pipe
	// grant full control to the LocalSystem account, administrators, and the
	// creator owner. They also grant read access to members of the Everyone
	// group and the anonymous account.
	return npipe.Listen(transportAddress)
}
