//go:build !wasip1

package rabtap

import (
	"net"
)

var Dialer = net.Dial
