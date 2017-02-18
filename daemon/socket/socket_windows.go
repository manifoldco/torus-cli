// +build windows

package socket

import (
	"net"

	"github.com/natefinch/npipe"
)

func makeSocket(transportAddress string, groupShared bool) (net.Listener, error) {
	return npipe.Listen(transportAddress)
}
