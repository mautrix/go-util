package exhttp

import (
	"errors"
	"net"
	"syscall"

	"golang.org/x/net/http2"
)

func IsNetworkError(err error) bool {
	if errno := syscall.Errno(0); errors.As(err, &errno) {
		// common errnos for network related operations
		return errno == syscall.ENETDOWN ||
			errno == syscall.ENETUNREACH ||
			errno == syscall.ENETRESET ||
			errno == syscall.ECONNABORTED ||
			errno == syscall.ECONNRESET ||
			errno == syscall.ENOBUFS ||
			errno == syscall.ETIMEDOUT ||
			errno == syscall.ECONNREFUSED ||
			errno == syscall.EHOSTDOWN ||
			errno == syscall.EHOSTUNREACH
	} else if netError := net.Error(nil); errors.As(err, &netError) {
		return true
	} else if errors.As(err, &http2.StreamError{}) {
		return true
	}

	return false
}
