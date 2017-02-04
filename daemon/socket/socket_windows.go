// +build windows

package socket

import "net"

func makeSocket(socketPath string, groupShared bool) (net.Listener, error) {
	return net.Listen("tcp", ":50")
}
