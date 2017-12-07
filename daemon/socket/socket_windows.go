// +build windows

package socket

import (
	"fmt"
	"net"
	"os/user"

	"github.com/Microsoft/go-winio"
)

func makeSocket(transportAddress string, groupShared bool) (net.Listener, error) {
	// Gets current user's SID 
	usr, err := user.Current()
	if err != nil {
		fmt.Printf("Error getting user SID: %s\n", transportAddress)
		return nil, err
	}

	// Configures pipe security descriptor to allow full control for system, administrators and current user. 
	// Security Descriptor String Format
	// https://msdn.microsoft.com/en-us/library/windows/desktop/aa379570(v=vs.85).aspx
	c := winio.PipeConfig{
		SecurityDescriptor: fmt.Sprintf("O:%s", usr.Uid) +
			fmt.Sprintf("G:%s", usr.Uid) +
			fmt.Sprintf("D:P(A;;FA;;;SY)(A;;FA;;;BA)(A;;FA;;;%s)", usr.Uid),
	}

	return winio.ListenPipe(transportAddress, &c)
}
