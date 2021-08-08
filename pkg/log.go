package rabtap

// we don't want to be this package dependent on a logging framework.
// So a logger is injected from the client code.
type Logger interface {
	Debugf(format string, a ...interface{})
	Errorf(format string, a ...interface{})
	Infof(format string, a ...interface{})
}
