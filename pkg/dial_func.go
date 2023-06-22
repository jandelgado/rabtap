package rabtap

import "net"

type DialFunc func(network, addr string) (net.Conn, error)
