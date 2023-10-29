//go:build wasip1

package rabtap

import (
	"github.com/stealthrocket/net/wasip1"
)

var Dialer = wasip1.Dial
