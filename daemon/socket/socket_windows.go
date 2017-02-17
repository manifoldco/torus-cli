// +build windows

package socket

import "net"

func makeSocket(transportAddress string, groupShared bool) (net.Listener, error) {
	return net.Listen("tcp", transportAddress)
}
